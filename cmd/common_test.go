package cmd_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/flanksource/karina/cmd"
	"github.com/flanksource/karina/pkg/types"
	. "github.com/onsi/gomega"
)

func TestGetConfigSimple(t *testing.T) {
	cfg, g := newFixture([]string{"simple.yaml"}, t)

	g.Expect(cfg.Name).To(Equal("test"))
	g.Expect(cfg.Domain).To(Equal("127.0.0.1.nip.io"))
	g.Expect(cfg.Kubernetes.Version).To(Equal("v1.15.7"))
}

func TestGetConfigSetDefaults(t *testing.T) {
	cfg, g := newFixture([]string{"simple.yaml"}, t)

	g.Expect(cfg.Ldap.GroupObjectClass).To(Equal("group"))
	g.Expect(cfg.Ldap.GroupNameAttr).To(Equal("name"))
	g.Expect(cfg.Kubernetes).To(Equal(types.Kubernetes{
		Version:             "v1.15.7",
		APIServerExtraArgs:  map[string]string{},
		ControllerExtraArgs: map[string]string{},
		SchedulerExtraArgs:  map[string]string{},
		KubeletExtraArgs:    map[string]string{},
		EtcdExtraArgs:       map[string]string{},
		ContainerRuntime:    "docker",
	}))
}

func TestGetConfigOverwriteDefaults(t *testing.T) {
	cfg, g := newFixture([]string{"ldap.yaml"}, t)

	g.Expect(cfg.Ldap.GroupObjectClass).To(Equal("groupOfNames"))
	g.Expect(cfg.Ldap.GroupNameAttr).To(Equal("DN"))
	g.Expect(cfg.Ldap.Username).To(Equal("uid=admin,ou=system"))
}

func TestMergeTwoConfigs(t *testing.T) {
	cfg, g := newFixture([]string{"simple.yaml", "overwrite.yaml"}, t)

	g.Expect(cfg.Kubernetes.Version).To(Equal("v1.15.7"))
	g.Expect(cfg.Calico.Version).To(Equal("v3.8.2"))
}

func TestUnrecognizedConfig(t *testing.T) {
	// Work around for testing exit -1 scenario
	// https://stackoverflow.com/questions/26225513/how-to-test-os-exit-scenarios-in-go#33404435
	if os.Getenv("MAKE_CONFIG") == "1" {
		_, _ = newFixture([]string{"unrecognized-fields.yaml"}, t)
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestUnrecognizedConfig")
	cmd.Env = append(os.Environ(), "MAKE_CONFIG=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("Process ran with err %v, want exit status 1", err)
}

func newFixture(paths []string, t *testing.T) (*types.PlatformConfig, *WithT) {
	g := NewWithT(t)
	fullPaths := make([]string, len(paths))
	for i := range paths {
		fullPaths[i] = fmt.Sprintf("../test/fixtures/%s", paths[i])
	}

	cfg := cmd.NewConfig(fullPaths, []string{})
	return &cfg, g
}
