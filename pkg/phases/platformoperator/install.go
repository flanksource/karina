package platformoperator

import (
	"github.com/moshloop/platform-cli/pkg/constants"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/types"
)

const Namespace = constants.PlatformSystem

func Install(platform *platform.Platform) error {
	if platform.PlatformOperator == nil || platform.PlatformOperator.Disabled {
		platform.PlatformOperator = &types.PlatformOperator{}
		if err := platform.DeleteSpecs("", "platform-operator.yaml"); err != nil {
			platform.Warnf("failed to delete specs: %v", err)
		}
		return nil
	}

	labels := map[string]string{
		"control-plane":            "controller-manager",
		"quack.pusher.com/enabled": "true",
	}
	if err := platform.CreateOrUpdateNamespace(constants.PlatformSystem, labels, nil); err != nil {
		return err
	}

	var secrets = make(map[string][]byte)

	secrets["AWS_ACCESS_KEY_ID"] = []byte(platform.S3.AccessKey)
	secrets["AWS_SECRET_ACCESS_KEY"] = []byte(platform.S3.SecretKey)

	if platform.Ldap != nil {
		secrets["LDAP_USERNAME"] = []byte(platform.Ldap.Username)
		secrets["LDAP_PASSWORD"] = []byte(platform.Ldap.Password)
	}

	if err := platform.CreateOrUpdateSecret("secrets", constants.PlatformSystem, secrets); err != nil {
		return err
	}

	platform.Infof("Installing platform operator")
	if platform.PlatformOperator == nil {
		platform.PlatformOperator = &types.PlatformOperator{}
	}
	if platform.PlatformOperator.WhitelistedPodAnnotations == nil {
		platform.PlatformOperator.WhitelistedPodAnnotations = []string{}
	}
	if platform.PlatformOperator.Version == "" {
		platform.PlatformOperator.Version = "0.3"
	}

	return platform.ApplySpecs("", "platform-operator.yaml")
}
