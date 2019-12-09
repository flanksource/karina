/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vmware

import (
	"context"
	"net/url"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/moshloop/platform-cli/pkg/types"
	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/soap"
)

var sessionCache = map[string]Session{}
var sessionMU sync.Mutex

// Session is a vSphere session with a configured Finder.
type Session struct {
	*govmomi.Client
	Finder     *find.Finder
	datacenter *object.Datacenter
}

func GetOrCreateCachedSession(datacenter, user, pass, vcenter string) (*Session, error) {
	sessionMU.Lock()
	defer sessionMU.Unlock()

	sessionKey := vcenter + user + datacenter

	if session, ok := sessionCache[sessionKey]; ok && session.IsVC() {
		if ok, _ := session.SessionManager.SessionIsActive(context.TODO()); ok {
			return &session, nil
		}
	}

	log.Infof("Logging into vcenter: %s@%s", user, vcenter)
	soapURL, err := soap.ParseURL(vcenter)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing vSphere URL %q", vcenter)
	}
	if soapURL == nil {
		return nil, errors.Errorf("error parsing vSphere URL %q", vcenter)
	}

	soapURL.User = url.UserPassword(user, pass)

	// Temporarily setting the insecure flag True
	// TODO(ssurana): handle the certs better
	client, err := govmomi.NewClient(context.TODO(), soapURL, true)
	if err != nil {
		return nil, errors.Wrapf(err, "error setting up new vSphere SOAP client")
	}

	session := Session{Client: client}

	// TODO(frapposelli): replace `dev` with version string
	session.UserAgent = "platform-cli"

	// Assign the finder to the session.
	session.Finder = find.NewFinder(session.Client.Client, false)

	// Assign the datacenter if one was specified.
	dc, err := session.Finder.DatacenterOrDefault(context.TODO(), datacenter)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to find datacenter %q", datacenter)
	}
	session.datacenter = dc
	session.Finder.SetDatacenter(dc)

	// Cache the session.
	sessionCache[sessionKey] = session
	return &session, nil
}

// FindByInstanceUUID finds an object by its instance UUID.
func (s *Session) FindByInstanceUUID(ctx context.Context, uuid string) (object.Reference, error) {
	if s.Client == nil {
		return nil, errors.New("vSphere client is not initialized")
	}
	si := object.NewSearchIndex(s.Client.Client)
	findFlag := true
	ref, err := si.FindByUuid(ctx, s.datacenter, uuid, true, &findFlag)
	if err != nil {
		return nil, errors.Wrapf(err, "error finding object by instance uuid %q", uuid)
	}
	return ref, nil
}

// FindByUUID finds an object by its UUID.
func (s *Session) FindByUUID(ctx context.Context, uuid string) (object.Reference, error) {
	if s.Client == nil {
		return nil, errors.New("vSphere client is not initialized")
	}
	si := object.NewSearchIndex(s.Client.Client)
	findFlag := false
	ref, err := si.FindByUuid(ctx, s.datacenter, uuid, true, &findFlag)
	if err != nil {
		return nil, errors.Wrapf(err, "error finding object by uuid %q", uuid)
	}
	return ref, nil
}

// FindVM finds a template based either on a UUID or name.
func (s Session) FindVM(nameOrId string) (*object.VirtualMachine, error) {
	tpl, err := s.findVmByUuid(nameOrId)
	if err != nil {
		return nil, err
	}
	if tpl != nil {
		return tpl, nil
	}
	return s.findVmByName(nameOrId)
}

func (s Session) findVmByUuid(templateID string) (*object.VirtualMachine, error) {
	if !isValidUUID(templateID) {
		return nil, nil
	}
	ref, err := s.FindByInstanceUUID(context.TODO(), templateID)
	if err != nil {
		return nil, errors.Wrap(err, "error querying template by instance UUID")
	}
	if ref != nil {
		return object.NewVirtualMachine(s.Client.Client, ref.Reference()), nil
	}
	return nil, nil
}

func (s Session) findVmByName(templateID string) (*object.VirtualMachine, error) {

	tpl, err := s.Finder.VirtualMachine(context.TODO(), templateID)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to find tempate by name %q", templateID)
	}
	return tpl, nil
}

func isValidUUID(str string) bool {
	_, err := uuid.Parse(str)
	return err == nil
}

func LoadGovcEnvVars(vm *types.VM) {
	if vm.Datastore == "" {
		vm.Datastore = os.Getenv("GOVC_DATASTORE")
	}
	if vm.Network == "" {
		vm.Network = os.Getenv("GOVC_NETWORK")
	}

	if vm.Folder == "" {
		vm.Folder = os.Getenv("GOVC_FOLDER")
	}
	if vm.Cluster == "" {
		vm.Cluster = os.Getenv("GOVC_CLUSTER")
	}
	if vm.ResourcePool == "" {
		vm.ResourcePool = os.Getenv("GOVC_RESOURCE_POOL")
	}
}

func GetSessionFromEnv() (*Session, error) {
	return GetOrCreateCachedSession(
		os.Getenv("GOVC_DATACENTER"),
		os.Getenv("GOVC_USER"),
		os.Getenv("GOVC_PASS"),
		os.Getenv("GOVC_FQDN"))
}
