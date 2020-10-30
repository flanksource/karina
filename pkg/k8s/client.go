package k8s

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"net"
	"net/http"
	"reflect"

	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/AlekSi/pointer"
	certs "github.com/flanksource/commons/certs"
	"github.com/flanksource/commons/files"
	"github.com/flanksource/commons/logger"
	utils "github.com/flanksource/commons/utils"
	"github.com/flanksource/karina/pkg/k8s/drain"
	"github.com/flanksource/karina/pkg/k8s/etcd"
	"github.com/flanksource/karina/pkg/k8s/kustomize"
	"github.com/flanksource/karina/pkg/k8s/proxy"
	"github.com/flanksource/karina/pkg/types"
	"github.com/go-test/deep"
	"github.com/mitchellh/mapstructure"
	perrors "github.com/pkg/errors"
	"gopkg.in/flanksource/yaml.v3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/protobuf"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	cliresource "k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/transport"
)

type Client struct {
	logger.Logger
	GetKubeConfigBytes  func() ([]byte, error)
	ApplyDryRun         bool
	ApplyHook           ApplyHook
	Trace               bool
	GetKustomizePatches func() ([]string, error)
	client              *kubernetes.Clientset
	dynamicClient       dynamic.Interface
	restConfig          *rest.Config
	etcdClientGenerator *etcd.EtcdClientGenerator
	kustomizeManager    *kustomize.Manager
	restMapper          meta.RESTMapper
}

func (c *Client) ResetConnection() {
	c.client = nil
	c.dynamicClient = nil
	c.restConfig = nil
	c.etcdClientGenerator = nil
}

func (c *Client) GetEtcdClientGenerator(ca *certs.Certificate) (*etcd.EtcdClientGenerator, error) {
	if c.etcdClientGenerator != nil {
		return c.etcdClientGenerator, nil
	}
	client, err := c.GetClientset()
	if err != nil {
		return nil, err
	}
	rest, _ := c.GetRESTConfig()
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(ca.EncodedCertificate())
	cert, _ := tls.X509KeyPair(ca.EncodedCertificate(), ca.EncodedPrivateKey())
	return etcd.NewEtcdClientGenerator(client, rest, &tls.Config{
		RootCAs:      caPool,
		Certificates: []tls.Certificate{cert},
	}), nil
}

func (c *Client) getDrainHelper() (*drain.Helper, error) {
	client, err := c.GetClientset()
	if err != nil {
		return nil, err
	}
	return &drain.Helper{
		Ctx:                 context.Background(),
		ErrOut:              os.Stderr,
		Out:                 os.Stdout,
		Client:              client,
		DeleteLocalData:     true,
		IgnoreAllDaemonSets: true,
		Timeout:             120 * time.Second,
	}, nil
}

func (c *Client) ScalePod(pod v1.Pod, replicas int32) error {
	client, err := c.GetClientset()
	if err != nil {
		return err
	}

	for _, owner := range pod.GetOwnerReferences() {
		if owner.Kind == "ReplicaSet" {
			replicasets := client.AppsV1().ReplicaSets(pod.Namespace)
			rs, err := replicasets.Get(owner.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if *rs.Spec.Replicas != replicas {
				c.Infof("Scaling %s/%s => %d", pod.Namespace, owner.Name, replicas)
				rs.Spec.Replicas = &replicas
				_, err := replicasets.Update(rs)
				if err != nil {
					return err
				}
			} else {
				c.Infof("Scaling %s/%s => %d (no-op)", pod.Namespace, owner.Name, replicas)
			}
		}
	}
	return nil
}

func (c *Client) GetPodReplicas(pod v1.Pod) (int, error) {
	client, err := c.GetClientset()
	if err != nil {
		return 0, err
	}

	for _, owner := range pod.GetOwnerReferences() {
		if owner.Kind == "ReplicaSet" {
			replicasets := client.AppsV1().ReplicaSets(pod.Namespace)
			rs, err := replicasets.Get(owner.Name, metav1.GetOptions{})
			if err != nil {
				return 0, err
			}
			return int(*rs.Spec.Replicas), nil
		}
	}
	return 1, nil
}

func (c *Client) Drain(nodeName string, timeout time.Duration) error {
	c.Infof("[%s] draining", nodeName)
	if err := c.Cordon(nodeName); err != nil {
		return fmt.Errorf("error cordoning %s: %v", nodeName, err)
	}

	if err := backoff(func() error {
		return c.EvictNode(nodeName)
	}, c.Logger, nil); err != nil {
		return fmt.Errorf("failed to evict node %s: %v", nodeName, err)
	}

	return nil
}

func (c *Client) EvictPod(pod v1.Pod) error {
	if IsPodDaemonSet(pod) || IsPodFinished(pod) || IsDeleted(&pod) || IsStaticPod(pod) {
		return nil
	}
	client, err := c.GetClientset()
	if err != nil {
		return err
	}
	drainer, err := c.getDrainHelper()
	if err != nil {
		return err
	}
	replicas, err := c.GetPodReplicas(pod)
	if err != nil {
		return err
	}
	if replicas == 1 {
		if err := c.ScalePod(pod, int32(2)); err != nil {
			return err
		}
		defer func() {
			if err := c.ScalePod(pod, int32(1)); err != nil {
				c.Warnf("Failed to scale back pod: %v", err)
			}
		}()
	}

	if pod.ObjectMeta.Labels["spilo-role"] == "master" {
		c.Infof("Conducting failover of %s", pod.Name)
		var stdout, stderr string
		if stdout, stderr, err = c.ExecutePodf(pod.Namespace, pod.Name, "postgres", "curl", "-s", "http://localhost:8008/switchover", "-XPOST", fmt.Sprintf("-d {\"leader\":\"%s\"}", pod.Name)); err != nil {
			return fmt.Errorf("failed to failover instance, aborting: %v %s %s", err, stderr, stdout)
		}
		c.Infof("Failed over: %s %s", stdout, stderr)
	}
	if err := drainer.DeleteOrEvictPods(pod); err != nil {
		return err
	}

	pvcs := client.CoreV1().PersistentVolumeClaims(pod.Namespace)
	for _, vol := range pod.Spec.Volumes {
		if vol.PersistentVolumeClaim != nil {
			pvc, err := pvcs.Get(vol.PersistentVolumeClaim.ClaimName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if pvc != nil && pvc.Spec.StorageClassName == nil || strings.Contains(*pvc.Spec.StorageClassName, "local") {
				c.Infof("[%s] deleting", pvc.Name)
				if err := pvcs.Delete(pvc.Name, &metav1.DeleteOptions{}); err != nil {
					return err
				}
				//nolint: errcheck
				wait.PollImmediate(1*time.Second, 2*time.Minute, func() (bool, error) {
					_, err := pvcs.Get(pvc.Name, metav1.GetOptions{})
					return errors.IsNotFound(err), nil
				})
				pvc.ObjectMeta.SetAnnotations(nil)
				pvc.SetFinalizers([]string{})
				pvc.SetSelfLink("")
				pvc.SetResourceVersion("")
				pvc.Spec.VolumeName = ""
				new, err := pvcs.Create(pvc)
				if err != nil {
					return err
				}
				c.Infof("Created new PVC %s -> %s", pvc.UID, new.UID)
			}
		}
	}
	return nil
}

func (c *Client) EvictNode(nodeName string) error {
	client, err := c.GetClientset()
	if err != nil {
		return nil
	}

	pods, err := client.CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{
		FieldSelector: fields.SelectorFromSet(fields.Set{"spec.nodeName": nodeName}).String(),
	})

	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		if err := c.EvictPod(pod); err != nil {
			return err
		}
	}

	volumeAttachments, err := client.StorageV1().VolumeAttachments().List(metav1.ListOptions{
		FieldSelector: fields.SelectorFromSet(fields.Set{"spec.nodeName": nodeName}).String(),
	})

	if err != nil {
		return err
	}

	for _, va := range volumeAttachments.Items {
		if err := c.RemoveVolumeAttachment(va); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) Cordon(nodeName string) error {
	c.Infof("[%s] cordoning", nodeName)

	client, err := c.GetClientset()
	if err != nil {
		return nil
	}
	nodes := client.CoreV1().Nodes()
	node, err := nodes.Get(nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	node.Spec.Unschedulable = true
	_, err = nodes.Update(node)
	return err
}

func (c *Client) Uncordon(nodeName string) error {
	c.Infof("[%s] uncordoning", nodeName)
	client, err := c.GetClientset()
	if err != nil {
		return nil
	}
	nodes := client.CoreV1().Nodes()
	node, err := nodes.Get(nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	node.Spec.Unschedulable = false
	_, err = nodes.Update(node)
	return err
}

func (c *Client) GetKustomize() (*kustomize.Manager, error) {
	if c.kustomizeManager != nil {
		return c.kustomizeManager, nil
	}
	dir, _ := ioutil.TempDir("", "platform-cli-kustomize")
	patches, err := c.GetKustomizePatches()
	if err != nil {
		return nil, err
	}
	no := 1
	var (
		patchData *[]byte
		name      string
	)
	for _, patch := range patches {
		if files.Exists(patch) {
			name = filepath.Base(patch)
			patchBytes, err := ioutil.ReadFile(patch)
			if err != nil {
				return nil, err
			}
			patchData = &patchBytes
		} else {
			patchBytes := []byte(patch)
			patchData = &patchBytes
			name = fmt.Sprintf("patch-%d.yaml", no)
			no++
		}
		patchData, err = templatizePatch(patchData)
		if err != nil {
			return nil, err
		}
		if c.Trace {
			c.Tracef("patch file %v after templating:\n%v\n\n", name, string(*patchData))
		}
		if _, err := files.CopyFromReader(bytes.NewBuffer(*patchData), dir+"/"+name, 0644); err != nil {
			return nil, err
		}
	}
	kustomizeManager, err := kustomize.GetManager(dir)
	c.kustomizeManager = kustomizeManager
	return c.kustomizeManager, err
}

// templatizePatch takes a patch stream (possibly containing multiple
// YAML documents) and templatizes each.
// blank documents are skipped.
func templatizePatch(patch *[]byte) (*[]byte, error) {
	var result []byte
	remainingData := patch
	for {
		first, rest := getDocumentsFromYamlFile(*remainingData)
		remainingData = &rest
		if len(first) == 0 {
			continue
		}
		templated, err := templatizeDocument(first)
		if err != nil {
			return nil, err
		}
		if len(result) > 0 {
			result = append(result, []byte("---\n")...)
		}
		result = append(result, *templated...)
		if len(rest) == 0 {
			break
		}
	}
	return &result, nil
}

// templatizeDocument applies templating to a supplied YAML
// document via the templating functionality in
// "gopkg.in/flanksource/yaml.v3"
// NOTE: only the first YAML document in a stream will be processed.
func templatizeDocument(patch []byte) (*[]byte, error) {
	var body interface{}
	if err := yaml.Unmarshal(patch, &body); err != nil {
		return nil, err
	}
	if body == nil {
		return &[]byte{}, nil
	}
	templated, err := yaml.Marshal(body)
	if err != nil {
		return nil, err
	}
	return &templated, nil
}

// getDocumentsFromYamlFile returns the first YAML document
// from a stream and a byte slice containing the remainder of the stream.
// This is needed since yaml.v3 (and the flanksource derived yaml.v3) only
// unmarshalls the **first** document in a stream.
//
// (see https://pkg.go.dev/gopkg.in/flanksource/yaml.v3@v3.1.1?tab=doc#Unmarshal)
func getDocumentsFromYamlFile(yamlData []byte) (firstDoc []byte, rest []byte) {
	endIndex := bytes.Index(yamlData, []byte("---"))
	if endIndex == -1 {
		return yamlData, []byte{}
	}
	return yamlData[:endIndex], yamlData[endIndex+3:]
}

// GetDynamicClient creates a new k8s client
func (c *Client) GetDynamicClient() (dynamic.Interface, error) {
	if c.dynamicClient != nil {
		return c.dynamicClient, nil
	}
	cfg, err := c.GetRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("getClientset: failed to get REST config: %v", err)
	}
	c.dynamicClient, err = dynamic.NewForConfig(cfg)
	return c.dynamicClient, err
}

// GetClientset creates a new k8s client
func (c *Client) GetClientset() (*kubernetes.Clientset, error) {
	if c.client != nil {
		return c.client, nil
	}

	cfg, err := c.GetRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("getClientset: failed to get REST config: %v", err)
	}
	c.client, err = kubernetes.NewForConfig(cfg)
	return c.client, err
}

func (c *Client) GetRESTConfig() (*rest.Config, error) {
	if c.restConfig != nil {
		return c.restConfig, nil
	}
	data, err := c.GetKubeConfigBytes()

	if err != nil {
		return nil, fmt.Errorf("getRESTConfig: failed to get kubeconfig: %v", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("kubeConfig is empty")
	}

	c.restConfig, err = clientcmd.RESTConfigFromKubeConfig(data)
	return c.restConfig, err
}

// GetSecret returns the data of a secret or nil for any error
func (c *Client) GetSecret(namespace, name string) *map[string][]byte {
	k8s, err := c.GetClientset()
	if err != nil {
		c.Tracef("failed to get client %v", err)
		return nil
	}
	secret, err := k8s.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		c.Tracef("failed to get secret %s/%s: %v\n", namespace, name, err)
		return nil
	}
	return &secret.Data
}

// GetConfigMap returns the data of a secret or nil for any error
func (c *Client) GetConfigMap(namespace, name string) *map[string]string {
	k8s, err := c.GetClientset()
	if err != nil {
		c.Tracef("failed to get client %v", err)
		return nil
	}
	cm, err := k8s.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		c.Tracef("failed to get secret %s/%s: %v\n", namespace, name, err)
		return nil
	}
	return &cm.Data
}

