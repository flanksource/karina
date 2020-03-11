package k8s

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	certs "github.com/flanksource/commons/certs"
	"github.com/flanksource/commons/files"
	utils "github.com/flanksource/commons/utils"
	"github.com/go-test/deep"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	"sigs.k8s.io/yaml"

	"github.com/moshloop/platform-cli/pkg/k8s/kustomize"
)

type Client struct {
	GetKubeConfigBytes  func() ([]byte, error)
	ApplyDryRun         bool
	Trace               bool
	GetKustomizePatches func() ([]string, error)
	client              *kubernetes.Clientset
	dynamicClient       dynamic.Interface
	restConfig          *rest.Config

	kustomizeManager *kustomize.Manager
	restMapper       meta.RESTMapper
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
	for _, patch := range patches {
		if files.Exists(patch) {
			if err := files.Copy(patch, dir+"/"+files.GetBaseName(patch)); err != nil {
				return nil, err
			}
		} else {
			name := fmt.Sprintf("patch-%d.yml", no)
			no++
			if _, err := files.CopyFromReader(bytes.NewBufferString(patch), dir+"/"+name, 0644); err != nil {
				return nil, err
			}
		}
	}

	kustomizeManager, err := kustomize.GetManager(dir)
	c.kustomizeManager = kustomizeManager
	return c.kustomizeManager, err
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
	c.restConfig, err = clientcmd.RESTConfigFromKubeConfig(data)
	return c.restConfig, err
}

// GetSecret returns the data of a secret or nil for any error
func (c *Client) GetSecret(namespace, name string) *map[string][]byte {
	k8s, err := c.GetClientset()
	if err != nil {
		log.Tracef("Failed to get client %v", err)
		return nil
	}
	secret, err := k8s.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		log.Tracef("Failed to get secret %s/%s: %v\n", namespace, name, err)
		return nil
	}
	return &secret.Data
}

// GetConfigMap returns the data of a secret or nil for any error
func (c *Client) GetConfigMap(namespace, name string) *map[string]string {
	k8s, err := c.GetClientset()
	if err != nil {
		log.Tracef("Failed to get client %v", err)
		return nil
	}
	cm, err := k8s.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		log.Tracef("failed to get secret %s/%s: %v\n", namespace, name, err)
		return nil
	}
	return &cm.Data
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
	if log.IsLevelEnabled(log.TraceLevel) {
		spew.Dump(unstructuredObj.Object)
	}

	// FIXME(moshloop) getting the zalando operationconfiguration fails with "unrecognized type: int64" so we fall back to brute-force
	log.Warnf("Using mapstructure to decode %s: %v", obj.GetObjectKind().GroupVersionKind().Kind, err)
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

func (c *Client) GetDynamicClientFor(namespace string, obj runtime.Object) (dynamic.ResourceInterface, *schema.GroupVersionResource, *unstructured.Unstructured, error) {
	dynamicClient, err := c.GetDynamicClient()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("getDynamicClientFor: failed to get dynamic client: %v", err)
	}

	gvk := obj.GetObjectKind().GroupVersionKind()
	gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}
	rm, _ := c.GetRestMapper()
	mapping, err := rm.RESTMapping(gk, gvk.Version)
	if err != nil && meta.IsNoMatchError(err) {
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
	} else {
		if namespace == "" {
			namespace = unstructuredObj.GetNamespace()
		}
		return dynamicClient.Resource(mapping.Resource).Namespace(namespace), &resource, unstructuredObj, nil
	}

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

