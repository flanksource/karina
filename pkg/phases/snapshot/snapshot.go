package snapshot

import (
	"fmt"
	"io"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/moshloop/platform-cli/pkg/platform"
)

func Take(p *platform.Platform, dst string, since time.Duration) error {

	sinceTime := metav1.NewTime(time.Now().Add(-since))
	k8s, err := p.GetClientset()
	if err != nil {
		return err
	}

	namespaces, err := k8s.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, namespace := range namespaces.Items {
		pods := k8s.CoreV1().Pods(namespace.Name)
		list, err := pods.List(metav1.ListOptions{})
		if err != nil {
			return err
		}

		for _, pod := range list.Items {
			for _, container := range pod.Spec.Containers {
				path := dst + "/" + namespace.Name
				logs := pods.GetLogs(pod.Name, &v1.PodLogOptions{
					Container: container.Name,
					SinceTime: &sinceTime,
				})
				podLogs, err := logs.Stream()
				if err != nil {
					log.Errorf("Failed to stream logs %v", err)
					continue
				}
				defer podLogs.Close()
				os.MkdirAll(path, 0755)
				logFile := fmt.Sprintf("%s/%s-%s.log", path, pod.Name, container.Name)
				file, err := os.Create(logFile)
				if err != nil {
					log.Errorf("Failed to create file: %v", err)
					continue
				}
				count, err := io.Copy(file, podLogs)
				if err != nil {
					log.Errorf("Error saving logs for %s: %v", dst, err)
					continue
				} else if count == 0 {
					os.Remove(logFile)
				} else {
					log.Infof("Saving logs for %s/%s-%s (%dkb)", namespace.Name, pod.Name, container.Name, count/1024)
				}
			}

		}
	}
	return nil

}