// Remove volume attachment
func (c *Client) RemoveVolumeAttachment(va storagev1.VolumeAttachment) error {
	k8s, err := c.GetClientset()
	if err != nil {
		return fmt.Errorf("failed to get clientset: %v", err)
	}

	volumeAPI := k8s.StorageV1().VolumeAttachments()

	if len(va.Finalizers) > 0 {
		va.Finalizers = []string{}
		if _, err := volumeAPI.Update(&va); err != nil {
			return fmt.Errorf("failed to remove finalizers from volume attachment %s: %v", va.Name, err)
		}
	}

	if err := volumeAPI.Delete(va.Name, nil); err != nil {
		return fmt.Errorf("failed to delete volume attachment %s: %v", va.Name, err)
	}

	return nil
}

func (c *Client) Get(namespace string, name string, obj runtime.Object) error {
	client, _, _, err := c.GetDynamicClientFor(namespace, obj)
	if err != nil {
		return err
	}
	unstructuredObj, err := client.Get(name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get: failed to get client: %v", err)
	}

	err = runtime.DefaultUnstructuredConverter.
		FromUnstructured(unstructuredObj.Object, obj)
	if err == nil {
		return nil
	}
	// if c.IsLevelEnabled(logger.TraceLevel) {
	// 	spew.Dump(unstructuredObj.Object)
	// }

	// FIXME(moshloop) getting the zalando operationconfiguration fails with "unrecognized type: int64" so we fall back to brute-force
	c.Warnf("Using mapstructure to decode %s: %v", obj.GetObjectKind().GroupVersionKind().Kind, err)
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		TagName:          "json",
		DecodeHook:       mapstructure.ComposeDecodeHookFunc(decodeStringToTime, decodeStringToDuration, decodeStringToTimeDuration),
		Result:           obj,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return fmt.Errorf("get: failed to decode config: %v", err)
	}
	return decoder.Decode(unstructuredObj.Object)
}

func decodeStringToTimeDuration(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}
	if t != reflect.TypeOf(time.Duration(5)) {
		return data, nil
	}
	d, err := time.ParseDuration(data.(string))
	if err != nil {
		return data, fmt.Errorf("decodeStringToTimeDuration: Failed to parse duration: %v", err)
	}
	return d, nil
}

func decodeStringToDuration(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}
	if t != reflect.TypeOf(metav1.Duration{Duration: time.Duration(5)}) {
		return data, nil
	}
	d, err := time.ParseDuration(data.(string))
	if err != nil {
		return data, fmt.Errorf("decodeStringToDuration: Failed to parse duration: %v", err)
	}
	return metav1.Duration{Duration: d}, nil
}

func decodeStringToTime(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}
	if t != reflect.TypeOf(metav1.Time{Time: time.Now()}) {
		return data, nil
	}
	d, err := time.Parse(time.RFC3339, data.(string))
	if err != nil {
		return data, fmt.Errorf("decodeStringToTime: failed to decode to time: %v", err)
	}
	return metav1.Time{Time: d}, nil
}

