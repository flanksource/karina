package kafkaoperator

import (
	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace = "kafka-operator"
)

func Deploy(platform *platform.Platform) error {
	if platform.KafkaOperator.IsDisabled() {
		return platform.DeleteSpecs(Namespace, "kafka-operator.yaml")
	}

	if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}

	// Trim the first character e.g. v1.6.0 -> 1.6.0
	platform.KafkaOperator.Version = platform.KafkaOperator.Version[1:]

	return platform.ApplySpecs(Namespace, "kafka-operator.yaml")
}
