package opa

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/k8s"
	"github.com/flanksource/karina/pkg/platform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type OpaError struct {
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

func (e OpaError) String() string {
	msg := fmt.Sprintf("%s: %s", e.Error.Code, e.Error.Message)
	for _, item := range e.Error.Errors {
		msg += fmt.Sprintf(" (%s:%d:%d) %s:%s", item.Location.File, item.Location.Row, item.Location.Col, item.Code, item.Message)
	}
	return msg
}

func Test(p *platform.Platform, test *console.TestResults) {
	if p.OPA != nil && p.OPA.Disabled {
		test.Skipf("opa", "OPA is not configured")
		return
	}

	client, _ := p.GetClientset()
	k8s.TestNamespace(client, Namespace, test)
	configs, err := client.CoreV1().ConfigMaps(Namespace).List(metav1.ListOptions{})
	if err != nil {
		test.Failf(Namespace, "failed to list policies via configmap: %v", err)
	} else {
		for _, cm := range configs.Items {
			status, ok := cm.Annotations["openpolicyagent.org/policy-status"]
			if !ok || cm.Name == "opa-config" {
				// not an OPA policy
				continue
			}
			opaError := OpaError{}
			json.Unmarshal([]byte(status), &opaError)
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