func (c *Client) GetRestMapper() (meta.RESTMapper, error) {
	if c.restMapper != nil {
		return c.restMapper, nil
	}

	config, _ := c.GetRESTConfig()

	// re-use kubectl cache
	host := config.Host
	host = strings.ReplaceAll(host, "https://", "")
	host = strings.ReplaceAll(host, "-", "_")
	host = strings.ReplaceAll(host, ":", "_")
	cacheDir := os.ExpandEnv("$HOME/.kube/cache/discovery/" + host)
	cache, err := disk.NewCachedDiscoveryClientForConfig(config, cacheDir, "", 10*time.Minute)
	if err != nil {
		return nil, err
	}
	c.restMapper = restmapper.NewDeferredDiscoveryRESTMapper(cache)
	return c.restMapper, err
}

func (c *Client) GetClientByKind(kind string) (dynamic.NamespaceableResourceInterface, error) {
	dynamicClient, err := c.GetDynamicClient()
	if err != nil {
		return nil, err
	}
	rm, _ := c.GetRestMapper()
	gvk, err := rm.KindFor(schema.GroupVersionResource{
		Resource: kind,
	})
	if err != nil {
		return nil, err
	}
	gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}
	mapping, err := rm.RESTMapping(gk, gvk.Version)
	if err != nil {
		return nil, err
	}
	return dynamicClient.Resource(mapping.Resource), nil
}

func (c *Client) GetDynamicClientFor(namespace string, obj runtime.Object) (dynamic.ResourceInterface, *schema.GroupVersionResource, *unstructured.Unstructured, error) {
	dynamicClient, err := c.GetDynamicClient()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("getDynamicClientFor: failed to get dynamic client: %v", err)
	}

	return c.getDynamicClientFor(dynamicClient, namespace, obj)
}

func (c *Client) GetDynamicClientForUser(namespace string, obj runtime.Object, user string) (dynamic.ResourceInterface, *schema.GroupVersionResource, *unstructured.Unstructured, error) {
	data, err := c.GetKubeConfigBytes()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("getRESTConfig: failed to get kubeconfig: %v", err)
	}
	if len(data) == 0 {
		return nil, nil, nil, fmt.Errorf("kubeConfig is empty")
	}

	cfg, err := clientcmd.RESTConfigFromKubeConfig(data)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("getClientset: failed to get REST config: %v", err)
	}

	impersonate := transport.ImpersonationConfig{UserName: user}

	transportConfig, err := cfg.TransportConfig()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get transport config: %v", err)
	}
	tlsConfig, err := transport.TLSConfigFor(transportConfig)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get tls config: %v", err)
	}
	timeout := 5 * time.Second

	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: 30 * time.Second,
			DualStack: false, // K8s do not work well with IPv6
		}).DialContext,
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: 10 * time.Second,
		MaxIdleConns:          10,
		MaxIdleConnsPerHost:   2,
		IdleConnTimeout:       20 * time.Second,
		TLSClientConfig:       tlsConfig,
	}

	cfg.Transport = transport.NewImpersonatingRoundTripper(impersonate, tr)
	cfg.TLSClientConfig = rest.TLSClientConfig{}
	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get dynamic from config: %v", err)
	}

	return c.getDynamicClientFor(dynamicClient, namespace, obj)
}

func (c *Client) getDynamicClientFor(dynamicClient dynamic.Interface, namespace string, obj runtime.Object) (dynamic.ResourceInterface, *schema.GroupVersionResource, *unstructured.Unstructured, error) {
	gvk := obj.GetObjectKind().GroupVersionKind()
	gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}
	rm, _ := c.GetRestMapper()

	mapping, err := rm.RESTMapping(gk, gvk.Version)
	if err != nil && meta.IsNoMatchError(err) && !c.ApplyDryRun {
		// new CRD may still becoming ready, flush caches and retry
		time.Sleep(5 * time.Second)
		c.restMapper = nil
		rm, _ := c.GetRestMapper()
		mapping, err = rm.RESTMapping(gk, gvk.Version)
	}
	if err != nil {
		return nil, nil, nil, err
	}

	resource := mapping.Resource

	convertedObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("getDynamicClientFor: failed to convert object: %v", err)
	}

	unstructuredObj := &unstructured.Unstructured{Object: convertedObj}

	if mapping.Scope == meta.RESTScopeRoot {
		return dynamicClient.Resource(mapping.Resource), &resource, unstructuredObj, nil
	}
	if namespace == "" {
		namespace = unstructuredObj.GetNamespace()
	}
	return dynamicClient.Resource(mapping.Resource).Namespace(namespace), &resource, unstructuredObj, nil
}

func (c *Client) GetRestClient(obj unstructured.Unstructured) (*cliresource.Helper, error) {
	rm, _ := c.GetRestMapper()
	restConfig, _ := c.GetRESTConfig()
	// Get some metadata needed to make the REST request.
	gvk := obj.GetObjectKind().GroupVersionKind()
	gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}
	mapping, err := rm.RESTMapping(gk, gvk.Version)
	if err != nil {
		return nil, err
	}

	gv := mapping.GroupVersionKind.GroupVersion()
	restConfig.ContentConfig = cliresource.UnstructuredPlusDefaultContentConfig()
	restConfig.GroupVersion = &gv
	if len(gv.Group) == 0 {
		restConfig.APIPath = "/api"
	} else {
		restConfig.APIPath = "/apis"
	}

	restClient, err := rest.RESTClientFor(restConfig)
	if err != nil {
		return nil, err
	}

	return cliresource.NewHelper(restClient, mapping), nil
}

func (c *Client) GetProxyDialer(p proxy.Proxy) (*proxy.Dialer, error) {
	clientset, err := c.GetClientset()
	if err != nil {
		return nil, err
	}

	restConfig, err := c.GetRESTConfig()
	if err != nil {
		return nil, err
	}

	return proxy.NewDialer(p, clientset, restConfig)
}

func (c *Client) ApplyUnstructured(namespace string, objects ...*unstructured.Unstructured) error {
	for _, unstructuredObj := range objects {
		client, err := c.GetRestClient(*unstructuredObj)
		if err != nil {
			return err
		}

		if c.ApplyHook != nil {
			c.ApplyHook(namespace, *unstructuredObj)
		}
		if c.ApplyDryRun {
			c.Debugf("[dry-run] %s/%s/%s created/configured", client.Resource, unstructuredObj, unstructuredObj.GetName())
		} else {
			_, err = client.Create(namespace, true, unstructuredObj, &metav1.CreateOptions{})
			if errors.IsAlreadyExists(err) {
				_, err = client.Replace(namespace, unstructuredObj.GetName(), true, unstructuredObj)
				if err != nil {
					c.Errorf("error handling: %s : %+v", client.Resource, err)
				} else {
					// TODO(moshloop): Diff the old and new objects and log unchanged instead of configured where necessary
					c.Infof("%s/%s/%s configured", client.Resource, unstructuredObj.GetNamespace(), unstructuredObj.GetName())
				}
			} else if err == nil {
				c.Infof("%s/%s/%s created", client.Resource, unstructuredObj.GetNamespace(), unstructuredObj.GetName())
			} else {
				c.Errorf("error handling: %s : %+v", client.Resource, err)
			}
		}
	}
	return nil
}

func (c *Client) trace(msg string, objects ...runtime.Object) {
	if !c.Trace {
		return
	}
	for _, obj := range objects {
		data, err := yaml.Marshal(obj)
		if err != nil {
			c.Errorf("Error tracing %s", err)
		} else {
			fmt.Printf("%s\n%s", msg, string(data))
		}
	}
}

func (c *Client) DeleteUnstructured(namespace string, objects ...*unstructured.Unstructured) error {
	for _, unstructuredObj := range objects {
		client, err := c.GetRestClient(*unstructuredObj)
		if err != nil {
			return err
		}

		if c.ApplyDryRun {
			c.Debugf("[dry-run] %s/%s/%s removed", namespace, client.Resource, unstructuredObj.GetName())
		} else {
			if _, err := client.Delete(namespace, unstructuredObj.GetName()); err != nil {
				return err
			}
			c.Infof("%s/%s/%s removed", namespace, client.Resource, unstructuredObj.GetName())
		}
	}
	return nil
}

type ApplyHook func(namespace string, obj unstructured.Unstructured)

