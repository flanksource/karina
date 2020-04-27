package registrycreds

import (
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/pkg/errors"
)

func Install(p *platform.Platform) error {
	if p.RegistryCredentials == nil || p.RegistryCredentials.Disabled {
		return nil
	}
	if p.RegistryCredentials.Namespace == "" {
		p.RegistryCredentials.Namespace = "platform-system"
	}
	namespace := p.RegistryCredentials.Namespace

	if err := p.CreateOrUpdateNamespace(namespace, nil, nil); err != nil {
		return errors.Wrapf(err, "install: failed to create/update namespace: %s", namespace)
	}

	if err := p.ApplySpecs(namespace, "registry-creds-secrets.yaml"); err != nil {
		return errors.Wrap(err, "install: failed to apply registry-creds-secrets.yaml")
	}

	if err := p.ApplySpecs(namespace, "registry-creds.yaml"); err != nil {
		return errors.Wrap(err, "install: failed to apply registry-creds.yaml")
	}

	return nil
}
