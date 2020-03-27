package audit

import (
	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/platform"
	v1 "k8s.io/api/core/v1"
	"path/filepath"
	"strconv"
	"strings"
)

const JUnitAuditConfigClass = "AuditConfig"
const jUnitAuditApiServerStateClass = "AuditApiServerState"
const jUnitAuditLogs = "AuditLogsPresent"

func Test(p *platform.Platform, tr *console.TestResults) {
	ac := p.AuditConfig

	if ac == nil {
		// nil is a failure as the implicit default if it is not specified
		// should be 'disabled'
		tr.Failf(JUnitAuditConfigClass, "There is no AuditConfig")
		tr.Done()
	}

	if ac.Disabled {
		tr.Skipf(JUnitAuditConfigClass, "AuditConfig is disabled")
		tr.Done()
	}

	_, err := p.GetClientset()
	if err != nil {
		tr.Failf(JUnitAuditConfigClass, "Failed to get k8s client: %v", err)
		tr.Done()
	}

	pod, err := p.Client.GetFirstPodByLabelSelector("kube-system", "component=kube-apiserver")
	if err != nil {
		tr.Failf(jUnitAuditApiServerStateClass, "Failed to get api-server pod: %v", err)
		tr.Done()
	}
	tr.Passf(jUnitAuditApiServerStateClass, "api-server pod found")


	if logFilePath := p.AuditConfig.ApiServerOptions.LogOptions.Path;
		logFilePath != "" && logFilePath != "-" {


		dir := filepath.Dir(ac.ApiServerOptions.LogOptions.Path)
		stdout, stderr, err := p.ExecutePodf("kube-system", pod.Name, "kube-apiserver", "/usr/bin/du", "-s", dir)
		if err != nil || stderr != "" {
			tr.Failf(jUnitAuditLogs, "Failed to get file size statistics: %v\n%v", err, stderr)
			tr.Done()
		}

		tr.Passf(jUnitAuditLogs, "api-server pod log size is: %v", stdout)
	}


	argMap := createArgMap(pod.Spec.Containers, tr)


	if ac.PolicyFile == "" {
		tr.Failf(JUnitAuditConfigClass, "--audit-policy-file not configured"  )
	}


	testArgValue(&ac.PolicyFile, "--audit-policy-file", argMap, tr )
	testArgValue(&ac.ApiServerOptions.LogOptions.Format, "--audit-log-format", argMap, tr )

	maxAge := strconv.Itoa(ac.ApiServerOptions.LogOptions.MaxAge)
	testArgValue(&maxAge, "--audit-log-maxage", argMap, tr )

	maxBackups := strconv.Itoa(ac.ApiServerOptions.LogOptions.MaxBackups)
	testArgValue(&maxBackups, "--audit-log-maxbackup", argMap, tr )

	maxSize := strconv.Itoa(ac.ApiServerOptions.LogOptions.MaxSize)
	testArgValue(&maxSize, "--audit-log-maxsize", argMap, tr )



}

func createArgMap(containers []v1.Container, tr *console.TestResults ) (argMap map[string]string ) {
	argMap = map[string]string{}
	for _, container := range containers {
		if container.Name == "kube-apiserver" {
			if container.Command == nil {
				tr.Failf(JUnitAuditConfigClass, "api-server pod kube-apiserver container doesn't have command/args")
				tr.Done()
			}
			for i, cmd := range container.Command {
				if (i!=0) {
					parts := strings.Split(cmd,"=")
					if (len(parts) < 2) {
						//ignore this
						break;
					}
					argMap[parts[0]] = parts[1]

				}
			}
		}
	}
	return
}

func testArgValue(wantValue* string, argChecked string, argMap map[string]string , tr *console.TestResults ) {
	if wantValue != nil && *wantValue != "" {
		if argMap[argChecked] != *wantValue {
			tr.Failf(JUnitAuditConfigClass, argChecked + " configured incorrectly want %v, got %v", *wantValue, argMap[argChecked]  )
		} else {
			tr.Passf(JUnitAuditConfigClass, argChecked + " configured to %v", *wantValue)
		}
	}
}