func (c *Client) Apply(namespace string, objects ...runtime.Object) error {
	for _, obj := range objects {
		client, resource, unstructuredObj, err := c.GetDynamicClientFor(namespace, obj)
		if err != nil {
			if c.ApplyDryRun && strings.HasPrefix(err.Error(), "no matches for kind") {
				c.Debugf("[dry-run] failed to get dynamic client for namespace %s", namespace)
				continue
			}
			return fmt.Errorf("failed to get dynamic client for %v: %v", obj, err)
		}

		if c.ApplyHook != nil {
			c.ApplyHook(namespace, *unstructuredObj)
		}
		if c.ApplyDryRun {
			c.trace("apply", unstructuredObj)
			c.Debugf("[dry-run] %s/%s created/configured", resource.Resource, unstructuredObj.GetName())
			continue
		}

		existing, _ := client.Get(unstructuredObj.GetName(), metav1.GetOptions{})

		if existing == nil {
			c.trace("creating", unstructuredObj)
			_, err = client.Create(unstructuredObj, metav1.CreateOptions{})
			if err != nil {
				c.Errorf("error creating: %s/%s/%s : %+v", resource.Group, resource.Version, resource.Resource, err)
			} else {
				c.Infof("%s/%s/%s created", resource.Resource, unstructuredObj.GetNamespace(), unstructuredObj.GetName())
			}
		} else {
			if unstructuredObj.GetKind() == "Service" {
				// Workaround for immutable spec.clusterIP error message
				spec := unstructuredObj.Object["spec"].(map[string]interface{})
				spec["clusterIP"] = existing.Object["spec"].(map[string]interface{})["clusterIP"]
			} else if unstructuredObj.GetKind() == "ServiceAccount" {
				unstructuredObj.Object["secrets"] = existing.Object["secrets"]
			}
			// apps/DameonSet MatchExpressions:[]v1.LabelSelectorRequirement(nil)}: field is immutable
			// webhook CA's

			c.trace("updating", unstructuredObj)
			unstructuredObj.SetResourceVersion(existing.GetResourceVersion())
			unstructuredObj.SetSelfLink(existing.GetSelfLink())
			unstructuredObj.SetUID(existing.GetUID())
			unstructuredObj.SetCreationTimestamp(existing.GetCreationTimestamp())
			unstructuredObj.SetGeneration(existing.GetGeneration())
			if existing.GetAnnotations() != nil && existing.GetAnnotations()["deployment.kubernetes.io/revision"] != "" {
				annotations := unstructuredObj.GetAnnotations()
				if annotations == nil {
					annotations = make(map[string]string)
				}
				annotations["deployment.kubernetes.io/revision"] = existing.GetAnnotations()["deployment.kubernetes.io/revision"]
				unstructuredObj.SetAnnotations(annotations)
			}
			updated, err := client.Update(unstructuredObj, metav1.UpdateOptions{})
			if err != nil {
				c.Errorf("error updating: %s/%s/%s : %+v", unstructuredObj.GetNamespace(), resource.Resource, unstructuredObj.GetName(), err)
				continue
			}

			if updated.GetResourceVersion() == unstructuredObj.GetResourceVersion() {
				c.Debugf("%s/%s/%s (unchanged)", resource.Resource, unstructuredObj.GetNamespace(), unstructuredObj.GetName())
			} else {
				c.Infof("%s/%s/%s configured", resource.Resource, unstructuredObj.GetNamespace(), unstructuredObj.GetName())
				if logger.IsTraceEnabled() {
					// remove "runtime" fields from objects that woulds otherwise increase the verbosity of diffs
					unstructured.RemoveNestedField(unstructuredObj.Object, "metadata", "managedFields")
					unstructured.RemoveNestedField(unstructuredObj.Object, "metadata", "generation")
					unstructured.RemoveNestedField(unstructuredObj.Object, "metadata", "annotations", "deprecated.daemonset.template.generation")

					unstructured.RemoveNestedField(existing.Object, "metadata", "managedFields")
					unstructured.RemoveNestedField(existing.Object, "metadata", "generation")
					unstructured.RemoveNestedField(existing.Object, "metadata", "annotations", "deprecated.daemonset.template.generation")

					diff := deep.Equal(unstructuredObj.Object["metadata"], existing.Object["metadata"])
					if len(diff) > 0 {
						c.Tracef("%s", diff)
					}
				}
			}
		}
	}
	return nil
}

func (c *Client) Annotate(obj runtime.Object, annotations map[string]string) error {
	client, resource, unstructuredObj, err := c.GetDynamicClientFor("", obj)
	if err != nil {
		return fmt.Errorf("annotate: failed to get dynamic client: %s", err)
	}
	existing := unstructuredObj.GetAnnotations()
	for k, v := range annotations {
		existing[k] = v
	}
	unstructuredObj.SetAnnotations(existing)
	_, err = client.Update(unstructuredObj, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("annotate: failed to update object: #{err}")
	}
	c.Infof("%s/%s/%s annotated", resource.Resource, unstructuredObj.GetNamespace(), unstructuredObj.GetName())
	return nil
}

func (c *Client) Label(obj runtime.Object, labels map[string]string) error {
	client, resource, unstructuredObj, err := c.GetDynamicClientFor("", obj)
	if err != nil {
		return fmt.Errorf("label: failed to get dynamic client: %v", err)
	}
	existing := unstructuredObj.GetLabels()
	for k, v := range labels {
		existing[k] = v
	}
	unstructuredObj.SetLabels(existing)
	if _, err := client.Update(unstructuredObj, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("label: failed to update client: %v", err)
	}
	c.Infof("%s/%s/%s labelled", resource.Resource, unstructuredObj.GetNamespace(), unstructuredObj.GetName())
	return nil
}

func (c *Client) CreateOrUpdateNamespace(name string, labels, annotations map[string]string) error {
	k8s, err := c.GetClientset()
	if err != nil {
		return fmt.Errorf("createOrUpdateNamespace: failed to get client set: %v", err)
	}

	ns := k8s.CoreV1().Namespaces()
	cm, err := ns.Get(name, metav1.GetOptions{})

	if cm == nil || err != nil {
		cm = &v1.Namespace{}
		cm.Name = name
		cm.Labels = labels
		cm.Annotations = annotations

		c.Debugf("Creating namespace %s", name)
		if !c.ApplyDryRun {
			if _, err := ns.Create(cm); err != nil {
				return err
			}
		}
	} else {
		// update incoming and current labels
		if cm.ObjectMeta.Labels != nil {
			for k, v := range labels {
				cm.ObjectMeta.Labels[k] = v
			}
			labels = cm.ObjectMeta.Labels
		}

		// update incoming and current annotations
		if cm.ObjectMeta.Annotations != nil && annotations != nil {
			for k, v := range annotations {
				cm.ObjectMeta.Annotations[k] = v
			}
			annotations = cm.ObjectMeta.Annotations
		}
	}
	(*cm).Name = name
	(*cm).Labels = labels
	(*cm).Annotations = annotations
	if !c.ApplyDryRun {
		c.Debugf("Updating namespace %s", name)
		if _, err := ns.Update(cm); err != nil {
			return err
		}
	}
	return nil
}

// ForceDeleteNamespace deletes a namespace forcibly
// by overriding it's finalizers first
func (c *Client) ForceDeleteNamespace(ns string, timeout time.Duration) error {
	c.Warnf("Clearing finalizers for %v", ns)
	k8s, err := c.GetClientset()
	if err != nil {
		return fmt.Errorf("ForceDeleteNamespace: failed to get client set: %v", err)
	}

	namespace, err := k8s.CoreV1().Namespaces().Get(ns, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("ForceDeleteNamespace: failed to get namespace: %v", err)
	}
	namespace.Spec.Finalizers = []v1.FinalizerName{}
	_, err = k8s.CoreV1().Namespaces().Finalize(namespace)
	if err != nil {
		return fmt.Errorf("ForceDeleteNamespace: error removing finalisers: %v", err)
	}
	err = k8s.CoreV1().Namespaces().Delete(ns, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("ForceDeleteNamespace: error deleting namespace: %v", err)
	}
	return nil
}

