package sealedsecrets

import (
	"fmt"
	"sort"

	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/pkg/errors"
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
	if platform.SealedSecrets == nil || platform.SealedSecrets.Disabled {
		if err := platform.DeleteSpecs(Namespace, "sealed-secrets.yaml"); err != nil {
			platform.Warnf("failed to delete specs: %v", err)
		}
		return nil
	}

	if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("install: failed to create/update namespace: %v", err)
	}

	if platform.SealedSecrets.Certificate != nil {
		ca, err := platform.ReadCA(platform.SealedSecrets.Certificate)
		if err != nil {
			return errors.Wrap(err, "failed to read platform ca")
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
		sort.Sort(ByCreationTimestamp(items))
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
				Data: ca.AsTLSSecret(),
				Type: "kubernetes.io/tls",
			}

			platform.Infof("Creating %s/secret/%s", Namespace, SecretPrefix)

			if _, err := secrets.Create(secret); err != nil {
				return errors.Wrap(err, "failed to create new secret")
			}
		} else {
			secret := items[len(items)-1]
			secret.Data = ca.AsTLSSecret()

			platform.Infof("Updating %s/secret/%s", Namespace, secret.Name)

			if _, err := secrets.Update(&secret); err != nil {
				return errors.Wrapf(err, "failed to update secret %s", secret.Name)
			}
		}
	}

	return platform.ApplySpecs(Namespace, "sealed-secrets.yaml")
}

// ByCreationTimestamp is used to sort a list of secrets
type ByCreationTimestamp []v1.Secret

func (s ByCreationTimestamp) Len() int {
	return len(s)
}

func (s ByCreationTimestamp) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByCreationTimestamp) Less(i, j int) bool {
	return s[i].GetCreationTimestamp().Unix() < s[j].GetCreationTimestamp().Unix()
}
