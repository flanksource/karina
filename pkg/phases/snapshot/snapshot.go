package snapshot

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/moshloop/platform-cli/pkg/platform"
)

type SnapshotOptions struct {
	IncludeSpecs  bool
	IncludeLogs   bool
	IncludeEvents bool
	Namespaces    []string
	Destination   string
	LogsSince     time.Duration
}

func Take(p *platform.Platform, opts SnapshotOptions) error {
	var namespaceList *v1.NamespaceList

	k8s, err := p.GetClientset()
	if err != nil {
		return fmt.Errorf("take: failed to get clientset: %v", err)
	}
	if len(opts.Namespaces) == 0 {
		namespaceList, err = k8s.CoreV1().Namespaces().List(metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("take: failed to list namespaces: %v", err)
		}
	} else {
		var namespaces []v1.Namespace
		for _, ns := range opts.Namespaces {
			nsData, _ := k8s.CoreV1().Namespaces().Get(ns, metav1.GetOptions{})
			namespaces = append(namespaces, *nsData)
		}
		namespaceList = &v1.NamespaceList{Items: namespaces, TypeMeta: metav1.TypeMeta{}}
	}

	if err := extractLogs(k8s, namespaceList, opts); err != nil {
		return err
	}
	if err := extractSpecs(p, k8s, namespaceList, opts); err != nil {
		return err
	}

	if err := extractEvents(k8s, namespaceList, opts); err != nil {
		return err
	}

	return nil
}

func extractEvents(k8s *kubernetes.Clientset, namespaceList *v1.NamespaceList, opts SnapshotOptions) error {
	if !opts.IncludeEvents {
		return nil
	}
	for _, namespace := range namespaceList.Items {

		events := k8s.CoreV1().Events(namespace.Name)
		eventList, err := events.List(metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to get events for %s : %s ", namespace.Name, err)
		}
		path := fmt.Sprintf("%s/%s/events.csv", opts.Destination, namespace.Name)
		os.Remove(path)
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
		w := tabwriter.NewWriter(file, 0, 0, 2, ' ', 0)
		log.Infof("Saving %d events to: %s", len(eventList.Items), path)

		fmt.Fprint(w, "LAST SEEN\tCount\tTYPE\tREASON\tOBJECT\tMESSAGE\n")

		for _, event := range eventList.Items {
			fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\t%s\n", event.LastTimestamp, event.Count, event.Type, event.Reason, event.InvolvedObject.Kind+"/"+event.InvolvedObject.Name, event.Message)
		}
		w.Flush()
		if err := file.Sync(); err != nil {
			return err
		}
	}

	return nil
}

func extractLogs(k8s *kubernetes.Clientset, namespaceList *v1.NamespaceList, opts SnapshotOptions) error {
	if !opts.IncludeLogs {
		return nil
	}
	sinceTime := metav1.NewTime(time.Now().Add(-opts.LogsSince))
	for _, namespace := range namespaceList.Items {
		pods := k8s.CoreV1().Pods(namespace.Name)
		list, err := pods.List(metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("take: failed to list pods: %v", err)
		}

		for _, pod := range list.Items {
			for _, container := range append(pod.Spec.Containers, pod.Spec.InitContainers...) {
				path := fmt.Sprintf("%s/%s/logs", opts.Destination, namespace.Name)

				log.Infof("Looking for %s/%s/%s", namespace.Name, pod.Name, container.Name)

				var logs *rest.Request
				if opts.LogsSince.Seconds() > 0 {
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
					log.Errorf("Error saving logs for %s: %v", opts.Destination, err)
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

func extractSpecs(p *platform.Platform, k8s *kubernetes.Clientset, namespaceList *v1.NamespaceList, opts SnapshotOptions) error {
	if !opts.IncludeSpecs {
		return nil

	}
	for _, namespace := range namespaceList.Items {

		path := fmt.Sprintf("%s/%s", opts.Destination, namespace.Name)
		os.MkdirAll(path, 0755)
		ketall := p.GetBinaryWithKubeConfig("ketall")
		_ = ketall(fmt.Sprintf("-n %s -o yaml --exclude= > %s/specs.yaml", namespace.Name, path))
	}

	return nil
}
