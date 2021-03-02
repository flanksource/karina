package opa

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Namespace      = "gatekeeper-system"
	WebhookService = "gatekeeper-webhook"
)

func Install(p *platform.Platform) error {
	if p.Gatekeeper.IsDisabled() {
		if err := p.DeleteValidatingWebhook(Namespace, WebhookService); err != nil {
			return err
		}
		return p.DeleteSpecs(Namespace, "opa-gatekeeper.yaml")
	}

	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}

	ca, err := p.CreateOrGetWebhookCertificate(Namespace, WebhookService)
	if err != nil {
		return err
	}

	if err := p.ApplySpecs(Namespace, "opa-gatekeeper.yaml", "opa-gatekeeper-monitoring.yaml.raw"); err != nil {
		return err
	}

	if p.DryRun {
		return nil
	}

	webhooks, err := p.CreateWebhookBuilder(Namespace, WebhookService, ca)
	if err != nil {
		return err
	}

	webhooks = webhooks.NewHook("validation.gatekeeper.sh", "/v1/admit").
		WithoutNamespaceLabel(constants.ManagedBy, constants.Karina).
		MatchAny().
		Add()

	webhooks = webhooks.NewHook("check-ignore-label.gatekeeper.sh", "/v1/admitlabel").
		WithNamespaceLabel(constants.ManagedBy, constants.Karina).
		MatchKinds("namespaces").
		Fail().
		Add()

	if err := p.Apply(Namespace, webhooks.Build()); err != nil {
		return err
	}

	if err := deployDefaultTemplates(p); err != nil {
		return err
	}

	if err := deployTemplates(p, p.Gatekeeper.Templates); err != nil {
		return err
	}

	if err := waitForTemplates(p); err != nil {
		return err
	}

	if err := deployConstraints(p, p.Gatekeeper.Constraints); err != nil {
		return err
	}

	return nil
}

func deployDefaultTemplates(p *platform.Platform) error {
	defaultConstraints, err := p.GetResourcesByDir("gatekeeper", "manifests")
	if err != nil {
		return err
	}
	for name := range defaultConstraints {
		constraint, err := p.GetResourceByName("gatekeeper/"+name, "manifests")
		if err != nil {
			return err
		}
		if err = p.ApplyText("", constraint); err != nil {
			return err
		}
	}

	return nil
}

func waitForTemplates(p *platform.Platform) error {
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
	return nil
}
