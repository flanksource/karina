package base

import (
	"context"
	"fmt"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
	"github.com/flanksource/kommons"
	"github.com/flanksource/konfigadm/pkg/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func Test(platform *platform.Platform, test *console.TestResults) {
	client, err := platform.GetClientset()
	if err != nil {
		test.Errorf("Base tests failed to get clientset: %v", err)
		return
	}
	if client == nil {
		test.Errorf("Base tests failed to get clientset: nil clientset ")
		return
	}

	kommons.TestNamespace(client, "kube-system", test)

	if platform.E2E {
		TestImportSecrets(platform, test)
	}
}

type testSecretsFn func(*console.TestResults, *platform.Platform) bool

type testSecret struct {
	Name string
	Data map[string]string
}

type testSecretsFixture struct {
	Name           string
	Secrets        []testSecret
	PlatformConfig *types.PlatformConfig
	ValidateFn     testSecretsFn
}

func TestImportSecrets(p *platform.Platform, test *console.TestResults) {
	ctx := context.Background()
	clientset, err := p.GetClientset()
	if err != nil {
		test.Failf("base", "failed to get clientset: %v", err)
		return
	}

	ns := fmt.Sprintf("e2e-import-secrets-%s", utils.RandomString(4))
	namespace := &v1.Namespace{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{Name: ns},
	}
	if _, err := clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{}); err != nil {
		test.Failf("base", "failed to create namespace %s", ns)
		return
	}
	defer func() {
		clientset.CoreV1().Namespaces().Delete(context.Background(), ns, metav1.DeleteOptions{}) // nolint: errcheck
	}()

	templateOperatorVersion := p.TemplateOperator.Version
	platformWithMonitoringEnabled := newPlatformWithMonitoring(p, false)
	platformWithMonitoringDisabled := newPlatformWithMonitoring(p, true)

	fixtures := []testSecretsFixture{
		{
			Name: "TestXDisabled",
			Secrets: []testSecret{
				{
					Name: "config-t1-1",
					Data: map[string]string{
						"templateOperator.version":  "v2.3.4",
						"templateOperator.disabled": "true",
					},
				},
			},
			ValidateFn: func(test *console.TestResults, p *platform.Platform) bool {
				if p.TemplateOperator.Version != "v2.3.4" {
					test.Failf("base", "Expected templateOperator.Version to equal v2.3.4 got %s", p.TemplateOperator.Version)
					return false
				}
				if p.TemplateOperator.Disabled != "true" {
					test.Failf("base", "Expected templateOperator.Disabled to equal true got %s", p.TemplateOperator.Disabled)
					return false
				}
				return true
			},
		},
		{
			Name: "TestXDisabledOnlyVersion",
			Secrets: []testSecret{
				{
					Name: "config-t2-1",
					Data: map[string]string{
						"templateOperator.version": "v2.3.4",
					},
				},
			},
			ValidateFn: func(test *console.TestResults, p *platform.Platform) bool {
				if p.TemplateOperator.Version != "v2.3.4" {
					test.Failf("base", "Expected templateOperator.Version to equal v2.3.4 got %s", p.TemplateOperator.Version)
					return false
				}
				if p.TemplateOperator.IsDisabled() != false {
					test.Failf("base", "Expected templateOperator.IsDisabled to equal false got %t", p.TemplateOperator.IsDisabled())
					return false
				}
				return true
			},
		},
		{
			Name: "TestXDisabledOnlyDisabled",
			Secrets: []testSecret{
				{
					Name: "config-t3-1",
					Data: map[string]string{
						"templateOperator.disabled": "false",
					},
				},
			},
			ValidateFn: func(test *console.TestResults, p *platform.Platform) bool {
				if p.TemplateOperator.Version != templateOperatorVersion {
					test.Failf("base", "Expected templateOperator.Version to equal %s got %s", templateOperatorVersion, p.TemplateOperator.Version)
					return false
				}
				if p.TemplateOperator.IsDisabled() != false {
					test.Failf("base", "Expected templateOperator.Disabled to equal false got %s", p.TemplateOperator.IsDisabled())
					return false
				}
				return true
			},
		},
		{
			Name: "TestDisabledOnMonitoring",
			Secrets: []testSecret{
				{
					Name: "config-tmonitoring-1",
					Data: map[string]string{
						"monitoring.disabled": "false",
					},
				},
			},
			PlatformConfig: platformWithMonitoringDisabled,
			ValidateFn: func(test *console.TestResults, p *platform.Platform) bool {
				if !p.Monitoring.Disabled {
					test.Failf("base", "Expected monitoring to be disabled got %t", p.Monitoring.Disabled)
					return false
				}
				return true
			},
		},
		{
			Name: "TestDisabledOnMonitoringFalse",
			Secrets: []testSecret{
				{
					Name: "config-tmonitoring-2",
					Data: map[string]string{
						"monitoring.disabled": "true",
					},
				},
			},
			PlatformConfig: platformWithMonitoringEnabled,
			ValidateFn: func(test *console.TestResults, p *platform.Platform) bool {
				if p.Monitoring.Disabled {
					test.Failf("base", "Expected monitoring to not be disabled got %t", p.Monitoring.Disabled)
					return false
				}
				return true
			},
		},
		{
			Name: "TestMinio",
			Secrets: []testSecret{
				{
					Name: "config-t4-1",
					Data: map[string]string{
						"minio.yaml": `
minio:
  version: v1.2.3.4
  access_key: foo
  secret_key: bar
  replicas: 123
`,
					},
				},
			},
			ValidateFn: func(test *console.TestResults, p *platform.Platform) bool {
				if p.Minio.Version != "v1.2.3.4" {
					test.Failf("base", "Expected minio.version to equal v1.2.3.4 got %s", p.Minio.Version)
					return false
				}
				if p.Minio.AccessKey != "foo" {
					test.Failf("base", "Expected minio.access_key to equal foo got %s", p.Minio.AccessKey)
					return false
				}
				if p.Minio.SecretKey != "bar" {
					test.Failf("base", "Expected minio.secret_key to equal foo got %s", p.Minio.SecretKey)
					return false
				}
				if p.Minio.Replicas != 123 {
					test.Failf("base", "Expected minio.replicas to equal 123 got %d", p.Minio.Replicas)
					return false
				}
				return true
			},
		},
		{
			Name: "TestMinioMultipleSecrets",
			Secrets: []testSecret{
				{
					Name: "config-t5-1",
					Data: map[string]string{
						"minio.yaml": `
minio:
  disabled: true,
  version: v1.2.3.4
  access_key: foo
  secret_key: bar
  replicas: 123
`,
					},
				},
				{
					Name: "config-t5-2",
					Data: map[string]string{
						"minio.disabled": "false",
					},
				},
			},
			ValidateFn: func(test *console.TestResults, p *platform.Platform) bool {
				if p.Minio.Version != "v1.2.3.4" {
					test.Failf("base", "Expected minio.version to equal v1.2.3.4 got %s", p.Minio.Version)
					return false
				}
				if p.Minio.AccessKey != "foo" {
					test.Failf("base", "Expected minio.access_key to equal foo got %s", p.Minio.AccessKey)
					return false
				}
				if p.Minio.SecretKey != "bar" {
					test.Failf("base", "Expected minio.secret_key to equal foo got %s", p.Minio.SecretKey)
					return false
				}
				if p.Minio.Replicas != 123 {
					test.Failf("base", "Expected minio.replicas to equal 123 got %d", p.Minio.Replicas)
					return false
				}
				return true
			},
		},
	}

	for _, fixture := range fixtures {
		secrets := []v1.SecretReference{}
		for _, s := range fixture.Secrets {
			secret := &v1.Secret{
				TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
				ObjectMeta: metav1.ObjectMeta{Name: s.Name, Namespace: ns},
				StringData: s.Data,
			}
			if _, err := clientset.CoreV1().Secrets(ns).Create(ctx, secret, metav1.CreateOptions{}); err != nil {
				test.Failf("base", "failed to create secret %s: %v", s.Name, err)
			}
			secrets = append(secrets, v1.SecretReference{Name: s.Name, Namespace: ns})
		}

		testImportSecrets(p, fixture.PlatformConfig, test, fixture.Name, secrets, fixture.ValidateFn)
	}
}

