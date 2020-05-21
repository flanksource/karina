package quack

import (
	"fmt"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/utils"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Test(platform *platform.Platform, test *console.TestResults) {
	if platform.Quack != nil && platform.Quack.Disabled {
		return
	}

	client, _ := platform.GetClientset()
	k8s.TestNamespace(client, Namespace, test)

	if !platform.E2E {
		return
	}

	namespace := fmt.Sprintf("quack-test-%s", utils.RandomKey(6))

	if err := platform.CreateOrUpdateWorkloadNamespace(namespace, EnabledLabels, nil); err != nil {
		test.Failf("quack", "Failed to create test namespace %s: %v", namespace, err)
		return
	}

	defer func() {
		client.CoreV1().Namespaces().Delete(namespace, nil) // nolint: errcheck
	}()

	configMapName := "test-configmap"
	data := map[string]string{
		"prometheus":   "prometheus.{{ .domain }}",
		"grafana":      "grafana.{{ .domain }}",
		"cluster.name": "Cluster {{ .name }}",
	}

	if err := platform.CreateOrUpdateConfigMap(configMapName, namespace, data); err != nil {
		test.Failf("quack", "failed to create configmap %s in namespace %s: %v", configMapName, namespace, err)
		return
	}

	rcm := platform.GetConfigMap(namespace, configMapName)
	if rcm == nil {
		test.Failf("quack", "failed to retrieve configmap %s in namespace %s", configMapName, namespace)
		return
	}

	cm := *rcm

	if cm["prometheus"] != fmt.Sprintf("prometheus.%s", platform.Domain) {
		test.Failf("quack", "expected prometheus config value to equal %s got %s", fmt.Sprintf("prometheus.%s", platform.Domain), cm["prometheus"])
		return
	}

	if cm["grafana"] != fmt.Sprintf("grafana.%s", platform.Domain) {
		test.Failf("quack", "expected grafana config value to equal %s got %s", fmt.Sprintf("grafana.%s", platform.Domain), cm["grafana"])
		return
	}

	if cm["cluster.name"] != fmt.Sprintf("Cluster %s", platform.Name) {
		test.Failf("quack", "expected cluster name config value to equal %s got %s", fmt.Sprintf("Cluster %s", platform.Name), cm["cluster.name"])
		return
	}

	test.Passf("quack", "configmap %s in namespace %s was successfully updated by quack", configMapName, namespace)
}
