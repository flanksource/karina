package flux

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
)

func Install(p *platform.Platform) error {
	if len(p.GitOps) == 0 {
		return nil
	}
	if err := p.ApplySpecs("", "helm-operator-crd.yaml"); err != nil {
		return err
	}
	for _, gitops := range p.GitOps {
		if gitops.Namespace != "" && gitops.Namespace != constants.KubeSystem && gitops.Namespace != constants.PlatformSystem {
			if err := p.CreateOrUpdateWorkloadNamespace(gitops.Namespace, nil, nil); err != nil {
				return fmt.Errorf("install: failed to create namespace: %s: %v", gitops.Namespace, err)
			}
		}

		if err := p.Apply(gitops.Namespace, NewFluxDeployment(&gitops)...); err != nil {
			return fmt.Errorf("install: failed to apply deployment: %v", err)
		}
	}
	return nil
}

// Create flux command arguments from CR
func defaults(cr *types.GitOps) {
	if cr.Namespace == "" {
		cr.Namespace = constants.KubeSystem
	}
	if cr.Name == "" {
		cr.Name = cr.Namespace
	}

	if cr.GitBranch == "" {
		cr.GitBranch = "master"
	}

	if cr.GitPath == "" {
		cr.GitPath = "./"
	}

	if cr.GitPollInterval == "" {
		cr.GitPollInterval = "60s"
	}

	if cr.SyncInterval == "" {
		cr.SyncInterval = "5m00s"
	}

	if cr.FluxVersion == "" {
		cr.FluxVersion = "1.20.0"
	}
	if cr.DisableScanning == nil {
		t := true
		cr.DisableScanning = &t
	}
}

func getArgs(cr *types.GitOps, argMap map[string]string) []string {
	var args []string
	for key, value := range cr.Args {
		argMap[key] = value
	}
	for key, value := range argMap {
		args = append(args, fmt.Sprintf("--%s=%s", key, value))
	}

	sort.Strings(args)
	return args
}

// NewFluxDeployment creates a new flux pod
func NewFluxDeployment(cr *types.GitOps) []runtime.Object {
	defaults(cr)
	memcacheName := fmt.Sprintf("flux-memcache-%s", cr.Name)
	secretName := fmt.Sprintf("flux-git-deploy-%s", cr.Name)
	sshConfig := fmt.Sprintf("flux-ssh-%s", cr.Name)
	saName := fmt.Sprintf("flux-" + cr.Name)
	argMap := map[string]string{
		"git-url":                cr.GitURL,
		"git-branch":             cr.GitBranch,
		"git-path":               cr.GitPath,
		"git-poll-interval":      cr.GitPollInterval,
		"sync-interval":          cr.SyncInterval,
		"k8s-secret-name":        secretName,
		"ssh-keygen-dir":         "/etc/fluxd/ssh",
		"memcached-hostname":     memcacheName,
		"manifest-generation":    "true",
		"registry-exclude-image": "*",
		// use ClusterIP rather than DNS SRV lookup
		"memcached-service": "",
	}

	spec := k8s.Builder{
		Namespace: cr.Namespace,
	}

	if *cr.DisableScanning {
		argMap["git-readonly"] = "true"
		argMap["registry-disable-scanning"] = "true"
	} else {
		// memcache is only deployed for scanning
		spec.Deployment(memcacheName, "docker.io/memcached:1.4.36-alpine").
			Args("-m 512", "-p 11211", "-I 5m").
			Expose(11211).
			Build()
	}

	repo := "docker.io/fluxcd/flux"
	if strings.Contains(cr.FluxVersion, "flanksource") {
		repo = "docker.io/flanksource/flux"
	}
	spec.Deployment("flux-"+cr.Name, fmt.Sprintf("%s:%s", repo, cr.FluxVersion)).
		Labels(map[string]string{
			"app": "flux",
		}).
		Args(getArgs(cr, argMap)...).
		ServiceAccount(saName).
		MountSecret(secretName, "/etc/fluxd/ssh", int32(0400)).
		MountConfigMap(sshConfig, "/root/.ssh").
		Expose(3030).
		Build()

	var sa *k8s.ServiceAccountBuilder
	if cr.Namespace == constants.KubeSystem {
		spec.ServiceAccount(saName).AddClusterRole("cluster-admin")
	} else {
		sa = spec.ServiceAccount(saName).AddRole("namespace-admin").AddRole("namespace-creator")
	}

	if cr.HelmOperatorVersion != "" {
		args := []string{"--enabled-helm-versions=v3"}
		if cr.Namespace != constants.KubeSystem {
			args = append(args, "--allow-namespace="+cr.Namespace)
		}
		spec.Deployment("helm-operator-"+cr.Name, fmt.Sprintf("docker.io/fluxcd/helm-operator:%s", cr.HelmOperatorVersion)).
			Labels(map[string]string{
				"app": "helm-operator",
			}).
			Args(args...).
			ServiceAccount(saName).
			MountSecret(secretName, "/etc/fluxd/ssh", int32(0400)).
			MountConfigMap(sshConfig, "/root/.ssh").
			Expose(3030).
			Build()
		if sa != nil {
			sa.AddClusterRole("helm-operator-admin")
		}
	}
	//TODO: else delete existing helm-operator deployment

	data, _ := base64.StdEncoding.DecodeString(cr.GitKey)
	spec.Secret(secretName, map[string][]byte{
		"identity": data,
	})
	spec.ConfigMap(sshConfig, map[string]string{
		"known_hosts": cr.KnownHosts,
		"config":      cr.SSHConfig,
	})

	return spec.Objects
}
