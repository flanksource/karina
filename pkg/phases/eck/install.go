package eck

import (
	"fmt"
	"github.com/flanksource/commons/net"
	"github.com/moshloop/platform-cli/pkg/platform"
	log "github.com/sirupsen/logrus"
	"strings"
)

func Deploy(p *platform.Platform) error {
	if p.ECK == nil || p.ECK.Disabled {
		log.Infof("Skipping deployment of ECK, it is disabled")
		return nil
	} else {
		log.Infof("Deploying ECK %s", p. ECK.Version)
	}
	if err := net.Download("https://download.elastic.co/downloads/eck/"+normalizeVersion(p.ECK.Version + "/all-in-one.yaml"), "build/eck.yaml"); err != nil {
		return fmt.Errorf("deploy: failed to download ECK: %v", err)
	}
	kubectl := p.GetKubectl()
	return kubectl("apply -f build/eck.yaml")
	return nil
}

func normalizeVersion(version string) string {
	if strings.HasPrefix(version, "v") {
		return strings.TrimSuffix(version, "v")
	}
	return version
}