// Undelete an object by removing terminationTimestamp and gracePeriod
func (c *Client) Undelete(kind, name, namespace string, object types.RuntimeObjectWithMetadata) error {
	ctx := context.Background()

	apiResource, err := c.GetAPIResource(kind)
	if err != nil {
		return perrors.Wrap(err, "failed to get api resource")
	}

	etcdClient, err := c.GetEtcdClient(ctx)
	if err != nil {
		return perrors.Wrap(err, "failed to get etcd client")
	}

	var key string
	if apiResource.Namespaced {
		key = fmt.Sprintf("/registry/%s/%s/%s", apiResource.Name, namespace, name)
	} else {
		key = fmt.Sprintf("/registry/%s/%s", apiResource.Name, name)
	}
	resp, err := etcdClient.EtcdClient.Get(ctx, key)
	if err != nil {
		return err
	}
	if len(resp.Kvs) < 1 {
		return perrors.Errorf("no results found for key %s", key)
	}

	protoSerializer, err := c.decodeProtobufResource(kind, object, resp.Kvs[0].Value)
	if err != nil {
		return perrors.Wrap(err, "failed to decode protobuf resource")
	}

	objectMeta := object.GetObjectMeta()
	if objectMeta.GetDeletionTimestamp() == nil {
		return fmt.Errorf("%s [%s] is not in terminating status", apiResource.Kind, name)
	}

	objectMeta.SetDeletionTimestamp(nil)
	objectMeta.SetDeletionGracePeriodSeconds(nil)

	var fixedResource bytes.Buffer
	// Encode fixed resource to protobuf value
	err = protoSerializer.Encode(object, &fixedResource)
	if err != nil {
		return perrors.Wrap(err, "failed to encode protobuf")
	}

	_, err = etcdClient.EtcdClient.Put(ctx, key, fixedResource.String())
	if err != nil {
		return perrors.Wrap(err, "failed to update resource in etcd")
	}

	return nil
}

// Undelete an object by removing terminationTimestamp and gracePeriod
func (c *Client) UndeleteCRD(kind, name, namespace string) error {
	ctx := context.Background()

	apiResource, err := c.GetAPIResource(kind)
	if err != nil {
		return perrors.Wrap(err, "failed to get api resource")
	}

	etcdClient, err := c.GetEtcdClient(ctx)
	if err != nil {
		return perrors.Wrap(err, "failed to get etcd client")
	}

	var key string
	if apiResource.Namespaced {
		key = fmt.Sprintf("/registry/%s/%s/%s/%s", apiResource.Group, apiResource.Name, namespace, name)
	} else {
		key = fmt.Sprintf("/registry/%s/%s/%s", apiResource.Group, apiResource.Name, name)
	}
	resp, err := etcdClient.EtcdClient.Get(ctx, key)
	if err != nil {
		return err
	}
	if len(resp.Kvs) < 1 {
		return perrors.Errorf("no results found for key %s", key)
	}

	object := &unstructured.Unstructured{}
	if err := json.Unmarshal(resp.Kvs[0].Value, &object.Object); err != nil {
		return perrors.Wrap(err, "failed to unmarshal json crd")
	}

	if object.GetDeletionTimestamp() == nil {
		return fmt.Errorf("%s [%s] is not in terminating status", apiResource.Kind, name)
	}

	object.SetDeletionTimestamp(nil)
	object.SetDeletionGracePeriodSeconds(nil)

	fixedResource, err := json.Marshal(object.Object)
	if err != nil {
		return perrors.Wrap(err, "failed to encode json object")
	}
	_, err = etcdClient.EtcdClient.Put(ctx, key, string(fixedResource))
	if err != nil {
		return perrors.Wrap(err, "failed to update resource in etcd")
	}

	return nil
}

// Orphan an object by removing ownerReferences
func (c *Client) Orphan(kind, name, namespace string, object types.RuntimeObjectWithMetadata) error {
	ctx := context.Background()

	apiResource, err := c.GetAPIResource(kind)
	if err != nil {
		return perrors.Wrap(err, "failed to get api resource")
	}

	etcdClient, err := c.GetEtcdClient(ctx)
	if err != nil {
		return perrors.Wrap(err, "failed to get etcd client")
	}

	var key string
	if apiResource.Namespaced {
		key = fmt.Sprintf("/registry/%s/%s/%s", apiResource.Name, namespace, name)
	} else {
		key = fmt.Sprintf("/registry/%s/%s", apiResource.Name, name)
	}
	resp, err := etcdClient.EtcdClient.Get(ctx, key)
	if err != nil {
		return err
	}
	if len(resp.Kvs) < 1 {
		return perrors.Errorf("no results found for key %s", key)
	}

	protoSerializer, err := c.decodeProtobufResource(kind, object, resp.Kvs[0].Value)
	if err != nil {
		return perrors.Wrap(err, "failed to decode protobuf resource")
	}

	objectMeta := object.GetObjectMeta()
	ownerReferences := objectMeta.GetOwnerReferences()
	if len(ownerReferences) == 0 {
		return fmt.Errorf("%s [%s] has no ownerReferences", apiResource.Kind, name)
	}

	objectMeta.SetOwnerReferences([]metav1.OwnerReference{})

	var fixedResource bytes.Buffer
	// Encode fixed resource to protobuf value
	err = protoSerializer.Encode(object, &fixedResource)
	if err != nil {
		return perrors.Wrap(err, "failed to encode protobuf")
	}

	_, err = etcdClient.EtcdClient.Put(ctx, key, fixedResource.String())
	if err != nil {
		return perrors.Wrap(err, "failed to update resource in etcd")
	}

	return nil
}

// Orphan an object by removing ownerReferences
func (c *Client) OrphanCRD(kind, name, namespace string) error {
	ctx := context.Background()

	apiResource, err := c.GetAPIResource(kind)
	if err != nil {
		return perrors.Wrap(err, "failed to get api resource")
	}

	etcdClient, err := c.GetEtcdClient(ctx)
	if err != nil {
		return perrors.Wrap(err, "failed to get etcd client")
	}

	var key string
	if apiResource.Namespaced {
		key = fmt.Sprintf("/registry/%s/%s/%s/%s", apiResource.Group, apiResource.Name, namespace, name)
	} else {
		key = fmt.Sprintf("/registry/%s/%s/%s", apiResource.Group, apiResource.Name, name)
	}
	resp, err := etcdClient.EtcdClient.Get(ctx, key)
	if err != nil {
		return err
	}
	if len(resp.Kvs) < 1 {
		return perrors.Errorf("no results found for key %s", key)
	}

	object := &unstructured.Unstructured{}
	if err := json.Unmarshal(resp.Kvs[0].Value, &object.Object); err != nil {
		return perrors.Wrap(err, "failed to unmarshal json crd")
	}

	if len(object.GetOwnerReferences()) == 0 {
		return fmt.Errorf("%s [%s] has no ownerReferences", apiResource.Kind, name)
	}

	object.SetOwnerReferences([]metav1.OwnerReference{})

	fixedResource, err := json.Marshal(object.Object)
	if err != nil {
		return perrors.Wrap(err, "failed to encode json object")
	}
	_, err = etcdClient.EtcdClient.Put(ctx, key, string(fixedResource))
	if err != nil {
		return perrors.Wrap(err, "failed to update resource in etcd")
	}

	return nil
}

func (c *Client) GetEtcdClient(ctx context.Context) (*etcd.Client, error) {
	clientset, err := c.GetClientset()
	if err != nil {
		return nil, perrors.Wrap(err, "failed to get clientset")
	}
	secret, err := clientset.CoreV1().Secrets("kube-system").Get("etcd-certs", metav1.GetOptions{})
	if err != nil {
		return nil, perrors.Wrap(err, "failed to get secret etcd-certs in namespace kube-system")
	}
	cert, err := certs.DecodeCertificate(secret.Data["tls.crt"], secret.Data["tls.key"])
	if err != nil {
		return nil, perrors.Wrap(err, "failed to decode etcd certificates")
	}
	etcdClientGenerator, err := c.GetEtcdClientGenerator(cert)
	if err != nil {
		return nil, perrors.Wrap(err, "failed to get etcd client generator")
	}

	masterNode, err := c.GetMasterNode()
	if err != nil {
		return nil, perrors.Wrap(err, "failed to get master node")
	}
	etcdClient, err := etcdClientGenerator.ForNode(ctx, masterNode)
	if err != nil {
		return nil, perrors.Wrap(err, "failed to get etcd client")
	}

	return etcdClient, nil
}

func (c *Client) GetAPIResource(name string) (*metav1.APIResource, error) {
	clientset, err := c.GetClientset()
	if err != nil {
		return nil, perrors.Wrap(err, "failed to get clientset")
	}

	rm, err := c.GetRestMapper()
	if err != nil {
		return nil, perrors.Wrap(err, "failed to get rest mapper")
	}

	resources, err := clientset.ServerResources()
	if err != nil {
		return nil, perrors.Wrap(err, "failed to get server resources")
	}

	for _, list := range resources {
		for _, resource := range list.APIResources {
			var singularName string
			if resource.Name != name {
				singularName, err = rm.ResourceSingularizer(resource.Name)
				if err != nil {
					continue
				}
			}

			if resource.Name == name || singularName == name {
				parts := strings.Split(list.GroupVersion, "/")
				if len(parts) >= 2 {
					resource.Group = parts[0]
					resource.Version = parts[1]
				} else {
					resource.Group = ""
					resource.Version = parts[0]
				}
				return &resource, nil
			}
		}
	}

	return nil, perrors.Errorf("no resource with name %s found", name)
}

