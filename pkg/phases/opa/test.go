package opa

import (
	// "fmt"
	"io/ioutil"

	"github.com/moshloop/commons/console"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
	log "github.com/sirupsen/logrus"
)

func generateFixtures(kind, attribute string) {

}

func TestNamespace(p *platform.Platform, test *console.TestResults) {
	client, _ := p.GetClientset()
	k8s.TestNamespace(client, "opa", test)
}

func TestPolicies(p *platform.Platform, fixturesPath string, test *console.TestResults) {
	kubectl := p.GetKubectl()
	kubectl("apply -f test/opa/namespaces/")
	kubectl("apply -f test/opa/ingress-duplicate.yaml")
	// time.Sleep(60 * time.Second)
	rejectedFixturesPath := fixturesPath + "/rejected"
	acceptedFixturesPath := fixturesPath + "/accepted"

	rejectedFixtureFiles, err := ioutil.ReadDir(rejectedFixturesPath)
	if err != nil {
		log.Fatal(err)
	}

	acceptedFixtureFiles, err := ioutil.ReadDir(acceptedFixturesPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, rejectedFixture := range rejectedFixtureFiles {
		if err := kubectl("apply -f %s &> /dev/null", rejectedFixturesPath+"/"+rejectedFixture.Name()); err != nil {
			test.Passf(rejectedFixture.Name(), "%s rejected by Gatekeeper as expected", rejectedFixture.Name())
		} else {
			test.Failf(rejectedFixture.Name(), "%s accepted by Gatekeeper as not expected", rejectedFixture.Name())

		}
	}

	for _, acceptedFixture := range acceptedFixtureFiles {
		if err := kubectl("apply -f %s &> /dev/null", acceptedFixturesPath+"/"+acceptedFixture.Name()); err != nil {
			test.Failf(acceptedFixture.Name(), "%s rejected by Gatekeeper as not expected", acceptedFixture.Name())
		} else {
			test.Passf(acceptedFixture.Name(), "%s accepted by Gatekeeper as expected", acceptedFixture.Name())
		}
	}
}
