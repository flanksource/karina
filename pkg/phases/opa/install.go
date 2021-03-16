package opa

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
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

	if p.DryRun && !p.Gatekeeper.IsDisabled() {
		_ = p.ApplySpecs(Namespace, "opa-gatekeeper.yaml", "opa-gatekeeper-monitoring.yaml.raw")
	} else if p.DryRun {
		return nil
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