func (c *Client) HasSecret(ns, name string) bool {
	client, err := c.GetClientset()
	if err != nil {
		c.Tracef("hasSecret: failed to get client set: %v", err)
		return false
	}
	secrets := client.CoreV1().Secrets(ns)
	cm, err := secrets.Get(name, metav1.GetOptions{})
	return cm != nil && err == nil
}

func (c *Client) HasConfigMap(ns, name string) bool {
	client, err := c.GetClientset()
	if err != nil {
		c.Tracef("hasConfigMap: failed to get client set: %v", err)
		return false
	}
	configmaps := client.CoreV1().ConfigMaps(ns)
	cm, err := configmaps.Get(name, metav1.GetOptions{})
	return cm != nil && err == nil
}

func (c *Client) GetOrCreateSecret(name, ns string, data map[string][]byte) error {
	if c.HasSecret(name, ns) {
		return nil
	}
	return c.CreateOrUpdateSecret(name, ns, data)
}

func (c *Client) CreateOrUpdateSecret(name, ns string, data map[string][]byte) error {
	if c.ApplyDryRun {
		c.Debugf("[dry-run] secrets/%s/%s created/configured", ns, name)
		return nil
	}
	return c.Apply(ns, &v1.Secret{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Data:       data,
	})
}

func (c *Client) CreateOrUpdateConfigMap(name, ns string, data map[string]string) error {
	if c.ApplyDryRun {
		c.Debugf("[dry-run] configmaps/%s/%s created/configured", ns, name)
		return nil
	}
	return c.Apply(ns, &v1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Data:       data})
}

func (c *Client) ExposeIngress(namespace, service string, domain string, port int, annotations map[string]string) error {
	k8s, err := c.GetClientset()
	if err != nil {
		return fmt.Errorf("exposeIngress: failed to get client set: %v", err)
	}
	ingresses := k8s.NetworkingV1beta1().Ingresses(namespace)
	ingress, err := ingresses.Get(service, metav1.GetOptions{})
	if ingress == nil || err != nil {
		ingress = &v1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:        service,
				Namespace:   namespace,
				Annotations: annotations,
			},
			Spec: v1beta1.IngressSpec{
				TLS: []v1beta1.IngressTLS{
					{
						Hosts: []string{domain},
					},
				},
				Rules: []v1beta1.IngressRule{
					{
						Host: domain,
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Backend: v1beta1.IngressBackend{
											ServiceName: service,
											ServicePort: intstr.FromInt(port),
										},
									},
								},
							},
						},
					},
				},
			},
		}
		c.Infof("Creating %s/ingress/%s", namespace, service)
		if !c.ApplyDryRun {
			if _, err := ingresses.Create(ingress); err != nil {
				return fmt.Errorf("exposeIngress: failed to create ingress: %v", err)
			}
		}
	}
	return nil
}

func (c *Client) GetOrCreatePVC(namespace, name, size, class string) error {
	client, err := c.GetClientset()
	if err != nil {
		return fmt.Errorf("getOrCreatePVC: failed to get client set: %v", err)
	}
	qty, err := resource.ParseQuantity(size)
	if err != nil {
		return fmt.Errorf("getOrCreatePVC: failed to parse quantity: %v", err)
	}
	pvcs := client.CoreV1().PersistentVolumeClaims(namespace)

	existing, err := pvcs.Get(name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		c.Tracef("GetOrCreatePVC: failed to get PVC: %s", err)
		c.Infof("Creating PVC %s/%s (%s %s)\n", namespace, name, size, class)
		_, err = pvcs.Create(&v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: v1.PersistentVolumeClaimSpec{
				StorageClassName: &class,
				AccessModes: []v1.PersistentVolumeAccessMode{
					v1.ReadWriteOnce,
				},
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: qty,
					},
				},
			},
		})
	} else if err != nil {
		return fmt.Errorf("getOrCreatePVC: failed to create PVC: %v", err)
	} else {
		c.Infof("Found existing PVC %s/%s (%s %s) ==> %s\n", namespace, name, size, class, existing.UID)
		return nil
	}
	return err
}

func (c *Client) StreamLogs(namespace, name string) error {
	client, err := c.GetClientset()
	if err != nil {
		return err
	}
	pods := client.CoreV1().Pods(namespace)
	pod, err := pods.Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	c.Debugf("Waiting for %s/%s to be running", namespace, name)
	if err := c.WaitForPod(namespace, name, 120*time.Second, v1.PodRunning, v1.PodSucceeded); err != nil {
		return err
	}
	c.Debugf("%s/%s running, streaming logs", namespace, name)
	var wg sync.WaitGroup
	for _, container := range append(pod.Spec.Containers, pod.Spec.InitContainers...) {
		logs := pods.GetLogs(pod.Name, &v1.PodLogOptions{
			Container: container.Name,
		})

		prefix := pod.Name
		if len(pod.Spec.Containers) > 1 {
			prefix += "/" + container.Name
		}
		podLogs, err := logs.Stream()
		if err != nil {
			return err
		}
		wg.Add(1)
		go func() {
			defer podLogs.Close()
			defer wg.Done()

			scanner := bufio.NewScanner(podLogs)
			for scanner.Scan() {
				incoming := scanner.Bytes()
				buffer := make([]byte, len(incoming))
				copy(buffer, incoming)
				fmt.Printf("\x1b[38;5;244m[%s]\x1b[0m %s\n", prefix, string(buffer))
			}
		}()
	}
	wg.Wait()
	if err = c.WaitForPod(namespace, name, 120*time.Second, v1.PodSucceeded); err != nil {
		return err
	}
	pod, err = pods.Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if pod.Status.Phase == v1.PodSucceeded {
		return nil
	}
	return fmt.Errorf("pod did not finish successfully %s - %s", pod.Status.Phase, pod.Status.Message)
}

