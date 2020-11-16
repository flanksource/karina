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
	"fmt"
	"net/url"
	"os"
	"sync"

	"github.com/flanksource/commons/logger"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vapi/library"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vim25/soap"

	"github.com/flanksource/karina/pkg/types"
)

var sessionCache = map[string]Session{}
var sessionMU sync.Mutex

// Session is a vSphere session with a configured Finder.
type Session struct {
	logger.Logger
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

	logger.Debugf("Logging into vcenter: %s@%s", user, vcenter)
	soapURL, err := soap.ParseURL(vcenter)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing vSphere URL %q", vcenter)
	}
	if soapURL == nil {
		return nil, errors.Errorf("error parsing vSphere URL %q", vcenter)
	}

	soapURL.User = url.UserPassword(user, pass)

	// Temporarily setting the insecure flag True
	client, err := govmomi.NewClient(context.TODO(), soapURL, true)
	if err != nil {
		return nil, errors.Wrapf(err, "error setting up new vSphere SOAP client")
	}

	session := Session{Client: client, Logger: logger.StandardLogger()}

	session.UserAgent = "karina"

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
func (s Session) FindVM(nameOrID string) (*object.VirtualMachine, error) {
	tpl, err := s.findVMByUUID(nameOrID)
	if err != nil {
		return nil, fmt.Errorf("findVM: failed to find VM by UUID: %v", err)
	}
	if tpl != nil {
		return tpl, nil
	}
	return s.findVMByName(nameOrID)
}

func (s Session) FindTemplate(libraryName, nameOrID string) (*library.Item, error) {
	template, err := s.findTemplate(libraryName, nameOrID)
	if err != nil {
		return nil, fmt.Errorf("findTemplate: failed to find template %s in library %s: %v", nameOrID, libraryName, err)
	}

	return template, nil
}

func (s Session) findVMByUUID(templateID string) (*object.VirtualMachine, error) {
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

func (s Session) findVMByName(templateID string) (*object.VirtualMachine, error) {
	tpl, err := s.Finder.VirtualMachine(context.TODO(), templateID)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to find tempate by name %q", templateID)
	}
	return tpl, nil
}

func (s Session) findLibrary(libraryName string) (*library.Library, error) {
	restClient := rest.NewClient(s.Client.Client)
	user := url.UserPassword(os.Getenv("GOVC_USER"), os.Getenv("GOVC_PASS"))
	if err := restClient.Login(context.TODO(), user); err != nil {
		return nil, errors.Wrap(err, "failed to login")
	}
	manager := library.NewManager(restClient)

	return manager.GetLibraryByName(context.TODO(), libraryName)
}

func (s Session) findTemplate(libraryName, nameOrID string) (*library.Item, error) {
	restClient := rest.NewClient(s.Client.Client)
	user := url.UserPassword(os.Getenv("GOVC_USER"), os.Getenv("GOVC_PASS"))
	if err := restClient.Login(context.TODO(), user); err != nil {
		return nil, errors.Wrap(err, "failed to login")
	}
	manager := library.NewManager(restClient)

	lib, err := manager.GetLibraryByName(context.TODO(), libraryName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get library with name %s", libraryName)
	}

	items, err := manager.GetLibraryItems(context.TODO(), lib.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to list library items for library %s", lib.ID)
	}

	for _, i := range items {
		if i.ID == nameOrID || i.Name == nameOrID {
			return &i, nil
		}
	}

	return nil, fmt.Errorf("could not find any library item with name or id %s", nameOrID)
}

func isValidUUID(str string) bool {
	_, err := uuid.Parse(str)
	return err == nil
}

func LoadGovcEnvVars(vsphere types.Vsphere, vm *types.VM) {
	if vm.Datastore == "" {
		vm.Datastore = vsphere.Datastore
	}
	if len(vm.Network) == 0 {
		vm.Network = []string{vsphere.Network}
	}

	if vm.Folder == "" {
		vm.Folder = vsphere.Folder
	}
	if vm.Cluster == "" {
		vm.Cluster = vsphere.Cluster
	}
	if vm.ResourcePool == "" {
		vm.ResourcePool = vsphere.ResourcePool
	}
	if vm.Tags == nil {
		vm.Tags = make(map[string]string)
	}
}

func GetSessionFromEnv() (*Session, error) {
	return GetOrCreateCachedSession(
		os.Getenv("GOVC_DATACENTER"),
		os.Getenv("GOVC_USER"),
		os.Getenv("GOVC_PASS"),
		os.Getenv("GOVC_FQDN"))
}
