package opa

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/moshloop/platform-cli/pkg/platform"
)

const (
	Namespace = "opa"
)

func Install(platform *platform.Platform) error {
	if platform.OPA == nil || platform.OPA.Disabled {
		if err := platform.DeleteSpecs("", "opa.yaml"); err != nil {
			platform.Warnf("failed to delete specs: %v", err)
		}
		return nil
	}
	if platform.OPA.KubeMgmtVersion == "" {
		platform.OPA.KubeMgmtVersion = "0.8"
	}

	if platform.OPA.LogLevel == "" {
		platform.OPA.LogLevel = "error"
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

	if err := platform.ApplySpecs(Namespace, "opa.yaml"); err != nil {
		return err
	}
	if platform.OPA.Policies != "" {
		return deploy(platform, platform.OPA.Policies)
	}
	return nil
}