func CreateKubeConfig(clusterName string, ca certs.CertificateAuthority, endpoint string, group string, user string, expiry time.Duration) ([]byte, error) {
	contextName := fmt.Sprintf("%s@%s", user, clusterName)
	cert := certs.NewCertificateBuilder(user).Organization(group).Client().Certificate
	if cert.X509.PublicKey == nil && cert.PrivateKey != nil {
		cert.X509.PublicKey = cert.PrivateKey.Public()
	}
	signed, err := ca.Sign(cert.X509, expiry)
	if err != nil {
		return nil, fmt.Errorf("createKubeConfig: failed to sign certificate: %v", err)
	}
	cert = &certs.Certificate{
		X509:       signed,
		PrivateKey: cert.PrivateKey,
	}

	cfg := api.Config{
		Clusters: map[string]*api.Cluster{
			clusterName: {
				Server:                "https://" + endpoint + ":6443",
				InsecureSkipTLSVerify: true,
				// The CA used for signing the client certificate is not the same as the
				// as the CA (kubernetes-ca) that signed the api-server cert. The kubernetes-ca
				// is ephemeral.
				// TODO dynamically download CA from master server
				// CertificateAuthorityData: []byte(platform.Certificates.CA.X509),
			},
		},
		Contexts: map[string]*api.Context{
			contextName: {
				Cluster:   clusterName,
				AuthInfo:  contextName,
				Namespace: "kube-system",
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			contextName: {
				ClientKeyData:         cert.EncodedPrivateKey(),
				ClientCertificateData: cert.EncodedCertificate(),
			},
		},
		CurrentContext: contextName,
	}

	return clientcmd.Write(cfg)
}

func CreateOIDCKubeConfig(clusterName string, ca certs.CertificateAuthority, endpoint, idpURL, idToken, accessToken, refreshToken string) ([]byte, error) {
	if !strings.HasPrefix("https://", endpoint) {
		endpoint = "https://" + endpoint
	}

	if !strings.HasPrefix("https://", idpURL) {
		idpURL = "https://" + idpURL
	}
	cfg := api.Config{
		Clusters: map[string]*api.Cluster{
			clusterName: {
				Server:                endpoint + ":6443",
				InsecureSkipTLSVerify: true,
			},
		},
		Contexts: map[string]*api.Context{
			clusterName: {
				Cluster:  clusterName,
				AuthInfo: "sso@" + clusterName,
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			"sso@" + clusterName: {
				AuthProvider: &api.AuthProviderConfig{
					Name: "oidc",
					Config: map[string]string{
						"client-id":                      "kubernetes",
						"client-secret":                  "ZXhhbXBsZS1hcHAtc2VjcmV0",
						"extra-scopes":                   "offline_access openid profile email groups",
						"idp-certificate-authority-data": base64.StdEncoding.EncodeToString(ca.GetPublicChain()[0].EncodedCertificate()),
						"idp-issuer-url":                 idpURL,
						"id-token":                       idToken,
						"access-token":                   accessToken,
						"refresh-token":                  refreshToken,
					},
				},
			},
		},
		CurrentContext: clusterName,
	}

	return clientcmd.Write(cfg)
}

// CreateMultiKubeConfig creates a kubeconfig file contents for a map of
// cluster name -> cluster API endpoint hosts, all with a shared
// user name, group and cert expiry.
// NOTE: these clusters all need to share the same plaform CA
func CreateMultiKubeConfig(ca certs.CertificateAuthority, clusters map[string]string, group string, user string, expiry time.Duration) ([]byte, error) {
	if len(clusters) < 1 {
		return []byte{}, fmt.Errorf("CreateMultiKubeConfig failed since it was given an empty cluster map")
	}
	cfg := api.Config{
		Clusters:       map[string]*api.Cluster{},
		Contexts:       map[string]*api.Context{},
		AuthInfos:      map[string]*api.AuthInfo{},
		CurrentContext: "",
	}
	for clusterName, endpoint := range clusters {
		cert := certs.NewCertificateBuilder(user).Organization(group).Client().Certificate
		if cert.X509.PublicKey == nil && cert.PrivateKey != nil {
			cert.X509.PublicKey = cert.PrivateKey.Public()
		}
		signed, err := ca.Sign(cert.X509, expiry)
		if err != nil {
			return nil, fmt.Errorf("createKubeConfig: failed to sign certificate: %v", err)
		}
		cert = &certs.Certificate{
			X509:       signed,
			PrivateKey: cert.PrivateKey,
		}
		cfg.Clusters[clusterName] = &api.Cluster{
			Server:                endpoint,
			InsecureSkipTLSVerify: true,
		}
		context := fmt.Sprintf("%s@%s", user, clusterName)
		cfg.Contexts[clusterName] = &api.Context{
			Cluster:   clusterName,
			AuthInfo:  context,
			Namespace: "kube-system", //TODO: verify
		}
		cfg.AuthInfos[context] = &api.AuthInfo{
			ClientKeyData:         cert.EncodedPrivateKey(),
			ClientCertificateData: cert.EncodedCertificate(),
		}
	}
	return clientcmd.Write(cfg)
}

// PingMaster attempts to connect to the API server and list nodes and services
// to ensure the API server is ready to accept any traffic
func (c *Client) PingMaster() bool {
	client, err := c.GetClientset()
	if err != nil {
		c.Tracef("pingMaster: Failed to get clientset: %v", err)
		return false
	}

	nodes, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		c.Tracef("pingMaster: Failed to get nodes list: %v", err)
		return false
	}
	if nodes == nil && len(nodes.Items) == 0 {
		return false
	}

	_, err = client.CoreV1().ServiceAccounts("kube-system").Get("default", metav1.GetOptions{})
	if err != nil {
		c.Tracef("pingMaster: Failed to get service account: %v", err)
		return false
	}
	return true
}

func (c *Client) WaitForResource(kind, namespace, name string, timeout time.Duration) error {
	client, err := c.GetClientByKind(kind)
	if err != nil {
		return err
	}
	start := time.Now()
	for {
		item, err := client.Namespace(namespace).Get(name, metav1.GetOptions{})

		if errors.IsNotFound(err) {
			return err
		}

		if start.Add(timeout).Before(time.Now()) {
			return fmt.Errorf("timeout exceeded waiting for %s/%s is %s, error: %v", kind, name, "", err)
		}

		if err != nil {
			c.Debugf("Unable to get %s/%s: %v", kind, name, err)
			continue
		}

		conditions := item.Object["status"].(map[string]interface{})["conditions"].([]interface{})

		for _, raw := range conditions {
			condition := raw.(map[string]interface{})
			c.Debugf("%s/%s is %s/%s: %s", namespace, name, condition["type"], condition["status"], condition["message"])
			if condition["type"] == "Ready" && condition["status"] == "True" {
				return nil
			}
		}
		time.Sleep(1 * time.Second)
	}
}

// WaitForPod waits for a pod to be in the specified phase, or returns an
// error if the timeout is exceeded
func (c *Client) WaitForPod(ns, name string, timeout time.Duration, phases ...v1.PodPhase) error {
	client, err := c.GetClientset()
	if err != nil {
		return fmt.Errorf("waitForPod: Failed to get clientset: %v", err)
	}
	pods := client.CoreV1().Pods(ns)
	start := time.Now()
	for {
		pod, err := pods.Get(name, metav1.GetOptions{})
		if start.Add(timeout).Before(time.Now()) {
			return fmt.Errorf("timeout exceeded waiting for %s is %s, error: %v", name, pod.Status.Phase, err)
		}

		if pod == nil || pod.Status.Phase == v1.PodPending {
			time.Sleep(5 * time.Second)
			continue
		}
		if pod.Status.Phase == v1.PodFailed {
			return nil
		}

		for _, phase := range phases {
			if pod.Status.Phase == phase {
				return nil
			}
		}
	}
}

// WaitForDeployment waits for a deployment to have at least 1 ready replica, or returns an
// error if the timeout is exceeded
func (c *Client) WaitForDeployment(ns, name string, timeout time.Duration) error {
	client, err := c.GetClientset()
	if err != nil {
		return err
	}
	deployments := client.AppsV1().Deployments(ns)
	start := time.Now()
	msg := false
	for {
		deployment, _ := deployments.Get(name, metav1.GetOptions{})
		if start.Add(timeout).Before(time.Now()) {
			return fmt.Errorf("timeout exceeded waiting for deployment to become ready %s", name)
		}
		if deployment != nil && deployment.Status.ReadyReplicas >= 1 {
			return nil
		}

		if !msg {
			c.Infof("waiting for %s/%s to have 1 ready replica", ns, name)
			msg = true
		}

		time.Sleep(2 * time.Second)
	}
}

func (c *Client) GetConditionsForNode(name string) (map[v1.NodeConditionType]v1.ConditionStatus, error) {
	client, err := c.GetClientset()
	if err != nil {
		return nil, err
	}
	node, err := client.CoreV1().Nodes().Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, nil
	}

	var out = make(map[v1.NodeConditionType]v1.ConditionStatus)
	for _, condition := range node.Status.Conditions {
		out[condition.Type] = condition.Status
	}
	return out, nil
}

// WaitForNode waits for a pod to be in the specified phase, or returns an
// error if the timeout is exceeded
func (c *Client) WaitForNode(name string, timeout time.Duration, condition v1.NodeConditionType, statii ...v1.ConditionStatus) (map[v1.NodeConditionType]v1.ConditionStatus, error) {
	start := time.Now()
	for {
		conditions, err := c.GetConditionsForNode(name)
		if start.Add(timeout).Before(time.Now()) {
			return conditions, fmt.Errorf("timeout exceeded waiting for %s is %s, error: %v", name, conditions, err)
		}

		for _, status := range statii {
			if conditions[condition] == status {
				return conditions, nil
			}
		}
		time.Sleep(2 * time.Second)
	}
}

// WaitForNode waits for a pod to be in the specified phase, or returns an
// error if the timeout is exceeded
func (c *Client) WaitForTaintRemoval(name string, timeout time.Duration, taintKey string) error {
	start := time.Now()
outerLoop:
	for {
		if time.Since(start) > timeout {
			return fmt.Errorf("timeout exceeded waiting for %s to not have %s", name, taintKey)
		}

		client, err := c.GetClientset()
		if err != nil {
			return err
		}
		node, err := client.CoreV1().Nodes().Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		for _, taint := range node.Spec.Taints {
			if taint.Key == taintKey {
				time.Sleep(2 * time.Second)
				continue outerLoop
			}
		}
		// taint not found
		return nil
	}
}

// WaitForPodCommand waits for a command executed in pod to succeed with an exit code of 9
// error if the timeout is exceeded
func (c *Client) WaitForPodCommand(ns, name string, container string, timeout time.Duration, command ...string) error {
	start := time.Now()
	for {
		stdout, stderr, err := c.ExecutePodf(ns, name, container, command...)
		if err == nil {
			return nil
		}
		if start.Add(timeout).Before(time.Now()) {
			return fmt.Errorf("timeout exceeded waiting for %s stdout: %s, stderr: %s", name, stdout, stderr)
		}
		time.Sleep(5 * time.Second)
	}
}

