package audit

import (
	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/platform"
	"time"
     "encoding/json"
)


const JUnitAuditConfigClass = "AuditConfig"
const jUnitAuditApiServerStateClass = "AuditApiServerState"

func Test(p *platform.Platform, tr *console.TestResults) {
	//tr.Passf("PassTest","Hello %v", "test")
	//tr.Failf("FailTest","Hello %v", "test")

	ac := p.AuditConfig

	if ac == nil {
		tr.Failf(JUnitAuditConfigClass,"There is no AuditConfig")
		tr.Done()
	}

	if ac.Disabled {
		tr.Skipf(JUnitAuditConfigClass,"AuditConfig is disabled")
		tr.Done()
	}

	master, err := p.GetMasterNode()
	tr.Skipf("%v", master)
	if err != nil {
		tr.Failf(JUnitAuditConfigClass,"Failed to get a defined master node: %v",  err)
		return
	}

	if p.AuditConfig.ApiServerOptions.LogOptions.Path != "" {
		//stdout, err := p.ExecutePodf(master, 2*time.Minute, "crictl ps --name kube-apiserver -o json")
		//if err != nil {
		//	tr.Failf(jUnitAuditApiServerStateClass,"Failed to query for apiserver, crictl invocation error: %v",  err)
		//	return
		//}
		//p.Get("kube-system", "")


		err := p.WaitFor()
		if err != nil {
			tr.Failf(jUnitAuditApiServerStateClass,"Failed to get a active master node: %v",  err)
			return
		}

	}

}
