package nsx

import (
	"fmt"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/moshloop/platform-cli/pkg/platform"
)

var LogLevels = []string{
	"WARNING",
	"INFO",
	"DEBUG",
	"ERROR",
	"CRITICAL",
}

func SetLogLevel(p *platform.Platform, level string) error {
	level = strings.ToUpper(level)
	if !strings.Contains(strings.Join(LogLevels, " "), level) {
		return fmt.Errorf("invalid log level: %s, valid levels are: %v", level, LogLevels)
	}
	client, err := p.GetClientset()
	if err != nil {
		return err
	}

	pods, err := client.CoreV1().Pods(Namespace).List(v1.ListOptions{})
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		component := pod.Labels["component"]
		switch component {
		case "nsx-ncp":
			if stdout, stderr, err := p.ExecutePodf(Namespace, pod.Name, "nsx-ncp", "nsxcli", "-c set ncp-log-level "+level); err != nil {
				p.Errorf("Failed to set logging level for %s: %v", pod.Name, err)
			} else {
				p.Infof("[%s] stdout: %s stderr: %s", pod.Name, stdout, stderr)
			}
		case "nsx-node-agent":
			if stdout, stderr, err := p.ExecutePodf(Namespace, pod.Name, "nsx-node-agent", "nsxcli", "-c set node-agent-log-level "+level); err != nil {
				p.Errorf("Failed to set logging level for %s: %v", pod.Name, err)
			} else {
				p.Infof("[%s/nsx-node-agent] stdout: %s stderr: %s", pod.Name, stdout, stderr)
			}
			if stdout, stderr, err := p.ExecutePodf(Namespace, pod.Name, "nsx-kube-proxy", "nsxcli", "-c set kube-proxy-log-level "+level); err != nil {
				p.Errorf("Failed to set logging level for %s: %v", pod.Name, err)
			} else {
				p.Infof("[%s/nsx-kube-proxy] stdout: %s stderr: %s", pod.Name, stdout, stderr)
			}
		}
	}
	return nil
}
