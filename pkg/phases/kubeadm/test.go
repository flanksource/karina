package kubeadm

import (
	"path/filepath"

	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/platform"
)

const (
	testAuditName      = "auditing"
	testEncryptionName = "encryption"
)

// Test k8s auditing functionality.
func TestAudit(p *platform.Platform, tr *console.TestResults) {
	pf := p.Kubernetes.AuditConfig.PolicyFile

	if pf == "" {
		tr.Skipf(testAuditName, "No audit policy specified.")
		return
	}

	_, err := p.GetClientset()
	if err != nil {
		tr.Failf(testAuditName, "Failed to get k8s client: %v", err)
		// We're done, we can't test anything further.
		return
	}

	pod, err := p.Client.GetFirstPodByLabelSelector("kube-system", "component=kube-apiserver")
	if err != nil {
		tr.Failf(testAuditName, "Failed to get api-server pod: %v", err)
		return
	}
	tr.Passf(testAuditName, "api-server pod found")

	if logFilePath, ok := p.Kubernetes.APIServerExtraArgs["audit-log-path"]; !ok {
		tr.Failf(testAuditName, "No audit-log-path is specified!")
		return
	} else if logFilePath == "" {
		tr.Failf(testAuditName, "Empty audit-log-path is specified!")
		return
	} else if logFilePath == "-" {
		tr.Skipf(testAuditName, "api-server is configured lo log to stdout, not verifying output")
		return
	} else {
		dir := filepath.Dir(logFilePath)
		stdout, stderr, err := p.ExecutePodf("kube-system", pod.Name, "kube-apiserver", "/usr/bin/du", "-s", dir)
		if err != nil || stderr != "" {
			tr.Failf(testAuditName, "Failed to get file size statistics: %v\n%v", err, stderr)
		} else {
			tr.Passf(testAuditName, "api-server pod log size is: %v", stdout)
		}
	}
}

// Test k8s encryption provider functionality.
func TestEncryption(p *platform.Platform, tr *console.TestResults) {
	tc := p.Kubernetes.EncryptionConfig.EncryptionProviderConfigFile

	if tc == "" {
		tr.Skipf(testEncryptionName, "No encryption provider configuration specified.")
		return
	}

	_, err := p.GetClientset()
	if err != nil {
		tr.Failf(testEncryptionName, "Failed to get k8s client: %v", err)
		// We're done, we can't test anything further.
		return
	}

	pod, err := p.Client.GetFirstPodByLabelSelector("kube-system", "component=kube-apiserver")
	if err != nil {
		tr.Failf(testEncryptionName, "Failed to get api-server pod: %v", err)
		return
	}
	tr.Passf(testEncryptionName, "api-server pod found")

	if logFilePath, ok := p.Kubernetes.APIServerExtraArgs["encryption-provider-config"]; !ok {
		tr.Failf(testEncryptionName, "No encryption-provider-config is specified!")
		return
	} else if logFilePath == "" {
		tr.Failf(testEncryptionName, "Empty encryption-provider-config is specified!")
		return
	} else {
		//TODO: test encryption is working
	}
}
