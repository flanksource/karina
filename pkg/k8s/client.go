package k8s

import (
	"reflect"
	"strings"

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
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	GetKubeConfig func() (string, error)
	ApplyDryRun   bool
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
		log.Tracef("Failed tp get secret %s/%s: %v\n", namespace, name, err)
		return nil
	}
	return &cm.Data
}

// GetDynamicClient creates a new k8s client
func (c *Client) GetDynamicClient() (dynamic.Interface, error) {
	kubeconfig, err := c.GetKubeConfig()
	if err != nil {
		return nil, err
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
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

func getDynamicClient(dynamicClient dynamic.Interface, namespace string, obj runtime.Object) (dynamic.ResourceInterface, *schema.GroupVersionResource, *unstructured.Unstructured, error) {

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
		return nil, nil, nil, err
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
	dynamicClient, err := c.GetDynamicClient()
	if err != nil {
		return err
	}
	for _, obj := range objects {
		client, resource, unstructuredObj, err := getDynamicClient(dynamicClient, namespace, obj)
		if err != nil {
			return err
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
	kubeconfig, err := c.GetKubeConfig()
	if err != nil {
		return nil, err
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfigOrDie(cfg), nil
}

func (c *Client) Annotate(obj runtime.Object, annotations map[string]string) error {
	dynamicClient, err := c.GetDynamicClient()
	if err != nil {
		return err
	}
	client, resource, unstructuredObj, err := getDynamicClient(dynamicClient, "", obj)
	if err != nil {
		return err
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
	dynamicClient, err := c.GetDynamicClient()
	if err != nil {
		return err
	}
	client, resource, unstructuredObj, err := getDynamicClient(dynamicClient, "", obj)
	if err != nil {
		return err
	}
	existing := unstructuredObj.GetLabels()
	for k, v := range labels {
		existing[k] = v
	}
	unstructuredObj.SetLabels(existing)
	_, err = client.Update(unstructuredObj, metav1.UpdateOptions{})
	log.Infof("%s/%s/%s labelled", resource.Resource, unstructuredObj.GetNamespace(), unstructuredObj.GetName())
	return nil
}

func (c *Client) CreateOrUpdateNamespace(name string, labels map[string]string, annotations map[string]string) error {
	k8s, err := c.GetClientset()
	if err != nil {
		return err
	}
	ns := k8s.CoreV1().Namespaces()
	if _, err := ns.Get(name, metav1.GetOptions{}); errors.IsNotFound(err) {
		namespace := v1.Namespace{}
		namespace.Name = name
		namespace.Labels = labels
		namespace.Annotations = annotations
		if _, err := ns.Create(&namespace); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func (c *Client) HasSecret(ns, name string) bool {
	client, err := c.GetClientset()
	if err != nil {
		return false
	}
	secrets := client.CoreV1().Secrets(ns)
	cm, err := secrets.Get(name, metav1.GetOptions{})
	return cm != nil && err == nil

}

func (c *Client) HasConfigMap(ns, name string) bool {
	client, err := c.GetClientset()
	if err != nil {
		return false
	}
	configmaps := client.CoreV1().ConfigMaps(ns)
	cm, err := configmaps.Get(name, metav1.GetOptions{})
	return cm != nil && err == nil
}

func (c *Client) CreateOrUpdateSecret(name, ns string, data map[string][]byte) error {
	client, err := c.GetClientset()
	if err != nil {
		return err
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
				return err
			}
		}
	} else {
		(*cm).Data = data
		if !c.ApplyDryRun {
			log.Infof("Updating %s/secret/%s", ns, name)
			if _, err := secrets.Update(cm); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) CreateOrUpdateConfigMap(name, ns string, data map[string]string) error {
	client, err := c.GetClientset()
	if err != nil {
		return err
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
				return err
			}
		}
	} else {
		(*cm).Data = data
		if !c.ApplyDryRun {
			log.Infof("Updating %s/cm/%s", ns, name)
			if _, err := configs.Update(cm); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) ExposeIngress(namespace, service string, domain string, port int, annotations map[string]string) error {
	k8s, err := c.GetClientset()
	if err != nil {
		return err
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
				return err
			}
		}
	}
	return nil
}

func (c *Client) GetOrCreatePVC(namespace, name, size, class string) error {
	client, err := c.GetClientset()
	if err != nil {
		return err
	}
	qty, err := resource.ParseQuantity(size)
	if err != nil {
		return err
	}
	pvcs := client.CoreV1().PersistentVolumeClaims(namespace)

	existing, err := pvcs.Get(name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		log.Infof("Creating PVC %s/%s (%s %s)\n", namespace, name, size, class)
		_, err = pvcs.Create(&v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: v1.PersistentVolumeClaimSpec{
				StorageClassName: &class,
				AccessModes: []v1.PersistentVolumeAccessMode{
					v1.ReadWriteMany,
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
		return err
	} else {
		log.Infof("Found existing PVC %s/%s (%s %s) ==> %s\n", namespace, name, size, class, existing.UID)
		return nil
	}
	return err
}
