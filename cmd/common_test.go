package cmd_test

import (
	"fmt"
	"testing"

	"github.com/moshloop/platform-cli/cmd"
	"github.com/moshloop/platform-cli/pkg/types"
	. "github.com/onsi/gomega"
)

func TestGetConfigSimple(t *testing.T) {
	cfg, g := newFixture([]string{"simple.yml"}, t)

	g.Expect(cfg.Name).To(Equal("test"))
	g.Expect(cfg.Domain).To(Equal("127.0.0.1.nip.io"))
	g.Expect(cfg.Kubernetes).To(Equal(types.Kubernetes{Version: "v1.15.7"}))
}

func TestGetConfigSetDefaults(t *testing.T) {
	cfg, g := newFixture([]string{"simple.yml"}, t)

	g.Expect(cfg.Ldap.GroupObjectClass).To(Equal("group"))
	g.Expect(cfg.Ldap.GroupNameAttr).To(Equal("name"))
}

func TestGetConfigOverwriteDefaults(t *testing.T) {
	cfg, g := newFixture([]string{"ldap.yml"}, t)

	g.Expect(cfg.Ldap.GroupObjectClass).To(Equal("groupOfNames"))
	g.Expect(cfg.Ldap.GroupNameAttr).To(Equal("DN"))
	g.Expect(cfg.Ldap.Username).To(Equal("uid=admin,ou=system"))
}

func TestMergeTwoConfigs(t *testing.T) {
	cfg, g := newFixture([]string{"simple.yml", "overwrite.yml"}, t)

	g.Expect(cfg.Kubernetes.Version).To(Equal("v1.15.7"))
	g.Expect(cfg.Calico.Version).To(Equal("v3.8.2"))
}

func newFixture(paths []string, t *testing.T) (*types.PlatformConfig, *WithT) {
	g := NewWithT(t)
	fullPaths := make([]string, len(paths))
	for i := range paths {
		fullPaths[i] = fmt.Sprintf("../test/fixtures/%s", paths[i])
	}

	cfg := cmd.NewConfig(fullPaths, false, []string{}, false)
	return &cfg, g
}
