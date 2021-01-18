package base

import (
	"context"
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func Test(platform *platform.Platform, test *console.TestResults) {
	client, err := platform.GetClientset()
	if err != nil {
		test.Errorf("Base tests failed to get clientset: %v", err)
		return
	}
	if client == nil {
		test.Errorf("Base tests failed to get clientset: nil clientset ")
		return
	}

	kommons.TestNamespace(client, "kube-system", test)
	kommons.TestNamespace(client, "local-path-storage", test)
	kommons.TestNamespace(client, "cert-manager", test)

	if platform.Nginx == nil || !platform.Nginx.Disabled {
		platform.WaitForNamespace("ingress-nginx", 180*time.Second)
		kommons.TestNamespace(client, "ingress-nginx", test)
	}

	TestKommonsTemplate(platform, test)
}

func TestKommonsTemplate(platform *platform.Platform, test *console.TestResults) {
	clientset, err := platform.GetClientset()
	if err != nil {
		test.Failf("base", "failed to get clientset: %v", err)
		return
	}

	cm, err := clientset.CoreV1().ConfigMaps("kube-public").Get(context.Background(), "cluster-info", metav1.GetOptions{})
	if err != nil {
		test.Failf("base", "failed to get kube-public/cluster-info config map")
		return
	}

	jws, found := cm.Data["jws-kubeconfig-abcdef"]
	if !found {
		test.Failf("base", "failed to get kube-public/cluster-info field jws-kubeconfig-abcdef")
		return
	}

	template := `
apiVersion: v1
kind: Secret
metadata:
  name: test1
  namespace: test2
  annotations:
    "test1.com/test2": "{{ kget "configmap/kube-public/cluster-info" "data.jws-kubeconfig-abcdef" }}"
`
	templateResult, err := platform.TemplateText(template)
	if err != nil {
		test.Failf("base", "failed to template secret: %v", err)
		return
	}

	secret := &v1.Secret{}
	if err := yaml.Unmarshal([]byte(templateResult), secret); err != nil {
		test.Failf("base", "failed to unmarshal into secret: %v", err)
		return
	}

	annotation, found := secret.Annotations["test1.com/test2"]
	if !found {
		test.Failf("base", "failed to find annotation in secret")
		return
	}
	if annotation != jws {
		test.Failf("base", "expected templated value to equal %s, got %s", jws, annotation)
	} else {
		test.Passf("base", "kget successfully applied for secret")
	}
}
