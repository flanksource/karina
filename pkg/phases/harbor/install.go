package harbor

import (
	"fmt"
	"os"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/deps"
	"github.com/flanksource/commons/files"
	"github.com/flanksource/commons/text"
	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/phases/pgo"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Deploy(p *platform.Platform) error {
	if p.Harbor == nil || p.Harbor.Disabled {
		log.Infof("Skipping deployment of harbor, it is disabled")
		return nil
	} else {
		log.Infof("Deploying harbor %s", p.Harbor.Version)
	}
	defaults(p)
	if p.Harbor.DB == nil {
		db, err := pgo.GetOrCreateDB(p, dbCluster, p.Harbor.Replicas)
		if err != nil {
			return fmt.Errorf("deploy: failed to create db cluster: %v", err)
		}
		if err := pgo.WaitForDB(p, dbCluster, 120); err != nil {
			return fmt.Errorf("deploy: failed to wait for db, %v", err)
		}

		if err := pgo.CreateDatabase(p, dbCluster, dbNames...); err != nil {
			return fmt.Errorf("deploy: failed to create db: %v", err)
		}
		p.Harbor.DB = db
	}

	if err := files.Getter(fmt.Sprintf("github.com/goharbor/harbor-helm?ref=%s", p.Harbor.ChartVersion), "build/harbor"); err != nil {
		return fmt.Errorf("deploy: failed to download Harbor: %v", err)
	}

	values, err := p.Template("harbor.yml", "manifests")
	if err != nil {
		return fmt.Errorf("deploy: failed to template Harbor manifests: %v", err)
	}
	log.Tracef("Config: \n%s\n", console.StripSecrets(values))
	kubeconfig, err := p.GetKubeConfig()
	if err != nil {
		return fmt.Errorf("deploy: failed to get kubeconfig: %v", err)
	}
	helm := deps.BinaryWithEnv("helm", p.Versions["helm"], ".bin", map[string]string{
		"KUBECONFIG": kubeconfig,
		"HOME":       os.ExpandEnv("$HOME"),
		"HELM_HOME":  ".helm",
	})
	valuesFile := text.ToFile(values, ".yml")
	if !log.IsLevelEnabled(log.TraceLevel) {
		defer os.Remove(valuesFile)
	}
	ca := p.TrustedCA
	if p.TrustedCA != "" {
		ca = fmt.Sprintf("--ca-file \"%s\"", p.TrustedCA)
	}

	err = helm("init -c --skip-refresh=true")
	if err != nil {
		return fmt.Errorf("deploy: failed to init helm %v", err)
	}
	debug := ""
	if log.IsLevelEnabled((log.TraceLevel)) {
		debug = "--debug"
	}

	if err := helm("upgrade harbor build/harbor --wait -f %s --install --namespace harbor %s %s", valuesFile, ca, debug); err != nil {
		return fmt.Errorf("deploy: failed to deploy harbor helm chart %v", err)
	}

	client := NewHarborClient(p)
	return client.UpdateSettings(*p.Harbor.Settings)

}
