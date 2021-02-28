package keptn

import (
	"github.com/flanksource/karina/pkg/platform"
)

const (
	Namespace = "keptn"
)

func Deploy(platform *platform.Platform) error {
	if platform.Keptn.IsDisabled() {
		// TODO: Stop deleting template/mongo-db.yaml.raw once MongoDB Operator is implemented. Related issue: https://github.com/flanksource/karina/issues/658
		return platform.DeleteSpecs("", "template/mongo-db.yaml.raw", "keptn.yaml")
	}

	if err := platform.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return err
	}

	// TODO: Stop applying template/mongo-db.yaml.raw as part of keptn once MongoDB Operator is implemented. Related issue: https://github.com/flanksource/karina/issues/658
	return platform.ApplySpecs(Namespace, "template/mongo-db.yaml.raw", "keptn.yaml")
}
