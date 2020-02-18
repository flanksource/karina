package k8s

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"time"

	certs "github.com/flanksource/commons/certs"
	utils "github.com/flanksource/commons/utils"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/remotecommand"
)

type Client struct {
	GetKubeConfigBytes func() ([]byte, error)
	ApplyDryRun        bool
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

// GetDynamicClient creates a new k8s client
func (c *Client) GetDynamicClient() (dynamic.Interface, error) {
	data, err := c.GetKubeConfigBytes()
	if err != nil {
		return nil, fmt.Errorf("getDynamicClient: Failed to get k8s client: %v", err)
	}
	cfg, err := clientcmd.RESTConfigFromKubeConfig(data)
	if err != nil {
		return nil, fmt.Errorf("getDynamicClient: Failed to get REST config: %v", err)
	}
	return dynamic.NewForConfig(cfg)
}

func GetGVR(v interface{}) schema.GroupVersionResource {
	if reflect.ValueOf(v).Kind() == reflect.Ptr {
		return GetGVR(reflect.ValueOf(v).Elem().Interface())
	}
	typeOf := reflect.TypeOf(v)
	pkg := strings.Replace(typeOf.PkgPath(), "k8s.io/api/", "", 1)
	pkg = strings.Replace(pkg, "core/", "", -1)
	var group, version string
	if strings.Contains(pkg, "/") {
		group = strings.Split(pkg, "/")[0]
		version = strings.Split(pkg, "/")[1]
	} else {
		version = pkg
	}
	return schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: strings.ToLower(typeOf.Name()) + "s",
	}
}

func decodeStringToDuration(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}
	if t != reflect.TypeOf(metav1.Duration{time.Duration(5)}) {
		return data, nil
	}
	d, err := time.ParseDuration(data.(string))
	if err != nil {
		return data, fmt.Errorf("decodeStringToDuration: Failed to parse duration: %v", err)
	}
	return metav1.Duration{d}, nil
}

func decodeStringToTime(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}
	if t != reflect.TypeOf(metav1.Time{time.Now()}) {
		return data, nil
	}
	d, err := time.Parse(time.RFC3339, data.(string))
	if err != nil {
		return data, fmt.Errorf("decodeStringToTime: failed to decode to time: %v", err)
	}
	return metav1.Time{d}, nil
}

func (c *Client) Get(namespace string, name string, obj runtime.Object) error {
	client, _, _, err := c.GetDynamicClientFor(namespace, obj)
	unstructuredObj, err := client.Get(name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get: failed to get client: %v", err)
	}

	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		DecodeHook:       mapstructure.ComposeDecodeHookFunc(decodeStringToTime, decodeStringToDuration),
		Result:           &obj,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return fmt.Errorf("get: failed to decode config: %v", err)
	}
	return decoder.Decode(unstructuredObj.Object)

}
func (c *Client) GetDynamicClientFor(namespace string, obj runtime.Object) (dynamic.ResourceInterface, *schema.GroupVersionResource, *unstructured.Unstructured, error) {
	dynamicClient, err := c.GetDynamicClient()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("getDynamicClientFor: failed to get dynamic client: %v", err)
	}

	resource := schema.GroupVersionResource{
		Group:    obj.GetObjectKind().GroupVersionKind().Group,
		Version:  obj.GetObjectKind().GroupVersionKind().Version,
		Resource: strings.ToLower(obj.GetObjectKind().GroupVersionKind().Kind) + "s",
	}

	if resource.Group == "" {
		resource = GetGVR(obj)
	}

	convertedObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("getDynamicClientFor: failed to convert object: %v", err)
	}

	unstructuredObj := &unstructured.Unstructured{Object: convertedObj}

	if strings.HasPrefix(resource.Resource, "cluster") {
		return dynamicClient.Resource(resource), &resource, unstructuredObj, nil
	} else {
		if namespace == "" {
			namespace = unstructuredObj.GetNamespace()
		}
		return dynamicClient.Resource(resource).Namespace(namespace), &resource, unstructuredObj, nil
	}

}

func (c *Client) Apply(namespace string, objects ...runtime.Object) error {
	for _, obj := range objects {
		client, resource, unstructuredObj, err := c.GetDynamicClientFor(namespace, obj)
		if err != nil {
			return fmt.Errorf("apply: failed to get dynamic client: %v", err)
		}

		if log.IsLevelEnabled(log.TraceLevel) {
			data, _ := yaml.Marshal(unstructuredObj)
			log.Tracef("Applying resource: %s/%s/%s \n%s", resource.Group, resource.Version, resource.Resource, string(data))
		} else {
			log.Debugf("Applying resource: %s/%s/%s", resource.Group, resource.Version, resource.Resource)
		}

		if c.ApplyDryRun {
			log.Infof("[dry-run] %s/%s/%s created/configured", resource.Resource, unstructuredObj, unstructuredObj.GetName())
		} else {
			_, err := client.Create(unstructuredObj, metav1.CreateOptions{})
			if errors.IsAlreadyExists(err) {
				_, err = client.Update(unstructuredObj, metav1.UpdateOptions{})
				log.Infof("%s/%s/%s configured", resource.Resource, unstructuredObj.GetNamespace(), unstructuredObj.GetName())
			} else if err == nil {
				log.Infof("%s/%s/%s created", resource.Resource, unstructuredObj.GetNamespace(), unstructuredObj.GetName())
			}
			if err != nil {
				log.Errorf("error handling: %s/%s/%s : %v", resource.Group, resource.Version, resource.Resource, err)
			}
		}
	}
	return nil
}

