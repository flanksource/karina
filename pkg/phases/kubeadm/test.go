package kubeadm

import (
	"fmt"
	"path/filepath"
	"strings"

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

	//pod, err := p.Client.GetFirstPodByLabelSelector("kube-system", "component=kube-apiserver")
	//if err != nil {
	//	tr.Failf(testEncryptionName, "Failed to get api-server pod: %v", err)
	//	return
	//}
	//tr.Passf(testEncryptionName, "api-server pod found")

	etcdPod, err := p.Client.GetFirstPodByLabelSelector("kube-system", "component=etcd")
	if err != nil {
		tr.Failf(testEncryptionName, "Failed to get api-server pod: %v", err)
		return
	}
	tr.Passf(testEncryptionName, "etcd pod found")

	//if logFilePath, ok := p.Kubernetes.APIServerExtraArgs["encryption-provider-config"]; !ok {
	//	tr.Failf(testEncryptionName, "No encryption-provider-config is specified!")
	//	return
	//} else if logFilePath == "" {
	//	tr.Failf(testEncryptionName, "Empty encryption-provider-config is specified!")
	//	return
	//} else {
		//TODO: test encryption is working
		tr.Infof("%v",etcdPod)
		ns := "default"
		secretName := "encryption-test-secret"
		secretKey := "test-secret"
		secretValue := "correct-horse-battery-staple"
		tr.Infof("Creating secret %v in %v",secretName, ns)
		p.Client.CreateOrUpdateSecret(secretName,ns,
			map[string]([]byte){
				secretKey: []byte(secretValue),
			})

		verificationCommand := fmt.Sprintf("ETCDCTL_API=3 etcdctl get /registry/secrets/%v/%v"+
			" --endpoints https://127.0.0.1:2379"+
			" --cacert /etc/kubernetes/pki/etcd/ca.crt" +
			" --cert /etc/kubernetes/pki/etcd/peer.crt" +
			" --key /etc/kubernetes/pki/etcd/peer.key" +
			" | strings -n 6 -", ns,secretName)
		stdout, stderr, err := p.ExecutePodf("kube-system", etcdPod.Name, "etcd",
			"/bin/sh", "-c",
			// one long '-quoted command passed to /bin/sh
			"'"+verificationCommand+"'")
		// e.g.:
		// kubectl exec -n kube-system etcd-kind-control-plane -- /bin/sh -c 'ETCDCTL_API=3 etcdctl get /registry/secrets/default/secret1 --endpoints https://127.0.0.1:2379 --cacert /etc/kubernetes/pki/etcd/ca.crt --cert /etc/kubernetes/pki/etcd/peer.crt --key /etc/kubernetes/pki/etcd/peer.key | strings -n 6 -'
		///registry/secrets/default/secret1
		//k8s:enc:aescbc:v1:demokey:9
		if err != nil || stderr != "" {
			tr.Failf(testEncryptionName, "Failed to verify secret: %v\n%v", err, stderr)
		} else if strings.Contains(stdout,"k8s:enc:aescbc:v1" ) {
			tr.Passf(testEncryptionName, " %v", stdout)
			//TODO: read secret to be sure
		}
	//}
}
