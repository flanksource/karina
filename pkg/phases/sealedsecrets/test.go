package sealedsecrets

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/utils"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.SealedSecrets == nil || p.SealedSecrets.Disabled {
		test.Skipf("sealed-secrets", "sealed-secrets not installed or disabled")
		return
	}
	client, _ := p.GetClientset()
	k8s.TestNamespace(client, "sealed-secrets", test)
	if !p.E2E {
		return
	}
	secretName := "test-sealed-secrets"
	namespace := fmt.Sprintf("sealed-secrets-test-%s", utils.RandomKey(6))

	secretFile, err := ioutil.TempFile("", "secret.yaml")
	if err != nil {
		test.Failf("sealed-secrets", "Failed to create temporary file for secret.yaml %v", err)
		return
	}
	sealedSecretFile, err := ioutil.TempFile("", "sealed-secret.yaml")
	if err != nil {
		test.Failf("sealed-secrets", "Failed to create temporary file for sealed-secret.yaml %v", err)
		return
	}
	certFile, err := ioutil.TempFile("", "sealed-secret.crt")
	if err != nil {
		test.Failf("sealed-secrets", "Failed to create temporary file for sealed-secret.crt %v", err)
		return
	}

	secret := apiv1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"foo":   []byte("bar"),
			"other": []byte("true"),
		},
	}

	secretJSON, err := json.Marshal(&secret)
	if err != nil {
		test.Failf("sealed-secrets", "Failed to encode secret to json %v", err)
		return
	}

	if err := ioutil.WriteFile(secretFile.Name(), secretJSON, 0644); err != nil {
		test.Failf("sealed-secrets", "Failed to write secret to file %v", err)
		return
	}

	certPem, err := ioutil.ReadFile(p.SealedSecrets.Certificate.Cert)
	if err != nil {
		test.Failf("sealed-secrets", "Failed to read certificate file %s", p.SealedSecrets.Certificate.Cert)
		return
	}

	if err := ioutil.WriteFile(certFile.Name(), certPem, 0644); err != nil {
		test.Failf("sealed-secrets", "Failed to write certificate file %v", err)
		return
	}

	kubeseal := p.GetBinary("kubeseal")

	if err := kubeseal("--cert %s < %s > %s", certFile.Name(), secretFile.Name(), sealedSecretFile.Name()); err != nil {
		test.Failf("sealed-secrets", "Failed to run kubeseal %v", err)
		return
	}

	sealedSecret, err := ioutil.ReadFile(sealedSecretFile.Name())
	if err != nil {
		test.Failf("sealed-secrets", "Failed to read sealed secret file %v", err)
		return
	}

	if err := p.CreateOrUpdateWorkloadNamespace(namespace, nil, nil); err != nil {
		test.Failf("sealed-secrets", "Failed to create test namespace %s: %v", namespace, err)
		return
	}

	defer func() {
		client.CoreV1().Namespaces().Delete(namespace, nil) // nolint: errcheck
	}()

	if err := p.ApplyText(namespace, string(sealedSecret)); err != nil {
		test.Failf("sealed-secrets", "Failed to create sealed secret in namespace %s: %v", namespace, err)
		return
	}

	iterations := 30
	interval := 1 * time.Second
	var k8sSecret *apiv1.Secret

	for i := 0; i < iterations; i++ {
		ks, err := client.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
		if err != nil {
			time.Sleep(interval)
		} else {
			k8sSecret = ks
		}
	}

	if k8sSecret == nil {
		test.Failf("sealed-secrets", "Failed to get secret %s in namespace %s: %v", secretName, namespace, err)
		return
	}

	if string(k8sSecret.Data["foo"]) != "bar" {
		test.Failf("sealed-secrets", "Expected data value for key 'foo' to equal 'bar' got %s", string(k8sSecret.Data["foo"]))
	} else if string(k8sSecret.Data["other"]) != "true" {
		test.Failf("sealed-secrets", "Expected data value for key 'other' to equal 'true' got %s", string(k8sSecret.Data["foo"]))
	} else {
		test.Passf("sealed-secrets", "Secret %s created in namespace %s", secretName, namespace)
	}
}
