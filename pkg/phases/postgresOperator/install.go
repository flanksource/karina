package postgresOperator

import (
	"github.com/moshloop/platform-cli/pkg/platform"
)

const (
	Namespace = "postgres-operator"
)

func Deploy(platform *platform.Platform) error {

	if err := platform.CreateOrUpdateNamespace("postgres-operator", nil, nil); err != nil {
		return err
	}

	if err := platform.ApplySpecs(Namespace, "postgres-operator.crd.yml"); err != nil {
		return err
	}

	if err := platform.ApplySpecs(Namespace, "postgres-operator-config.yml"); err != nil {
		return err
	}
	return platform.ApplySpecs(Namespace, "postgres-operator.yml")
}
