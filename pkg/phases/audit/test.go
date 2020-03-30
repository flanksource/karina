package audit

import (
	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/platform"
	v1 "k8s.io/api/core/v1"
	"path/filepath"
	"strconv"
	"strings"
)

const jUnitAuditConfigClass = "AuditConfig"
const jUnitAuditAPIServerStateClass = "AuditApiServerState"
const jUnitAuditLogs = "AuditLogsPresent"

//Test k8s auditing functionality.
func Test(p *platform.Platform, tr *console.TestResults) {
	ac := p.AuditConfig

	if ac == nil {
		// A nil AuditConfig is a failure as the implicit default if it is not specified
		// should be 'disabled'
		tr.Failf(jUnitAuditConfigClass, "There is no AuditConfig")
		tr.Done()
	} else if ac.Disabled {
		tr.Skipf(jUnitAuditConfigClass, "AuditConfig is disabled")
		tr.Done()
	}

	_, err := p.GetClientset()
	if err != nil {
		tr.Failf(jUnitAuditConfigClass, "Failed to get k8s client: %v", err)
		// We're done, we can't test anything further.
		tr.Done()
	}

	pod, err := p.Client.GetFirstPodByLabelSelector("kube-system", "component=kube-apiserver")
	if err != nil {
		tr.Failf(jUnitAuditApiServerStateClass, "Failed to get api-server pod: %v", err)
		tr.Done()
	}
	tr.Passf(jUnitAuditApiServerStateClass, "api-server pod found")

	logFilePath := p.AuditConfig.APIServerOptions.LogOptions.Path
	//NOTE: '-' - means log to stdout, i.e. api-server logs
	if logFilePath == "-" {
		tr.Skipf(jUnitAuditConfigClass, "api-server is configured lo log to stdout, not verifying output")
	} else if logFilePath != ""  {
		dir := filepath.Dir(ac.APIServerOptions.LogOptions.Path)
		stdout, stderr, err := p.ExecutePodf("kube-system", pod.Name, "kube-apiserver", "/usr/bin/du", "-s", dir)
		if err != nil || stderr != "" {
			tr.Failf(jUnitAuditLogs, "Failed to get file size statistics: %v\n%v", err, stderr)
		} else {
			tr.Passf(jUnitAuditLogs, "api-server pod log size is: %v", stdout)
		}
	}

	argMap := createArgMap(pod.Spec.Containers, tr)

	if ac.PolicyFile == "" {
		tr.Failf(jUnitAuditConfigClass, "--audit-policy-file not configured"  )
	}


	//assignment helper to supply default
	wantedJSON := func(s string) string{
		if (s == "") {
			return "json"
		}
		return s
	}

	var parameterTests = []struct {
		description     string
		testParameter	string
		wantValue		string
	}{
		{
			description:	"Audit policy file set correctly",
			testParameter: "--audit-policy-file",
			wantValue: "/etc/kubernetes/policies/" + filepath.Base(ac.PolicyFile),
		},
		{
			description:   "Audit log format set correctly",
			testParameter: "--audit-log-format",
			wantValue:     wantedJSON(ac.APIServerOptions.LogOptions.Format),
		},
		{
			description:	"Audit log file maximum age set correctly",
			testParameter: "--audit-log-maxage",
			wantValue: strconv.Itoa(ac.APIServerOptions.LogOptions.MaxAge),
		},
		{
			description:	"Audit log file maximum backups set correctly",
			testParameter: "--audit-log-maxbackup",
			wantValue: strconv.Itoa(ac.APIServerOptions.LogOptions.MaxBackups),
		},
		{
			description:	"Audit log file maximum size set correctly",
			testParameter: "--audit-log-maxsize",
			wantValue: strconv.Itoa(ac.APIServerOptions.LogOptions.MaxSize),
		},
	}

	for _, t := range parameterTests {
		if testArgValue(t.wantValue, t.testParameter, argMap ) {
			tr.Passf(jUnitAuditConfigClass, t.description + ": "+ t.testParameter + " configured to %v", t.wantValue)
		} else {
			tr.Failf(jUnitAuditConfigClass, t.description + ": "+ t.testParameter + " configured incorrectly want %v, got %v", t.wantValue, argMap[t.testParameter]  )
		}
	}
}

// create a map of api-server startup parameters for easier comparisons
func createArgMap(containers []v1.Container, tr *console.TestResults ) (argMap map[string]string ) {
	argMap = map[string]string{}
	for _, container := range containers {
		if container.Name == "kube-apiserver" {
			if container.Command == nil {
				tr.Failf(jUnitAuditConfigClass, "api-server pod kube-apiserver container doesn't have command/args")
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

// test if a given argument is set to the desired value
func testArgValue(wantValue string, argChecked string, argMap map[string]string  ) bool {
	if wantValue != "" {
		if argMap[argChecked] != wantValue {
			return false
		}
		return true
	}
	return false
}
