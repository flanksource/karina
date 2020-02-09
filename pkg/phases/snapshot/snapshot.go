package snapshot

import (
	"fmt"
	"io"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"github.com/moshloop/platform-cli/pkg/platform"
)

func Take(p *platform.Platform, dst string, since time.Duration, nsArgs []string, includeSpecs bool) error {
	var namespaceList *v1.NamespaceList
	sinceTime := metav1.NewTime(time.Now().Add(-since))
	k8s, err := p.GetClientset()
	if err != nil {
		return fmt.Errorf("take: failed to get clientset: %v", err)
	}
	if len(nsArgs) == 0 {
		namespaceList, err = k8s.CoreV1().Namespaces().List(metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("take: failed to list namespaces: %v", err)
		}
	} else {
		var namespaces []v1.Namespace
		for _, ns := range nsArgs {
			nsData, _ := k8s.CoreV1().Namespaces().Get(ns, metav1.GetOptions{})
			namespaces = append(namespaces, *nsData)
		}
		namespaceList = &v1.NamespaceList{Items: namespaces, TypeMeta: metav1.TypeMeta{}}
	}


	for _, namespace := range namespaceList.Items {
		pods := k8s.CoreV1().Pods(namespace.Name)
		list, err := pods.List(metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("take: failed to list pods: %v", err)
		}

		for _, pod := range list.Items {
			for _, container := range append(pod.Spec.Containers, pod.Spec.InitContainers...) {
				path := fmt.Sprintf("%s/%s/logs", dst, namespace.Name)

				log.Infof("Looking for %s/%s/%s", namespace.Name, pod.Name, container.Name)

				var logs *rest.Request
				if since.Seconds() > 0 {
					logs = pods.GetLogs(pod.Name, &v1.PodLogOptions{
						Container: container.Name,
						SinceTime: &sinceTime,
					})
				} else {
					logs = pods.GetLogs(pod.Name, &v1.PodLogOptions{
						Container: container.Name,
					})
				}

				podLogs, err := logs.Stream()
				if err != nil {
					log.Errorf("Failed to stream logs %v", err)
					continue
				}
				defer podLogs.Close()
				err = os.MkdirAll(path, 0755)
				if err != nil {
					log.Errorf("Failed to create directory: %v", err)
				}
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
		if includeSpecs {
			path := fmt.Sprintf("%s/%s", dst, namespace.Name)
			err = os.MkdirAll(path, 0755)
			ketall := p.GetBinaryWithKubeConfig("ketall")
			_ = ketall(fmt.Sprintf("-n %s -o yaml --exclude= > %s/specs.yaml", namespace.Name, path))
		}
	}
	return err
}
