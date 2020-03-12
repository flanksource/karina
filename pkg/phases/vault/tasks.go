package vault

import (
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"

	"github.com/moshloop/platform-cli/pkg/platform"
)

func Init(p *platform.Platform) error {
	if err := p.WaitForPod("vault", "vault-0", 120*time.Second, v1.PodRunning); err != nil {
		return err
	}

	stdout, stderr, err := p.ExecutePodf("vault", "vault-0", "vault", "/bin/vault", "operator", "init", "-tls-skip-verify", "-status")
	log.Debugf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
	if err != nil && strings.Contains(stdout, "Vault is not initialized") {
		log.Infof("Vault is not initialized, initializing")
		stdout, stderr, err = p.ExecutePodf("vault", "vault-0", "vault", "/bin/vault", "operator", "init", "-tls-skip-verify")
		log.Infof("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
	} else if strings.Contains(stdout, "Vault is initialized") {
		log.Infof("Vault is initialized.")
	}

	return nil
}
