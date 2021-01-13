package quack

import (
	"time"

	"github.com/flanksource/karina/pkg/platform"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	EnabledLabels = map[string]string{
		"quack.pusher.com/enabled": "true",
	}
)

const Namespace = "quack"
const Certs = "quack-certs"

func Install(platform *platform.Platform) error {
	if platform.Quack != nil && platform.Quack.Disabled {
		return platform.DeleteSpecs(v1.NamespaceAll, "quack.yaml")
	}
	if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return errors.Wrap(err, "failed to create/update namespace quack")
	}

	if !platform.HasSecret(Namespace, Certs) {
		secret := platform.NewSelfSigned("quack.quack.svc").AsTLSSecret()
		secret["ca.crt"] = secret["tls.crt"]
		if err := platform.Apply(Namespace, &v1.Secret{
			TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      Certs,
				Namespace: Namespace,
				Annotations: map[string]string{
					"cert-manager.io/allow-direct-injection": "true",
				},
			},
			Data: secret,
		}); err != nil {
			return err
		}
	}
	// quack gets deployed across both quack and kube-system namespaces
	if err := platform.ApplySpecs(v1.NamespaceAll, "quack.yaml"); err != nil {
		return err
	}

	if !platform.ApplyDryRun {
		platform.WaitForNamespace(Namespace, 60*time.Second)
	}
	return nil
}