// ExecutePodf runs the specified shell command inside a container of the specified pod
func (c *Client) ExecutePodf(namespace, pod, container string, command ...string) (string, string, error) {
	client, err := c.GetClientset()
	if err != nil {
		return "", "", fmt.Errorf("executePodf: Failed to get clientset: %v", err)
	}
	c.Debugf("[%s/%s/%s] %s", namespace, pod, container, command)
	const tty = false
	req := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
		SubResource("exec").
		Param("container", container)
	req.VersionedParams(&v1.PodExecOptions{
		Container: container,
		Command:   command,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       tty,
	}, scheme.ParameterCodec)

	rc, err := c.GetRESTConfig()
	if err != nil {
		return "", "", fmt.Errorf("ExecutePodf: Failed to get REST config: %v", err)
	}

	exec, err := remotecommand.NewSPDYExecutor(rc, "POST", req.URL())
	if err != nil {
		return "", "", fmt.Errorf("ExecutePodf: Failed to get SPDY Executor: %v", err)
	}
	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    tty,
	})

	_stdout := safeString(&stdout)
	_stderr := safeString(&stderr)
	if err != nil {
		return _stdout, _stderr, fmt.Errorf("exec returned an error: %+v", err)
	}

	c.Tracef("[%s/%s/%s] %s => %s %s ", namespace, pod, container, command, _stdout, _stderr)
	return _stdout, _stderr, nil
}

func safeString(buf *bytes.Buffer) string {
	if buf == nil || buf.Len() == 0 {
		return ""
	}
	return buf.String()
}

// Executef runs the specified shell command on a node by creating
// a pre-scheduled pod that runs in the host namespace
func (c *Client) Executef(node string, timeout time.Duration, command string, args ...interface{}) (string, error) {
	client, err := c.GetClientset()
	if err != nil {
		return "", fmt.Errorf("executef: Failed to get clientset: %v", err)
	}
	pods := client.CoreV1().Pods("kube-system")
	command = fmt.Sprintf(command, args...)
	pod, err := pods.Create(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("command-%s-%s", node, utils.ShortTimestamp()),
		},
		Spec: NewCommandJob(node, command),
	})
	c.Tracef("[%s] executing '%s' in pod %s", node, command, pod.Name)
	if err != nil {
		return "", fmt.Errorf("executef: Failed to create pod: %v", err)
	}
	defer pods.Delete(pod.ObjectMeta.Name, &metav1.DeleteOptions{}) // nolint: errcheck

	logs := pods.GetLogs(pod.Name, &v1.PodLogOptions{
		Container: pod.Spec.Containers[0].Name,
	})

	err = c.WaitForPod("kube-system", pod.ObjectMeta.Name, timeout, v1.PodSucceeded)
	logString := read(logs)
	if err != nil {
		return logString, fmt.Errorf("failed to execute command, pod did not complete: %v", err)
	}
	c.Tracef("[%s] stdout: %s", node, logString)
	return logString, nil
}

func read(req *rest.Request) string {
	stream, err := req.Stream()
	if err != nil {
		return fmt.Sprintf("Failed to stream logs %v", err)
	}
	data, err := ioutil.ReadAll(stream)
	if err != nil {
		return fmt.Sprintf("Failed to stream logs %v", err)
	}
	return string(data)
}

func NewCommandJob(node, command string) v1.PodSpec {
	yes := true
	return v1.PodSpec{
		RestartPolicy: v1.RestartPolicyNever,
		NodeName:      node,
		Volumes: []v1.Volume{{
			Name: "root",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: "/",
				},
			},
		}},
		Containers: []v1.Container{{
			Name:  "shell",
			Image: "docker.io/ubuntu:18.04",
			Command: []string{
				"sh",
				"-c",
				"chroot /chroot bash -c \"" + command + "\"",
			},
			VolumeMounts: []v1.VolumeMount{{
				Name:      "root",
				MountPath: "/chroot",
			}},
			SecurityContext: &v1.SecurityContext{
				Privileged: &yes,
			},
		}},
		Tolerations: []v1.Toleration{
			{
				// tolerate all values
				Operator: "Exists",
			},
		},
		HostNetwork: true,
		HostPID:     true,
		HostIPC:     true,
	}
}

// GetMasterNode returns the name of the first node found labelled as a master
func (c *Client) GetMasterNode() (string, error) {
	client, err := c.GetClientset()
	if err != nil {
		return "", fmt.Errorf("GetMasterNode: Failed to get clientset: %v", err)
	}

	nodes, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	for _, node := range nodes.Items {
		if IsMasterNode(node) {
			return node.Name, nil
		}
	}
	return "", fmt.Errorf("no master nodes found")
}

// GetMasterNode returns a list of all master nodes
func (c *Client) GetMasterNodes() ([]string, error) {
	client, err := c.GetClientset()
	if err != nil {
		return nil, nil
	}

	nodes, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, nil
	}

	var nodeNames []string
	for _, node := range nodes.Items {
		if IsMasterNode(node) {
			nodeNames = append(nodeNames, node.Name)
		}
	}
	return nodeNames, nil
}

// Returns the first pod found by label
func (c *Client) GetFirstPodByLabelSelector(namespace string, labelSelector string) (*v1.Pod, error) {
	client, err := c.GetClientset()
	if err != nil {
		return nil, fmt.Errorf("GetFirstPodByLabelSelector: Failed to get clientset: %v", err)
	}

	pods, err := client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("GetFirstPodByLabelSelector: Failed to query for %v in namespace %v: %v", labelSelector, namespace, err)
	}

	if (pods != nil && len(pods.Items) < 1) || pods == nil {
		return nil, fmt.Errorf("GetFirstPodByLabelSelector: No pods found for query for %v in namespace %v: %v", labelSelector, namespace, err)
	}

	return &pods.Items[0], nil
}

func (c *Client) GetEventsFor(kind string, object metav1.Object) ([]v1.Event, error) {
	client, err := c.GetClientset()
	if err != nil {
		return nil, err
	}
	selector := client.CoreV1().Events(object.GetNamespace()).GetFieldSelector(
		pointer.ToString(object.GetName()),
		pointer.ToString(object.GetNamespace()),
		&kind,
		pointer.ToString(string(object.GetUID())))
	events, err := client.CoreV1().Events(object.GetNamespace()).List(metav1.ListOptions{
		FieldSelector: selector.String(),
	})
	if err != nil {
		return nil, err
	}
	return events.Items, nil
}

func (c *Client) GetHealth() Health {
	health := Health{}
	client, err := c.GetClientset()
	if err != nil {
		return Health{Error: err}
	}
	pods, err := client.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		return Health{Error: err}
	}

	for _, pod := range pods.Items {
		if IsDeleted(&pod) {
			continue
		}
		if pod.Spec.Priority != nil && *pod.Spec.Priority < 0 {
			continue
		}

		if IsPodCrashLoopBackoff(pod) {
			health.CrashLoopBackOff++
		} else if IsPodHealthy(pod) {
			health.RunningPods++
		} else if IsPodPending(pod) {
			health.PendingPods++
		} else {
			health.ErrorPods++
		}
	}
	return health
}

func (c *Client) decodeProtobufResource(kind string, object runtime.Object, message []byte) (*protobuf.Serializer, error) {
	rm, err := c.GetRestMapper()
	if err != nil {
		return nil, perrors.Wrap(err, "failed to get rest mapper")
	}
	gvks, err := rm.KindsFor(schema.GroupVersionResource{
		Resource: kind,
	})
	if err != nil {
		return nil, perrors.Wrapf(err, "failed to get kind for %s", kind)
	}
	if len(gvks) == 0 {
		return nil, perrors.Errorf("no gvks returned for kind %s", kind)
	}

	for _, gvk := range gvks {
		runtimeScheme := runtime.NewScheme()
		runtimeScheme.AddKnownTypeWithName(gvk, object)
		protoSerializer := protobuf.NewSerializer(runtimeScheme, runtimeScheme)

		// Decode protobuf value to Go pv struct
		_, _, err = protoSerializer.Decode(message, &gvk, object)
		if err == nil {
			return protoSerializer, nil
		}
	}

	return nil, perrors.Errorf("failed to decode protobuf message into runtime object, failed to find any suitable gvk")
}