func (c *Client) ApplyUnstructured(namespace string, objects ...*unstructured.Unstructured) error {
	for _, unstructuredObj := range objects {

		client, err := c.GetRestClient(*unstructuredObj)
		if err != nil {
			return err
		}

		if c.ApplyDryRun {
			log.Infof("[dry-run] %s/%s/%s created/configured", client.Resource, unstructuredObj, unstructuredObj.GetName())
		} else {
			_, err = client.Create(namespace, true, unstructuredObj, &metav1.CreateOptions{})
			if errors.IsAlreadyExists(err) {
				_, err = client.Replace(namespace, unstructuredObj.GetName(), true, unstructuredObj)
				if err != nil {
					log.Errorf("error handling: %s : %+v", client.Resource, err)
				} else {
					// TODO(moshloop): Diff the old and new objects and log unchanged instead of configured where necessary
					log.Infof("%s/%s/%s configured", client.Resource, unstructuredObj.GetNamespace(), unstructuredObj.GetName())
				}
			} else if err == nil {
				log.Infof("%s/%s/%s created", client.Resource, unstructuredObj.GetNamespace(), unstructuredObj.GetName())
			} else {
				log.Errorf("error handling: %s : %+v", client.Resource, err)
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
			log.Errorf("Error tracing %s", err)
		} else {
			fmt.Println(string(data))
		}
	}
}

func (c *Client) Apply(namespace string, objects ...runtime.Object) error {
	for _, obj := range objects {
		client, resource, unstructuredObj, err := c.GetDynamicClientFor(namespace, obj)
		if err != nil {
			return fmt.Errorf("failed to get dynamic client for %v: %v", obj, err)
		}

		if c.ApplyDryRun {
			c.trace("apply", unstructuredObj)
			log.Infof("[dry-run] %s/%s created/configured", resource.Resource, unstructuredObj.GetName())
			continue
		}

		existing, err := client.Get(unstructuredObj.GetName(), metav1.GetOptions{})

		if existing == nil {
			c.trace("creating", unstructuredObj)
			_, err = client.Create(unstructuredObj, metav1.CreateOptions{})
			if err != nil {
				log.Errorf("error creating: %s/%s/%s : %+v", resource.Group, resource.Version, resource.Resource, err)
			}
			log.Infof("%s/%s/%s created", resource.Resource, unstructuredObj.GetNamespace(), unstructuredObj.GetName())
		} else {
			if unstructuredObj.GetKind() == "Service" {
				// Workaround for immutable spec.clusterIP error message
				spec := unstructuredObj.Object["spec"].(map[string]interface{})
				spec["clusterIP"] = existing.Object["spec"].(map[string]interface{})["clusterIP"]
			}

			c.trace("updating", unstructuredObj)
			unstructuredObj.SetResourceVersion(existing.GetResourceVersion())
			updated, err := client.Update(unstructuredObj, metav1.UpdateOptions{})
			if err != nil {
				log.Errorf("error updating: %s/%s/%s : %+v", resource.Group, resource.Version, resource.Resource, err)
				continue
			}

			if updated.GetResourceVersion() == unstructuredObj.GetResourceVersion() {
				log.Debugf("%s/%s/%s (unchanged)", resource.Resource, unstructuredObj.GetNamespace(), unstructuredObj.GetName())
			} else {
				log.Infof("%s/%s/%s configured", resource.Resource, unstructuredObj.GetNamespace(), unstructuredObj.GetName())
				if log.IsLevelEnabled(log.TraceLevel) {
					diff := deep.Equal(unstructuredObj.Object["metadata"], existing.Object["metadata"])
					if len(diff) > 0 {
						log.Tracef("%s", diff)
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
	log.Infof("%s/%s/%s annotated", resource.Resource, unstructuredObj.GetNamespace(), unstructuredObj.GetName())
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
	log.Infof("%s/%s/%s labelled", resource.Resource, unstructuredObj.GetNamespace(), unstructuredObj.GetName())
	return nil
}

func (c *Client) CreateOrUpdateNamespace(name string, labels map[string]string, annotations map[string]string) error {
	k8s, err := c.GetClientset()
	if err != nil {
		return fmt.Errorf("createOrUpdateNamespace: failed to get client set: %v", err)
	}

	// set default labels
	defaultLabels := make(map[string]string)
	defaultLabels["openpolicyagent.org/webhook"] = "ignore"
	if labels != nil {
		for k, v := range defaultLabels {
			labels[k] = v
		}
	} else {
		labels = defaultLabels
	}

	ns := k8s.CoreV1().Namespaces()
	cm, err := ns.Get(name, metav1.GetOptions{})

	if cm == nil || err != nil {
		cm = &v1.Namespace{}
		cm.Name = name
		cm.Labels = labels
		cm.Annotations = annotations

		log.Infof("Creating namespace %s", name)
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
		switch {
		case cm.ObjectMeta.Annotations != nil && annotations == nil:
			annotations = cm.ObjectMeta.Annotations
		case cm.ObjectMeta.Annotations != nil && annotations != nil:
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
		log.Infof("Updating namespace %s", name)
		if _, err := ns.Update(cm); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) HasSecret(ns, name string) bool {
	client, err := c.GetClientset()
	if err != nil {
		log.Tracef("hasSecret: failed to get client set: %v", err)
		return false
	}
	secrets := client.CoreV1().Secrets(ns)
	cm, err := secrets.Get(name, metav1.GetOptions{})
	return cm != nil && err == nil

}

func (c *Client) HasConfigMap(ns, name string) bool {
	client, err := c.GetClientset()
	if err != nil {
		log.Tracef("hasConfigMap: failed to get client set: %v", err)
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
	client, err := c.GetClientset()
	if err != nil {
		return fmt.Errorf("createOrUpdateSecret: failed to get clientset: %v", err)
	}
	secrets := client.CoreV1().Secrets(ns)
	cm, err := secrets.Get(name, metav1.GetOptions{})
	if cm == nil || err != nil {
		cm = &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
			Data:       data,
		}
		log.Infof("Creating %s/secret/%s", ns, name)
		if !c.ApplyDryRun {
			if _, err := secrets.Create(cm); err != nil {
				return fmt.Errorf("createOrUpdateSecret: failed to namespace: %v", err)
			}
		}
	} else {
		(*cm).Data = data
		if !c.ApplyDryRun {
			log.Infof("Updating %s/secret/%s", ns, name)
			if _, err := secrets.Update(cm); err != nil {
				return fmt.Errorf("createOrUpdateSecret: failed to update configmap: %v", err)
			}
		}
	}
	return nil
}

func (c *Client) CreateOrUpdateConfigMap(name, ns string, data map[string]string) error {
	client, err := c.GetClientset()
	if err != nil {
		return fmt.Errorf("createOrUpdateConfigMap: failed to get client set: %v", err)
	}
	configs := client.CoreV1().ConfigMaps(ns)
	cm, err := configs.Get(name, metav1.GetOptions{})
	if cm == nil || err != nil {
		cm = &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
			Data:       data}
		log.Infof("Creating %s/cm/%s", ns, name)
		if !c.ApplyDryRun {
			if _, err := configs.Create(cm); err != nil {
				return fmt.Errorf("createOrUpdateConfigMap: failed to update configmap: %v", err)
			}
		}
	} else {
		(*cm).Data = data
		if !c.ApplyDryRun {
			log.Infof("Updating %s/cm/%s", ns, name)
			if _, err := configs.Update(cm); err != nil {
				return fmt.Errorf("createOrUpdateConfigMap: failed to update configmap: %v", err)
			}
		}
	}
	return nil
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
					v1beta1.IngressTLS{
						Hosts: []string{domain},
					},
				},
				Rules: []v1beta1.IngressRule{
					v1beta1.IngressRule{
						Host: domain,
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									v1beta1.HTTPIngressPath{
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
		log.Infof("Creating %s/ingress/%s", namespace, service)
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
		log.Tracef("GetOrCreatePVC: failed to get PVC: %s", err)
		log.Infof("Creating PVC %s/%s (%s %s)\n", namespace, name, size, class)
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
		log.Infof("Found existing PVC %s/%s (%s %s) ==> %s\n", namespace, name, size, class, existing.UID)
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
	log.Debugf("Waiting for %s/%s to be running", namespace, name)
	if err := c.WaitForPod(namespace, name, 120*time.Second, v1.PodRunning, v1.PodSucceeded); err != nil {
		return err
	}
	log.Debugf("%s/%s running, streaming logs", namespace, name)
	var wg sync.WaitGroup
	for _, container := range append(pod.Spec.Containers, pod.Spec.InitContainers...) {
		var logs *rest.Request
		logs = pods.GetLogs(pod.Name, &v1.PodLogOptions{
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
	c.WaitForPod(namespace, name, 120*time.Second, v1.PodSucceeded)
	pod, err = pods.Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if pod.Status.Phase == v1.PodSucceeded {
		return nil
	}
	return fmt.Errorf("Pod did not finish successfully %s - %s", pod.Status.Phase, pod.Status.Message)
}

func CreateKubeConfig(clusterName string, ca certs.CertificateAuthority, endpoint string, group string, user string) ([]byte, error) {
	contextName := fmt.Sprintf("%s@%s", user, clusterName)
	cert := certs.NewCertificateBuilder(user).Organization(group).Client().Certificate
	cert, err := ca.SignCertificate(cert, 1)
	if err != nil {
		return nil, fmt.Errorf("createKubeConfig: failed to sign certificate: %v", err)
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

func CreateOIDCKubeConfig(clusterName string, ca certs.CertificateAuthority, endpoint, idpUrl, idToken, accessToken, refreshToken string) ([]byte, error) {
	if !strings.HasPrefix("https://", endpoint) {
		endpoint = "https://" + endpoint
	}

	if !strings.HasPrefix("https://", idpUrl) {
		idpUrl = "https://" + idpUrl
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
						"idp-certificate-authority-data": string(base64.StdEncoding.EncodeToString([]byte(ca.GetPublicChain()[0].EncodedCertificate()))),
						"idp-issuer-url":                 idpUrl,
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

// PingMaster attempts to connect to the API server and list nodes and services
// to ensure the API server is ready to accept any traffic
func (c *Client) PingMaster() bool {
	client, err := c.GetClientset()
	if err != nil {
		log.Tracef("pingMaster: Failed to get clientset: %v", err)
		return false
	}

	nodes, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if nodes == nil && len(nodes.Items) == 0 {
		return false
	}

	_, err = client.CoreV1().ServiceAccounts("kube-system").Get("default", metav1.GetOptions{})
	if err != nil {
		log.Tracef("pingMaster: Failed to get service account: %v", err)
		return false
	}
	return true
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
			return fmt.Errorf("Timeout exceeded waiting for %s is %s, error: %v", name, pod.Status.Phase, err)
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

// ExecutePodf runs the specified shell command inside a container of the specified pod
func (c *Client) ExecutePodf(namespace, pod, container string, command ...string) (string, string, error) {
	client, err := c.GetClientset()
	if err != nil {
		return "", "", fmt.Errorf("executePodf: Failed to get clientset: %v", err)
	}
	log.Debugf("[%s/%s/%s] %s", namespace, pod, container, command)
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

	log.Tracef("[%s/%s/%s] %s => %s %s ", namespace, pod, container, command, _stdout, _stderr)
	return _stdout, _stderr, nil
}

func safeString(buf *bytes.Buffer) string {
	if buf == nil || buf.Len() == 0 {
		return ""
	}
	return string(buf.Bytes())
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
	log.Tracef("[%s] executing '%s' in pod %s", node, command, pod.Name)
	if err != nil {
		return "", fmt.Errorf("executef: Failed to create pod: %v", err)
	}
	defer pods.Delete(pod.ObjectMeta.Name, &metav1.DeleteOptions{})

	logs := pods.GetLogs(pod.Name, &v1.PodLogOptions{
		Container: pod.Spec.Containers[0].Name,
	})

	err = c.WaitForPod("kube-system", pod.ObjectMeta.Name, timeout, v1.PodSucceeded)
	logString := read(logs)
	if err != nil {
		return logString, fmt.Errorf("failed to execute command, pod did not complete: %v", err)
	} else {
		log.Tracef("[%s] stdout: %s", node, logString)
		return logString, nil
	}
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
		Volumes: []v1.Volume{v1.Volume{
			Name: "root",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: "/",
				},
			},
		}},
		Containers: []v1.Container{v1.Container{
			Name:  "shell",
			Image: "docker.io/ubuntu:18.04",
			Command: []string{
				"sh",
				"-c",
				"chroot /chroot bash -c \"" + command + "\"",
			},
			VolumeMounts: []v1.VolumeMount{v1.VolumeMount{
				Name:      "root",
				MountPath: "/chroot",
			}},
			SecurityContext: &v1.SecurityContext{
				Privileged: &yes,
			},
		}},
		Tolerations: []v1.Toleration{
			v1.Toleration{
				// tolerate all values
				Operator: "Exists",
			},
		},
		HostNetwork: true,
		HostPID:     true,
		HostIPC:     true,
	}
}
