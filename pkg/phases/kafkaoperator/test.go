package kafkaoperator

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/utils"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.KafkaOperator.IsDisabled() {
		return
	}

	client, _ := p.GetClientset()
	kommons.TestNamespace(client, Namespace, test)
	if p.E2E {
		TestE2E(p, test)
	}
}

func TestE2E(p *platform.Platform, test *console.TestResults) {
	testName := "kafka-operator-e2e"
	clusterName := fmt.Sprintf("e2e-test-%s", utils.RandomString(6))

	kafkaManifest := fmt.Sprintf(`
apiVersion: kafka.strimzi.io/v1beta2
kind: Kafka
metadata:
  name: %s
spec:
  kafka:
    version: 2.7.0
    replicas: 1
    listeners:
      - name: plain
        port: 9092
        type: internal
        tls: false
      - name: tls
        port: 9093
        type: internal
        tls: true
    config:
      offsets.topic.replication.factor: 1
      transaction.state.log.replication.factor: 1
      transaction.state.log.min.isr: 1
      log.message.format.version: "2.7"
      inter.broker.protocol.version: "2.7"
    storage:
      type: ephemeral
  zookeeper:
    replicas: 3
    storage:
      type: ephemeral
  entityOperator:
    topicOperator: {}
    userOperator: {}
`, clusterName)

	kafkaObj := &unstructured.Unstructured{}
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, _, err := dec.Decode([]byte(kafkaManifest), nil, kafkaObj)
	if err != nil {
		test.Failf(testName, "Failed to parse Kafka CR: %v", err)
		return
	}

	defer func() {
		err := p.DeleteUnstructured(Namespace, kafkaObj)
		if err != nil {
			test.Warnf("Failed to delete Kafka Cluster %s: %v", clusterName, err)
		}
	}()
	err = p.ApplyUnstructured(Namespace, kafkaObj)
	if err != nil {
		test.Failf(testName, "Failed to create Kafka Cluster %s: %v", clusterName, err)
		return
	}
	_, err = p.WaitForResource("Kafka", Namespace, clusterName, 5*time.Minute)
	if err != nil {
		test.Failf(testName, "Kafka Cluster didn't come up healthy within allowed time: %v", err)
		return
	}
	test.Passf(testName, "E2E Kafka Cluster created successfully and came up healthy.")
}
