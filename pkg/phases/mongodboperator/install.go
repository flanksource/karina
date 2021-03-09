package mongodboperator

import (
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/platform"
)

func Deploy(platform *platform.Platform) error {
	if platform.MongodbOperator.IsDisabled() {
		return platform.DeleteSpecs("", "mongodb-operator.yaml", "template/mongo-db.yaml.raw")
	}

	if err := platform.CreateOrUpdateNamespace(constants.PlatformSystem, nil, nil); err != nil {
		return err
	}

	// Trim the first character e.g. v1.6.0 -> 1.6.0
	platform.MongodbOperator.Version = platform.MongodbOperator.Version[1:]

	return platform.ApplySpecs(constants.PlatformSystem, "mongodb-operator.yaml", "template/mongo-db.yaml.raw")
}
