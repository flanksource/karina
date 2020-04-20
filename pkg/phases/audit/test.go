package audit

import (
	"path/filepath"

	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/platform"
)

const testName = "auditing"

// Test k8s auditing functionality.
func Test(p *platform.Platform, tr *console.TestResults) {
	if p.Kubernetes.AuditConfig == nil {
		tr.Skipf(testName, "auditing not configured")
		return
	}
	pf := p.Kubernetes.AuditConfig.PolicyFile

	if pf == "" {
		tr.Skipf(testName, "No audit policy specified.")
		return
	}

	_, err := p.GetClientset()
	if err != nil {
		tr.Failf(testName, "Failed to get k8s client: %v", err)
		// We're done, we can't test anything further.
		return
	}

	pod, err := p.Client.GetFirstPodByLabelSelector("kube-system", "component=kube-apiserver")
	if err != nil {
		tr.Failf(testName, "Failed to get api-server pod: %v", err)
		return
	}
	tr.Passf(testName, "api-server pod found")

	if logFilePath, ok := p.Kubernetes.APIServerExtraArgs["audit-log-path"]; !ok {
		tr.Failf(testName, "No audit-log-path is specified!")
		return
	} else if logFilePath == "" {
		tr.Failf(testName, "Empty audit-log-path is specified!")
		return
	} else if logFilePath == "-" {
		tr.Skipf(testName, "api-server is configured lo log to stdout, not verifying output")
		return
	} else {
		dir := filepath.Dir(logFilePath)
		stdout, stderr, err := p.ExecutePodf("kube-system", pod.Name, "kube-apiserver", "/usr/bin/du", "-s", dir)
		if err != nil || stderr != "" {
			tr.Failf(testName, "Failed to get file size statistics: %v\n%v", err, stderr)
		} else {
			tr.Passf(testName, "api-server pod log size is: %v", stdout)
		}
	}
}