// GetClientset creates a new k8s client
func (c *Client) GetClientset() (*kubernetes.Clientset, error) {
	cfg, err := c.GetRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("getClientset: failed to get REST config: %v", err)
	}
	return kubernetes.NewForConfigOrDie(cfg), nil
}

func (c *Client) GetRESTConfig() (*rest.Config, error) {
	data, err := c.GetKubeConfigBytes()
	if err != nil {
		return nil, fmt.Errorf("getRESTConfig: failed to get kubeconfig: %v", err)
	}
	return clientcmd.RESTConfigFromKubeConfig(data)
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
	ns := k8s.CoreV1().Namespaces()
	if _, err := ns.Get(name, metav1.GetOptions{}); errors.IsNotFound(err) {
		namespace := v1.Namespace{}
		namespace.Name = name
		namespace.Labels = labels
		namespace.Annotations = annotations
		if _, err := ns.Create(&namespace); err != nil {
			return fmt.Errorf("createOrUpdateNamespace: failed to namespace: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("createOrUpdateNamespace: failed to namespace: %v", err)
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
	cfg := api.Config{
		Clusters: map[string]*api.Cluster{
			clusterName: {
				Server:                "https://" + endpoint + ":6443",
				InsecureSkipTLSVerify: true,
			},
		},
		Contexts: map[string]*api.Context{
			clusterName: {
				Cluster:  clusterName,
				AuthInfo: "sso",
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			"sso": {
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

func (c *Client) WaitForPod(ns, name string, status v1.PodPhase, timeout time.Duration) error {
	client, err := c.GetClientset()
	if err != nil {
		return fmt.Errorf("waitForPod: Failed to get clientset: %v", err)
	}
	pods := client.CoreV1().Pods(ns)
	start := time.Now()
	for {
		pod, err := pods.Get(name, metav1.GetOptions{})
		if start.Add(timeout).Before(time.Now()) {
			return fmt.Errorf("Timeout exceeded waiting for %s to be %s: is %s, error: %v", name, status, pod.Status.Phase, err)
		}

		if pod != nil && pod.Status.Phase == status {
			return nil
		}
		time.Sleep(5 * time.Second)
	}

}

func (c *Client) ExecutePodf(namespace, pod, container, command string, args ...interface{}) (string, string, error) {
	client, err := c.GetClientset()
	if err != nil {
		return "", "", fmt.Errorf("executePodf: Failed to get clientset: %v", err)
	}
	const tty = false
	req := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
		SubResource("exec").
		Param("container", container)
	req.VersionedParams(&v1.PodExecOptions{
		Container: container,
		Command:   []string{fmt.Sprintf(command, args...)},
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
	var stdout, stderr *bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: stdout,
		Stderr: stderr,
		Tty:    tty,
	})

	return string(stdout.Bytes()), string(stderr.Bytes()), nil
}

// Execute runs the specified shell common on a node by creating
// a pre-scheduled pod that runs in the host namespace
func (c *Client) Executef(node string, timeout time.Duration, command string, args ...interface{}) (string, error) {
	client, err := c.GetClientset()
	if err != nil {
		return "", fmt.Errorf("executef: Failed to get clientset: %v", err)
	}
	pods := client.CoreV1().Pods("kube-system")

	pod, err := pods.Create(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("command-%s-%s", node, utils.ShortTimestamp()),
		},
		Spec: NewCommandJob(node, fmt.Sprintf(command, args...)),
	})
	if err != nil {
		return "", fmt.Errorf("executef: Failed to create pod: %v", err)
	}
	defer pods.Delete(pod.ObjectMeta.Name, &metav1.DeleteOptions{})

	logs := pods.GetLogs(pod.Name, &v1.PodLogOptions{
		Container: pod.Spec.Containers[0].Name,
	})

	err = c.WaitForPod("kube-system", pod.ObjectMeta.Name, v1.PodSucceeded, timeout)
	logString := read(logs)
	if err != nil {
		return logString, fmt.Errorf("failed to execute command, pod did not complete: %v", err)
	} else {
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
		HostNetwork: true,
		HostPID:     true,
		HostIPC:     true,
	}
}
