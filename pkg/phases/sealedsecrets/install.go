package sealedsecrets

import (
	"fmt"
	"sort"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	"github.com/flanksource/commons/certs"
	"github.com/flanksource/commons/files"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

const (
	Namespace = "sealed-secrets"
	// SealedSecretsKeyLabel is that label used to locate active key pairs used to decrypt sealed secrets.
	SealedSecretsKeyLabel = "sealedsecrets.bitnami.com/sealed-secrets-key"
	// Prefix is used to prefix tls secret key
	SecretPrefix = "sealed-secrets-key"
)

var (
	keySelector = fields.OneTermEqualSelector(SealedSecretsKeyLabel, "active")
)

func Install(platform *platform.Platform) error {
	if !platform.SealedSecrets.Enabled {
		return nil
	}

	if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	if platform.SealedSecrets.Cert != "" && platform.SealedSecrets.PrivateKey != "" {
		certBytes := []byte(files.SafeRead(platform.SealedSecrets.Cert))
		keyBytes := []byte(files.SafeRead(platform.SealedSecrets.PrivateKey))
		cert, err := certs.DecryptCertificate(certBytes, keyBytes, []byte(platform.SealedSecrets.Password))
		if err != nil {
			return errors.Wrap(err, "failed to decrypt certificate")
		}

		client, err := platform.GetClientset()
		if err != nil {
			return errors.Wrap(err, "failed to get k8s client")
		}

		secretList, err := client.CoreV1().Secrets(Namespace).List(metav1.ListOptions{
			LabelSelector: keySelector.String(),
		})

		if err != nil {
			return errors.Wrap(err, "failed to list secrets")
		}

		items := secretList.Items
		sort.Sort(ssv1alpha1.ByCreationTimestamp(items))
		secrets := client.CoreV1().Secrets(Namespace)

		if len(items) == 0 {
			secret := &v1.Secret{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: SecretPrefix,
					Labels: map[string]string{
						SealedSecretsKeyLabel: "active",
					},
				},
				Data: cert.AsTLSSecret(),
				Type: "kubernetes.io/tls",
			}

			log.Infof("Creating %s/secret/%s", Namespace, SecretPrefix)

			if _, err := secrets.Create(secret); err != nil {
				return errors.Wrap(err, "failed to create new secret")
			}
		} else {
			secret := items[len(items)-1]
			secret.Data = cert.AsTLSSecret()

			log.Infof("Updating %s/secret/%s", Namespace, secret.Name)

			if _, err := secrets.Update(&secret); err != nil {
				return errors.Wrapf(err, "failed to update secret %s", secret.Name)
			}
		}
	}

	return platform.ApplySpecs(Namespace, "sealed-secrets.yml")
}