func testImportSecrets(p *platform.Platform, pp *types.PlatformConfig, test *console.TestResults, name string, secrets []v1.SecretReference, validateFn testSecretsFn) {
	var newPlatformConfig *types.PlatformConfig
	var err error
	if pp != nil {
		newPlatformConfig = pp
	} else {
		newPlatformConfig, err = clonePlatform(p)
		if err != nil {
			test.Failf("base", "failed to clone platform: %v", err)
			return
		}
	}
	newPlatformConfig.ImportSecrets = secrets
	newP := &platform.Platform{
		PlatformConfig: *newPlatformConfig,
	}
	if p.KubeConfigPath != "" {
		newP.KubeConfigPath = p.KubeConfigPath
	}
	if err := newP.Init(); err != nil {
		test.Failf("base", "failed to init platform: %v", err)
		return
	}

	if ok := validateFn(test, newP); ok {
		test.Passf("base", "Imported secret for test %s successfully", name)
	}
}

func newPlatformWithMonitoring(p *platform.Platform, disabled bool) *types.PlatformConfig {
	newP, _ := clonePlatform(p)
	newP.Monitoring.Disabled = types.Boolean(disabled)
	return newP
}

func clonePlatform(platform *platform.Platform) (*types.PlatformConfig, error) {
	yml := platform.String()
	p := &types.PlatformConfig{}
	if err := yaml.Unmarshal([]byte(yml), p); err != nil {
		return nil, err
	}
	return p, nil
}
