package quack

import (
	"github.com/flanksource/karina/pkg/platform"
)

var (
	EnabledLabels = map[string]string{
		EnabledLabel: "true",
	}
)

const EnabledLabel = "quack.pusher.com/enabled"
const Namespace = "quack"
const WebhookService = "quack"
const Certs = "quack-certs"

func Install(platform *platform.Platform) error {
	if platform.Quack != nil && platform.Quack.Disabled {
		if err := platform.DeleteMutatingWebhook(Namespace, WebhookService); err != nil {
			return err
		}
		return platform.DeleteSpecs(Namespace, "quack.yaml")
	}
	if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}

	ca, err := platform.CreateOrGetWebhookCertificate(Namespace, WebhookService)
	if err != nil {
		return err
	}

	if err := platform.ApplySpecs(Namespace, "quack.yaml"); err != nil {
		return err
	}

	webhooks, err := platform.CreateWebhookBuilder(Namespace, WebhookService, ca)
	if err != nil {
		return err
	}

	webhooks = webhooks.NewHook("quack.pusher.com", "/apis/quack.pusher.com/v1alpha1/admissionreviews").
		WithNamespaceLabel(EnabledLabel, "true").
		MatchKinds("ingresses", "services", "canaries").
		Add()

	return platform.Apply(Namespace, webhooks.BuildMutating())
}
