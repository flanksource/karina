package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/kommons"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var paths = []string{"../test/elastic.yaml"}
var extras []string
var config = NewConfig(paths, extras)

var APIServerDefaultArgs = []string{
	"--advertise-address=127.0.0.1",
	"--etcd-servers={{ if .EtcdURL }}{{ .EtcdURL.String }}{{ end }}",
	"--cert-dir={{ .CertDir }}",
	"--insecure-port={{ if .URL }}{{ .URL.Port }}{{ end }}",
	"--insecure-bind-address={{ if .URL }}{{ .URL.Hostname }}{{ end }}",
	"--secure-port={{ if .SecurePort }}{{ .SecurePort }}{{ end }}",
	// we're keeping this disabled because if enabled, default SA is missing which would force all tests to create one
	// in normal apiserver operation this SA is created by controller, but that is not run in integration environment
	//"--disable-admission-plugins=ServiceAccount",
	"--service-cluster-ip-range=10.0.0.0/24",
	"--allow-privileged=true",
}

func Test1(t *testing.T) {
	os.Setenv("TEST_ASSET_KUBE_APISERVER", "/tmp/kubebuilder/bin/kube-apiserver")
	os.Setenv("TEST_ASSET_ETCD", "/tmp/kubebuilder/bin/etcd")
	os.Setenv("TEST_ASSET_KUBECTL", "/tmp/kubebuilder/bin/kubectl")
	os.Setenv("KUBEBUILDER_CONTROLPLANE_START_TIMEOUT", "5m")
	os.Setenv("KUBEBUILDER_CONTROLPLANE_STOP_TIMEOUT", "5m")

	var testEnv = &envtest.Environment{
		CRDDirectoryPaths:  []string{filepath.Join("..", "config", "crd", "bases")},
		KubeAPIServerFlags: APIServerDefaultArgs,
	}

	cfg, err1 := testEnv.Start()
	fmt.Print(cfg, err1)
	tests := []struct {
		name string
	}{
		{name: "foo"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := logger.StandardLogger()
			testClient := kommons.NewClient(cfg, l)
			err2 := testClient.CreateOrUpdateNamespace("foo", map[string]string{}, map[string]string{})
			if err2 != nil {
				fmt.Print("ERROR!", err2)
				return
			}

			err := testClient.Apply("foo", &v1.ConfigMap{
				TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bar",
					Namespace: "foo",
				},
				Data: map[string]string{"foo": "lala"},
			})
			if err != nil {
				fmt.Print("ERROR!", err)
				return
			}

			_, value, _ := testClient.GetEnvValue(config.Hooks["elastic"].Before, "foo")
			err3 := testClient.ApplyText("foo", value)
			if err3 != nil {
				fmt.Print("ERROR!", err3)
				return
			}
			assert.Equal(t, value, 1, "Output")
			_ = testEnv.Stop()
		})
	}
}