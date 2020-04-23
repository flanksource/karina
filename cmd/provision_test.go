package cmd

import (
	"bytes"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func toArgs(cmdString string) []string {
	return strings.Split(CreateRootCA," " )
}

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	os.Chdir(cwd)
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}
const CreateRootCA = "ca generate --name root-ca --cert-path .certs/root-ca.crt --private-key-path .certs/root-ca.key --password foobar  --expiry 1"
const CreateIngressCA = "ca generate --name ingress-ca --cert-path .certs/ingress-ca.crt --private-key-path .certs/ingress-ca.key --password foobar  --expiry 1"


func setupCAs() {
	executeCommand(GetRootCmd("Testing"),  toArgs(CreateRootCA)... )
	executeCommand(GetRootCmd("Testing"),  toArgs(CreateIngressCA)... )
}

var KindEncryptionPlatformFeature, _ = filepath.Abs("../test/platform-fixtures/kind-encryption.yaml")

func TestEncryptionYamlSerialises(t *testing.T) {
	args := []string{"provision", "kind-cluster", "-c", KindEncryptionPlatformFeature}
	root := GetRootCmd("Testing")
	root.SetArgs(args)
	platform := getPlatform(root)
	got := platform.Kubernetes.EncryptionConfig.EncryptionProviderConfigFile
	if want, _ :=filepath.Abs("test/fixtures/encryption-config.yaml"); got != want {
		t.Errorf("EncryptionProviderConfigFile: Wanted %v, Got %v", want, got)
	}
}

func TestProvisionKindWithEncryption(t *testing.T) {
	setupCAs()
	executeCommand(GetRootCmd("Testing"),  "provision", "kind-cluster", "-c", KindEncryptionPlatformFeature  )
}

var cwd = "."
func init() {
	cwd, _ = filepath.Abs(cwd)
}


