package opa

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

const e2eAnnotation = "gatekeeper.flanksource.com/e2e"

type Error struct {
	Status string `json:"status"`
	Error  struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Errors  []struct {
			Code     string `json:"code"`
			Message  string `json:"message"`
			Location struct {
				File string `json:"file"`
				Row  int    `json:"row"`
				Col  int    `json:"col"`
			} `json:"location"`
		} `json:"errors"`
	} `json:"error"`
}

func (e Error) String() string {
	msg := fmt.Sprintf("%s: %s", e.Error.Code, e.Error.Message)
	for _, item := range e.Error.Errors {
		msg += fmt.Sprintf(" (%s:%d:%d) %s:%s", item.Location.File, item.Location.Row, item.Location.Col, item.Code, item.Message)
	}
	return msg
}

func Test(p *platform.Platform, test *console.TestResults) {
	TestOPA(p, test)
	TestGatekeeper(p, test)
}

func TestOPA(p *platform.Platform, test *console.TestResults) {
	if p.OPA != nil && p.OPA.Disabled {
		test.Skipf("opa", "OPA is not configured")
		return
	}

	client, err := p.GetClientset()

	if err != nil {
		test.Failf(Namespace, "Could not connect to Platform client: %v", err)
		return
	}

	kommons.TestNamespace(client, Namespace, test)
	configs, err := client.CoreV1().ConfigMaps(Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		test.Failf(Namespace, "failed to list policies via configmap: %s", err)
	} else {
		for _, cm := range configs.Items {
			status, ok := cm.Annotations["openpolicyagent.org/policy-status"]
			if !ok || cm.Name == "opa-config" {
				// not an OPA policy
				continue
			}
			opaError := Error{}
			_ = json.Unmarshal([]byte(status), &opaError)
			if opaError.Status == "ok" {
				test.Passf(Namespace, "OPA policy %s loaded successfully", cm.Name)
			} else {
				test.Failf(Namespace, "OPA policy %s did not load: %s", cm.Name, opaError)
			}
		}
	}
	if p.E2E {
		testE2E(p, test)
	}
}

func testE2E(p *platform.Platform, test *console.TestResults) {
	if p.OPA == nil || p.OPA.Disabled {
		test.Skipf("opa", "OPA is not configured")
		return
	}

	if p.OPA.E2E.Fixtures == "" {
		test.Skipf("opa", "OPA fixtures path not configured under opa.e2e.fixtures")
		return
	}

	kubectl := p.GetKubectl()
	kubeconfig, err := p.GetKubeConfig()
	if err != nil {
		test.Failf("opa", "Failed to get kube config: %v", err)
		return
	}

	if err := kubectl("apply -f %s/resources --kubeconfig %s", p.OPA.E2E.Fixtures, kubeconfig); err != nil {
		test.Failf("opa", "Failed to setup namespaces: %v", err)
		return
	}
	defer func() {
		for _, path := range []string{"resources", "accepted", "rejected"} {
			kubectl("delete -f %s/%s --force  &> /dev/null", p.OPA.E2E.Fixtures, path) //nolint errcheck
		}
	}()

	rejectedFixturesPath := p.OPA.E2E.Fixtures + "/rejected"
	acceptedFixturesPath := p.OPA.E2E.Fixtures + "/accepted"

	rejectedFixtureFiles, err := ioutil.ReadDir(rejectedFixturesPath)
	if err != nil {
		test.Failf("opa", "Install: Failed to read dir: %s", err)
		return
	}

	acceptedFixtureFiles, err := ioutil.ReadDir(acceptedFixturesPath)
	if err != nil {
		test.Failf("opa", "Failed to list accepted fixtures: %v", err)
		return
	}

	for _, rejectedFixture := range rejectedFixtureFiles {
		if err := kubectl("apply -f %s --kubeconfig %s &> /dev/null", rejectedFixturesPath+"/"+rejectedFixture.Name(), kubeconfig); err != nil {
			test.Passf(rejectedFixture.Name(), "%s rejected as expected", rejectedFixture.Name())
		} else {
			test.Failf(rejectedFixture.Name(), "%s accepted as not expected", rejectedFixture.Name())
		}
	}

	for _, acceptedFixture := range acceptedFixtureFiles {
		if err := kubectl("apply -f %s --kubeconfig %s &> /dev/null", acceptedFixturesPath+"/"+acceptedFixture.Name(), kubeconfig); err != nil {
			test.Failf(acceptedFixture.Name(), "%s rejected as not expected", acceptedFixture.Name())
		} else {
			test.Passf(acceptedFixture.Name(), "%s accepted as expected", acceptedFixture.Name())
		}
	}
}

func TestGatekeeper(p *platform.Platform, test *console.TestResults) {
	if p.Gatekeeper.IsDisabled() {
		test.Skipf("opa", "Gatekeeper is not configured")
		return
	}

	client, _ := p.GetClientset()
	kommons.TestNamespace(client, GatekeeperNamespace, test)
	if p.E2E {
		testE2EGatekeeper(p, test)
	}
}

type Fixture struct {
	Kind     string            `yaml:"kind,omitempty"`
	Metadata metav1.ObjectMeta `yaml:"metadata,omitemoty"`
}

type ViolationConfig struct {
	Violations []Violation `yaml:"violations,omitempty"`
}

type Violation struct {
	Kind    string `yaml:"kind,omitempty"`
	Name    string `yaml:"name,omitempty"`
	Message string `yaml:"message,omitempty"`
}

type AuditResource struct {
	Status AuditResourceStatus `yaml:"status,omitempty"`
}

type AuditResourceStatus struct {
	Violations []AuditResourceViolation `yaml:"violations,omitempty"`
}

