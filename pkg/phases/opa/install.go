package opa

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	if err := p.ApplyText(namespace, "opa-gatekeeper-monitoring.yaml.raw"); err != nil {
		return err
	}

	p.WaitForNamespace(namespace, 600*time.Second)

	if !p.DryRun {
		defaultConstraints, err := p.GetResourcesByDir("/gatekeeper", "manifests")
		if err != nil {
			return err
		}
		for name := range defaultConstraints {
			constraint, err := p.GetResourceByName("/gatekeeper/"+name, "manifests")
			if err != nil {
				return err
			}
			if err = p.ApplyText("", constraint); err != nil {
				return err
			}
		}
	}

	if p.Gatekeeper.Templates != "" && !p.DryRun {
		if err := deployTemplates(p, p.Gatekeeper.Templates); err != nil {
			return err
		}
	}

	if !p.DryRun {
		start := time.Now()

		k8sclient, err := p.GetClientset()
		if err != nil {
			return errors.Wrap(err, "Failed to create clientset")
		}

		templateClient, err := p.GetClientByKind("ConstraintTemplate")
		if err != nil {
			return err
		}

		for {
			templateList, err := templateClient.List(context.TODO(), metav1.ListOptions{})

			if err != nil {
				return err
			}

			if start.Add(60 * time.Second).Before(time.Now()) {
				return fmt.Errorf("timeout exceeded waiting for ConstraintTemplates")
			}

			ready := true
			for _, template := range templateList.Items {
				ctName := template.Object["metadata"].(map[string]interface{})["name"].(string)
				p.Debugf("Checking creation status of %s", ctName)
				if template.Object["status"] != nil && !template.Object["status"].(map[string]interface{})["created"].(bool) {
					ready = false
				}
			}

			_, resources, err := k8sclient.ServerGroupsAndResources()
			if err != nil {
				return errors.Wrap(err, "Failed to obtain server resources")
			}

			var constraintTemplates []metav1.APIResource
			for _, res := range resources {
				if strings.HasPrefix(res.GroupVersion, "constraints.gatekeeper.sh") {
					constraintTemplates = res.APIResources
					break
				}
			}
			p.Debugf("Checking creation of constraint resources")
			if len(constraintTemplates) < len(templateList.Items) {
				ready = false
			}

			if ready {
				break
			}

			time.Sleep(1 * time.Second)
		}
	}

	start := time.Now()
	clientset, _ := p.GetClientset()
	for {
		// Wait until the validating webhooks have their CA injected by cert-manager so that
		// subsequent namespace creation requests don't fail
		webhook, _ := clientset.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Get(context.TODO(), "gatekeeper-validating-webhook-configuration", metav1.GetOptions{})
		if webhook != nil && len(webhook.Webhooks) > 0 && len(webhook.Webhooks[0].ClientConfig.CABundle) > 100 {
			break
		} else {
			time.Sleep(3 * time.Second)
		}

		if start.Add(120 * time.Second).Before(time.Now()) {
			return fmt.Errorf("timeout waiting for ValidatingWebhook to get a CA injected")
		}
	}

	if p.Gatekeeper.Constraints != "" && !p.DryRun {
		if err := deployConstraints(p, p.Gatekeeper.Constraints); err != nil {
			return err
		}
	}

	return nil
}
