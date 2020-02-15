package test

import (
	"fmt"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/pkg/errors"

	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
)

type KubeConfig struct {
	Platform *platform.Platform
	IDToken  string
}

func GenerateKubeConfigOidc(p *platform.Platform, idToken string) ([]byte, error) {
	ca := p.GetIngressCA()
	data, err := k8s.CreateOIDCKubeConfig(p.Name, ca, "localhost", fmt.Sprintf("https://dex.%s", p.Domain))
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate config")
	}
	config, err := clientcmd.Load(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load config")
	}

	for k, v := range config.AuthInfos {
		if k == "sso" {
			v.AuthProvider.Config["id-token"] = idToken
		}
	}

	return clientcmd.Write(*config)
}
