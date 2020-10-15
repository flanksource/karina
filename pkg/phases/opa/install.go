package opa

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flanksource/karina/pkg/platform"
	"github.com/pkg/errors"
)

const (
	Namespace           = "opa"
	GatekeeperNamespace = "gatekeeper-system"
)

func Install(platform *platform.Platform) error {
	if platform.OPA != nil && !platform.OPA.Disabled && !platform.Gatekeeper.IsDisabled() {
		platform.Fatalf("both opa kubemgmt and gatekeeper are enabled. Please disable one of them to continue")
	}

	if err := InstallGatekeeper(platform); err != nil {
		return errors.Wrap(err, "failed to install gatekeeper")
	}

	if err := InstallKubemgmt(platform); err != nil {
		return errors.Wrap(err, "failed to install opa kubemgmt")
	}

	return nil
}

func InstallKubemgmt(platform *platform.Platform) error {
	if platform.OPA == nil || platform.OPA.Disabled {
		if err := platform.DeleteSpecs("", "opa-kubemgmt.yaml"); err != nil {
			platform.Warnf("failed to delete specs: %v", err)
		}
		return nil
	}
	if platform.OPA.KubeMgmtVersion == "" {
		platform.OPA.KubeMgmtVersion = "0.11"
	}

	if platform.OPA.LogLevel == "" {
		platform.OPA.LogLevel = "info"
	}

	if err := platform.CreateOrUpdateNamespace(Namespace, map[string]string{
		"app": "opa",
	}, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	if err := platform.Apply(Namespace, &v1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "opa-server",
			Namespace: Namespace,
			Annotations: map[string]string{
				"cert-manager.io/allow-direct-injection": "true",
			},
		},
	}); err != nil {
		return fmt.Errorf("install: failed to create secret opa-server: %v", err)
	}

	if err := platform.ApplySpecs(Namespace, "opa-kubemgmt.yaml"); err != nil {
		return err
	}
	if platform.OPA.Policies != "" {
		return deploy(platform, platform.OPA.Policies)
	}
	return nil
}

func InstallGatekeeper(p *platform.Platform) error {
	namespace := GatekeeperNamespace

	if p.Gatekeeper.IsDisabled() {
		if err := p.DeleteSpecs("", "opa-gatekeeper.yaml"); err != nil {
			p.Warnf("failed to delete specs: %v", err)
		}
		return nil
	}

	if err := p.CreateOrUpdateNamespace(namespace, map[string]string{
		"admission.gatekeeper.sh/ignore": "no-self-managing",
		"control-plane":                  "controller-manager",
		"gatekeeper.sh/system":           "yes",
	}, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	if err := p.ApplySpecs(namespace, "opa-gatekeeper.yaml"); err != nil {
		return err
	}

	if p.Gatekeeper.Templates != "" {
		if err := deployTemplates(p, p.Gatekeeper.Templates); err != nil {
			return err
		}
	}

	if p.Gatekeeper.Constraints != "" {
		if err := deployConstraints(p, p.Gatekeeper.Constraints); err != nil {
			return err
		}
	}

	return nil
}
