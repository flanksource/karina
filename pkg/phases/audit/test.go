package audit

import (
	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/platform"
	"path/filepath"


)


const JUnitAuditConfigClass = "AuditConfig"
const jUnitAuditApiServerStateClass = "AuditApiServerState"

func Test(p *platform.Platform, tr *console.TestResults) {
	ac := p.AuditConfig

	if ac == nil {
		// nil is a failure as the implicit default if it is not specified
		// should be 'disabled'
		tr.Failf(JUnitAuditConfigClass,"There is no AuditConfig")
		tr.Done()
	}

	if ac.Disabled {
		tr.Skipf(JUnitAuditConfigClass,"AuditConfig is disabled")
		tr.Done()
	}

	if p.AuditConfig.ApiServerOptions.LogOptions.Path != "" {

		_, err := p.GetClientset()
		if err != nil {
			tr.Failf(JUnitAuditConfigClass,"Failed to get k8s client: %v",err)
			tr.Done()
		}

		pod, err := p.Client.GetFirstPodByLabelSelector("kube-system","component=kube-apiserver")
		if err != nil {
			tr.Failf(jUnitAuditApiServerStateClass,"Failed to get api-server pod: %v",err)
			tr.Done()
		}

		dir := filepath.Dir(ac.ApiServerOptions.LogOptions.Path)
		stdout, stderr, err := p.ExecutePodf("kube-system",pod, "kube-apiserver","/usr/bin/du","-s", dir  )
		if err != nil || stderr != ""{
			tr.Failf(jUnitAuditApiServerStateClass,"Failed to get api-server pod: %v\n%v",err,stderr)
			tr.Done()
		}
		tr.Passf(jUnitAuditApiServerStateClass,"api-server pod log size is: %v",stdout)

	}

}