type AuditResourceViolation struct {
	Kind      string `yaml:"kind,omitempty"`
	Name      string `yaml:"name,omitempty"`
	Namespace string `yaml:"namespace,omitempty"`
	Message   string `yaml:"message,omitempty"`
}

func testE2EGatekeeper(p *platform.Platform, test *console.TestResults) {
	if p.Gatekeeper.IsDisabled() {
		test.Skipf("opa", "Gatekeeper is not configured")
		return
	}

	if p.Gatekeeper.E2E.Fixtures == "" {
		test.Skipf("opa", "OPA fixtures path not configured under gatekeeper.e2e.fixtures")
		return
	}

	kubectl := p.GetKubectl()
	kubeconfig, err := p.GetKubeConfig()
	if err != nil {
		test.Failf("opa", "Failed to get kube config: %v", err)
		return
	}

	if err := kubectl("apply -f %s/resources --kubeconfig %s", p.OPA.E2E.Fixtures, kubeconfig); err != nil {
		test.Failf("opa", "Failed to setup namespaces: %v", err)
		return
	}
	defer func() {
		for _, path := range []string{"resources", "accepted", "rejected"} {
			kubectl("delete -f %s/%s --force  &> /dev/null", p.OPA.E2E.Fixtures, path) //nolint errcheck
		}
	}()

	rejectedFixturesPath := p.Gatekeeper.E2E.Fixtures + "/rejected"

	rejectedFixtureFiles, err := ioutil.ReadDir(rejectedFixturesPath)
	if err != nil {
		test.Failf("opa", "Install: Failed to read dir: %s", err)
		return
	}

	for _, rejectedFixture := range rejectedFixtureFiles {
		filename := rejectedFixturesPath + "/" + rejectedFixture.Name()
		if err := kubectl("apply -f %s --kubeconfig %s &> /dev/null", filename, kubeconfig); err != nil {
			test.Failf(rejectedFixture.Name(), "%s rejected by admission controller: %v", rejectedFixture.Name(), err)
		}

		fileContents, err := ioutil.ReadFile(filename)
		if err != nil {
			test.Failf(rejectedFixture.Name(), "failed to read file contents")
		}

		object := &Fixture{}
		if err := yaml.Unmarshal(fileContents, object); err != nil {
			test.Failf(rejectedFixture.Name(), "failed to unmarshal yaml")
		}

		configFile, found := object.Metadata.Annotations[e2eAnnotation]
		if !found {
			test.Failf(rejectedFixture.Name(), "failed to find annotation %s", e2eAnnotation)
		}

		config := &ViolationConfig{}
		if err := yaml.Unmarshal([]byte(configFile), config); err != nil {
			test.Failf(rejectedFixture.Name(), "failed to read violation config")
		}

		for _, violation := range config.Violations {
			timeout := time.Now().Add(120 * time.Second)
			findViolationUntil(p, test, violation, object, timeout)
		}
	}
}

func findViolationUntil(p *platform.Platform, test *console.TestResults, violation Violation, object *Fixture, timeout time.Time) {
	dynamicClient, err := p.GetDynamicClient()
	if err != nil {
		test.Failf("opa", "failed to get dynamic client: %v", err)
		return
	}
	rm, _ := p.GetRestMapper()
	gvk, err := rm.KindFor(schema.GroupVersionResource{
		Group:    "constraints.gatekeeper.sh",
		Version:  "v1beta1",
		Resource: violation.Kind,
	})
	if err != nil {
		test.Failf("opa", "failed to get rest mapper for kind %s: %v", violation.Kind, err)
		return
	}
	gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}
	mapping, err := rm.RESTMapping(gk, gvk.Version)
	if err != nil {
		test.Failf("opa", "failed to get rest mapping for kind %s: %v", violation.Kind, err)
		return
	}
	client := dynamicClient.Resource(mapping.Resource)

	for {
		if time.Now().After(timeout) {
			test.Failf("opa", "received timeout waiting for violation %s for %s/%s/%s", violation.Kind, object.Kind, object.Metadata.Namespace, object.Metadata.Name)
			return
		}
		found, err := findViolation(client, violation, object)
		p.Debugf("violation: %s found=%t err=%v", violation.Name, found, err)
		if err != nil {
			test.Failf("opa", "failed to find violation: %v", err)
			return
		}

		if found {
			test.Passf("opa", "found violation %s for %s/%s/%s", violation.Kind, object.Kind, object.Metadata.Namespace, object.Metadata.Name)
			return
		}
		time.Sleep(5 * time.Second)
	}
}

func findViolation(client dynamic.NamespaceableResourceInterface, violation Violation, object *Fixture) (bool, error) {
	obj, err := client.Get(context.TODO(), violation.Name, metav1.GetOptions{})
	if err != nil {
		return false, errors.Wrapf(err, "failed to get %s with name=%s", violation.Kind, violation.Name)
	}

	yml, err := yaml.Marshal(obj.Object)
	if err != nil {
		return false, errors.Wrapf(err, "failed to yaml encode object")
	}

	ar := &AuditResource{}
	if err := yaml.Unmarshal(yml, ar); err != nil {
		return false, errors.Wrapf(err, "failed to unmarshal into AuditResource")
	}

	foundWithDifferentMessage := false
	foundMessage := ""

	for _, v := range ar.Status.Violations {
		if v.Kind == object.Kind && v.Name == object.Metadata.Name && v.Namespace == object.Metadata.Namespace {
			if v.Message == violation.Message {
				return true, nil
			}
			foundWithDifferentMessage = true
			foundMessage = v.Message
		}
	}

	if foundWithDifferentMessage {
		return false, errors.Errorf("expected violation %s for %s/%s/%s to have message %s, got %s", violation.Name, object.Kind, object.Metadata.Namespace, object.Metadata.Name, violation.Message, foundMessage)
	}

	return false, nil
}
