package snapshot

import (
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/moshloop/platform-cli/pkg/platform"
)

type Options struct {
	IncludeSpecs  bool
	IncludeLogs   bool
	IncludeEvents bool
	Namespaces    []string
	Destination   string
	LogsSince     time.Duration
	Concurrency   int
}

// nolint: golint
type SnapshotFetcher struct {
	*platform.Platform
	k8s        *kubernetes.Clientset
	namespaces *v1.NamespaceList
	wg         *sync.WaitGroup
	ch         chan int
	opts       Options
}

func Take(p *platform.Platform, opts Options) error {
	var namespaceList *v1.NamespaceList

	k8s, err := p.GetClientset()
	if err != nil {
		return errors.Wrap(err, "failed to get clientset")
	}
	if len(opts.Namespaces) == 0 {
		namespaceList, err = k8s.CoreV1().Namespaces().List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to list namespaces")
		}
	} else {
		var namespaces []v1.Namespace
		for _, ns := range opts.Namespaces {
			nsData, _ := k8s.CoreV1().Namespaces().Get(ns, metav1.GetOptions{})
			namespaces = append(namespaces, *nsData)
		}
		namespaceList = &v1.NamespaceList{Items: namespaces, TypeMeta: metav1.TypeMeta{}}
	}

	sf := &SnapshotFetcher{
		Platform:   p,
		k8s:        k8s,
		namespaces: namespaceList,
		wg:         &sync.WaitGroup{},
		ch:         make(chan int, opts.Concurrency),
		opts:       opts,
	}

	sf.Fetch()

	return nil
}

func (s *SnapshotFetcher) Fetch() {
	for _, namespace := range s.namespaces.Items {
		if s.opts.IncludeEvents {
			s.queueFetchEvents(namespace)
		}
		if s.opts.IncludeSpecs {
			s.queueFetchSpecs(namespace)
		}
		if s.opts.IncludeLogs {
			sinceTime := metav1.NewTime(time.Now().Add(-s.opts.LogsSince))
			s.queueFetchLogs(namespace, sinceTime)
		}
	}

	s.wg.Wait()
}

func (s *SnapshotFetcher) queueFetchLogs(namespace v1.Namespace, sinceTime metav1.Time) {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		s.ch <- 1

		pods := s.k8s.CoreV1().Pods(namespace.Name)
		list, err := pods.List(metav1.ListOptions{})
		if err != nil {
			s.Errorf("failed to list pods for namespace %s: %v", namespace.Name, err)
			<-s.ch
			return
		}

		for _, pod := range list.Items {
			for _, container := range append(pod.Spec.Containers, pod.Spec.InitContainers...) {
				s.queueFetchPodLogs(namespace, pod, container, sinceTime)
			}
		}

		<-s.ch
	}()
}

func (s *SnapshotFetcher) queueFetchPodLogs(namespace v1.Namespace, pod v1.Pod, container v1.Container, sinceTime metav1.Time) {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		s.ch <- 1
		path := fmt.Sprintf("%s/%s/logs", s.opts.Destination, namespace.Name)

		s.Infof("Looking for %s/%s/%s", namespace.Name, pod.Name, container.Name)

		pods := s.k8s.CoreV1().Pods(namespace.Name)

		var logs *rest.Request
		if s.opts.LogsSince.Seconds() > 0 {
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
			s.Errorf("Failed to stream logs %v", err)
			<-s.ch
			return
		}
		defer podLogs.Close()
		err = os.MkdirAll(path, 0755)
		if err != nil {
			s.Errorf("Failed to create directory: %v", err)
			<-s.ch
			return
		}
		logFile := fmt.Sprintf("%s/%s-%s.log", path, pod.Name, container.Name)
		file, err := os.Create(logFile)
		if err != nil {
			s.Errorf("Failed to create file: %v", err)
			<-s.ch
			return
		}
		count, err := io.Copy(file, podLogs)
		if err != nil {
			s.Errorf("Error saving logs for %s: %v", s.opts.Destination, err)
		} else if count == 0 {
			os.Remove(logFile)
		} else {
			log.Infof("Saving logs for %s/%s-%s (%dkb)", namespace.Name, pod.Name, container.Name, count/1024)
		}
		<-s.ch
	}()
}

func (s *SnapshotFetcher) queueFetchSpecs(namespace v1.Namespace) {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		s.ch <- 1

		path := fmt.Sprintf("%s/%s", s.opts.Destination, namespace.Name)
		if err := os.MkdirAll(path, 0755); err != nil {
			s.Errorf("failed to mkdir path %s: %v", path, err)
			<-s.ch
			return
		}
		ketall := s.GetBinaryWithKubeConfig("ketall")
		if err := ketall(fmt.Sprintf("-n %s -o yaml --exclude=events,endpoints,secrets > %s/specs.yaml", namespace.Name, path)); err != nil {
			s.Errorf("failed to ketall specs for namespace %s: %v", namespace.Name, err)
		}
		<-s.ch
	}()
}

func (s *SnapshotFetcher) queueFetchEvents(namespace v1.Namespace) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.ch <- 1

		events := s.k8s.CoreV1().Events(namespace.Name)
		eventList, err := events.List(metav1.ListOptions{})
		if err != nil {
			s.Errorf("failed to get events for %s: %v", namespace.Name, err)
			<-s.ch
			return
		}

		directory := path.Join(s.opts.Destination, namespace.Name)
		err = os.MkdirAll(directory, 0755)
		if err != nil {
			s.Errorf("Failed to create directory: %v", err)
			<-s.ch
			return
		}

		eventsPath := path.Join(directory, "events.csv")
		os.Remove(eventsPath)

		file, err := os.Create(eventsPath)
		if err != nil {
			s.Errorf("failed to create path %s: %v", eventsPath, err)
			<-s.ch
			return
		}
		defer file.Close()

		w := tabwriter.NewWriter(file, 0, 0, 2, ' ', 0)
		s.Infof("Saving %d events to: %s", len(eventList.Items), eventsPath)
		fmt.Fprint(w, "LAST SEEN\tCount\tTYPE\tREASON\tOBJECT\tMESSAGE\n")
		for _, event := range eventList.Items {
			fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\t%s\n", event.LastTimestamp, event.Count, event.Type, event.Reason, event.InvolvedObject.Kind+"/"+event.InvolvedObject.Name, event.Message)
		}
		w.Flush()

		if err := file.Sync(); err != nil {
			s.Errorf("failed to sync file %s: %v", eventsPath, err)
		}
		<-s.ch
	}()
}
