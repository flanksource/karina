package fluentdOperator

import (
	"fmt"
	"github.com/flanksource/commons/deps"
	"github.com/flanksource/commons/files"
	"github.com/moshloop/platform-cli/pkg/platform"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

func Deploy(p *platform.Platform) error {
	if p.FluentdOperator == nil || p.FluentdOperator.Disabled {
		log.Infof("Skipping deployment of fluentd-operator, it is disabled")
		return nil
	} else {
		log.Infof("Deploying fluentd-operator %s", p.FluentdOperator.Version)
	}
	if err := files.Getter("git::https://github.com/vmware/kube-fluentd-operator.git?ref="+normalizeTag(p.FluentdOperator.Version), "build/kube-fluentd-operator"); err != nil {
		return fmt.Errorf("deploy: failed to download fluentd-operator: %v", err)
	}
	kubeconfig, err := p.GetKubeConfig()
	if err != nil {
		return fmt.Errorf("deploy: failed to get kubeconfig: %v", err)
	}
	helm := deps.BinaryWithEnv("helm", p.Versions["helm"], ".bin", map[string]string{
		"KUBECONFIG": kubeconfig,
		"HOME":       os.ExpandEnv("$HOME"),
		"HELM_HOME":  ".helm",
	})
	err = helm("init -c --skip-refresh=true")
	if err != nil {
		return fmt.Errorf("deploy: failed to init helm: %v", err)
	}
	debug := ""
	if log.IsLevelEnabled(log.TraceLevel) {
		debug = "--debug"
	}
	ca := p.TrustedCA
	if p.TrustedCA != "" {
		ca = fmt.Sprintf("--ca-file \"%s\"", p.TrustedCA)
	}
	setValues := fmt.Sprintf("--set rbac.create=true --set image.tag=%s --set image.repository=%s", normalizeTag(p.FluentdOperator.Version), p.FluentdOperator.ImageRepo)
	if err := helm("upgrade kube-fluentd-operator --wait  build/kube-fluentd-operator/log-router --install --namespace kube-fluentd-operator %s %s %s", ca, debug, setValues); err != nil {
		return fmt.Errorf("deploy: failed to deploy/upgrade fluentd-operator helm chart: %v", err)
	}
	return nil
}

func normalizeTag(tag string) string {
	if !strings.HasPrefix(tag, "v") {
		return "v"+tag
	}
	return tag
}