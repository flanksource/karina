package operator

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/flanksource/commons/lookup"
	"github.com/flanksource/commons/utils"
	karinav1 "github.com/flanksource/karina/pkg/api/operator/v1"
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/phases/harbor"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
	"github.com/go-logr/logr"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"golang.org/x/sync/semaphore"
	"gopkg.in/flanksource/yaml.v3"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const secretMountDirectory = "/var/karina/config"

var requeuePeriod = 30 * time.Second

// KarinaConfigReconciler reconciles a KarinaConfig object
type KarinaConfigReconciler struct {
	client.Client
	log    logr.Logger
	Scheme *runtime.Scheme
	mtx    *sync.Mutex
	locks  map[string]*semaphore.Weighted
}

func NewKarinaConfigReconciler(k8s client.Client, log logr.Logger, scheme *runtime.Scheme) *KarinaConfigReconciler {
	reconciler := &KarinaConfigReconciler{
		Client: k8s,
		log:    log,
		Scheme: scheme,
		mtx:    &sync.Mutex{},
		locks:  map[string]*semaphore.Weighted{},
	}
	return reconciler
}

func (r *KarinaConfigReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.log.WithValues("KarinaConfig", req.NamespacedName)

	karinaConfig := &karinav1.KarinaConfig{}
	if err := r.Get(ctx, req.NamespacedName, karinaConfig); err != nil {
		if kerrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		log.Error(err, "failed to get KarinaConfig")
		return ctrl.Result{}, err
	}

	if karinaConfig.Status.PodStatus != nil {
		if err := r.updateDeployStatus(ctx, karinaConfig); err != nil {
			return ctrl.Result{}, err
		}

		if karinaConfig.Status.PodStatus == nil {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{RequeueAfter: requeuePeriod}, nil
	}

	yml, err := yaml.Marshal(karinaConfig.Spec)
	if err != nil {
		log.Error(err, "failed to marshal spec")
		return ctrl.Result{}, err
	}

	h := sha1.New()
	io.WriteString(h, string(yml)) // nolint: errcheck
	checksum := hex.EncodeToString(h.Sum(nil))

	log.Info("Current ", "checksum", checksum)
	log.Info("Last applied ", "checksum", karinaConfig.Status.LastAppliedChecksum)
	if karinaConfig.Status.LastAppliedChecksum == checksum {
		log.Info("Karina config already deployed", "checksum", checksum, "name", karinaConfig.Name, "namespace", karinaConfig.Namespace)
		return ctrl.Result{}, nil
	}

	r.mtx.Lock()
	namespacedName := namespacedName(req.Name, req.Namespace)
	lock, found := r.locks[namespacedName]
	if !found {
		lock = semaphore.NewWeighted(1)
		r.locks[namespacedName] = lock
	}
	r.mtx.Unlock()

	if !lock.TryAcquire(1) {
		err := errors.Errorf("deploy already in progress")
		log.Error(err, "deploy already in progress, skipping")
		return ctrl.Result{}, err
	}
	defer func() { lock.Release(1) }()

	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{Kind: "Secret", APIVersion: "v1"},
		Data:     map[string][]byte{},
	}

	platformConfig := karinaConfig.Spec.Config
	defaultConfig := types.DefaultPlatformConfig()
	addDefaults(&platformConfig)
	if err := r.addExtra(karinaConfig, &platformConfig, secret); err != nil {
		log.Error(err, "failed to add extra variables")
		return ctrl.Result{}, err
	}

	if err := mergo.Merge(&platformConfig, defaultConfig); err != nil {
		log.Error(err, "failed to merge default config")
		return ctrl.Result{}, err
	}

	platformConfig.DryRun = karinaConfig.Spec.DryRun

	platform := platform.Platform{
		PlatformConfig: platformConfig,
	}
	platform.InClusterConfig = true
	if err := platform.Init(); err != nil {
		log.Error(err, "failed to initialise platform")
		return ctrl.Result{}, err
	}

	if err := r.deploy(ctx, karinaConfig, &platform, secret); err != nil {
		log.Error(err, "failed to deploy config")
		return ctrl.Result{}, err
	}

	karinaConfig.Status.LastAppliedChecksum = checksum
	karinaConfig.Status.LastApplied = metav1.Now()
	if err := r.Status().Update(ctx, karinaConfig); err != nil {
		log.Error(err, "failed to update status")
		return ctrl.Result{RequeueAfter: requeuePeriod}, err
	}

	return ctrl.Result{}, nil
}

func (r *KarinaConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&karinav1.KarinaConfig{}).
		Complete(r)
}

func (r *KarinaConfigReconciler) deploy(ctx context.Context, karinaConfig *karinav1.KarinaConfig, p *platform.Platform, secret *v1.Secret) error {
	name := fmt.Sprintf("karina-deploy-%s-%s", karinaConfig.Name, utils.RandomKey(5))
	configYaml := p.String()
	configMap := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-config", name),
			Namespace: karinaConfig.Namespace,
			Labels: map[string]string{
				"karina.flanksource.com/config": karinaConfig.Name,
			},
		},
		Data: map[string]string{
			"config.yaml": configYaml,
		},
	}

	secret.ObjectMeta = metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-files", name),
		Namespace: karinaConfig.Namespace,
		Labels: map[string]string{
			"karina.flanksource.com/config": karinaConfig.Name,
		},
	}

	if err := r.Create(ctx, configMap); err != nil {
		return errors.Wrapf(err, "failed to create configmap %s", configMap.Name)
	}
	if err := r.Create(ctx, secret); err != nil {
		return errors.Wrapf(err, "failed to create secret %s", secret.Name)
	}

	image := karinaConfig.Spec.Image
	version := karinaConfig.Spec.Version
	if image == "" {
		if version == "" {
			version = "latest"
		}
		image = fmt.Sprintf("docker.io/flanksource/karina:%s", version)
	}

	args := []string{"deploy", "all", "-c", "/var/run/karina/config.yaml", "--in-cluster"}
	if karinaConfig.Spec.DryRun {
		args = append(args, "--dry-run")
	}

	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: karinaConfig.Namespace,
			Labels: map[string]string{
				"karina.flanksource.com/config": karinaConfig.Name,
			},
		},
		Spec: v1.PodSpec{
			ServiceAccountName: "karina-operator-manager",
			Containers: []v1.Container{
				{
					Name:            "karina",
					Image:           image,
					ImagePullPolicy: v1.PullIfNotPresent,
					Command:         []string{"karina"},
					Args:            args,
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "karina-files",
							MountPath: secretMountDirectory,
						},
						{
							Name:      "karina-config",
							MountPath: "/var/run/karina",
						},
					},
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "karina-files",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: secret.Name,
						},
					},
				},
				{
					Name: "karina-config",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{Name: configMap.Name},
						},
					},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
		},
	}

	karinaConfig.Status.PodName = pod.Name
	status := v1.PodUnknown
	karinaConfig.Status.PodStatus = &status

	if err := r.Create(ctx, pod); err != nil {
		r.cleanup(karinaConfig, pod.Name)
		return errors.Wrapf(err, "failed to create pod %s", pod.Name)
	}

	status = v1.PodPending
	karinaConfig.Status.PodStatus = &status

	return nil
}

func (r *KarinaConfigReconciler) cleanup(karinaConfig *karinav1.KarinaConfig, name string) {
	log := r.log.WithValues("KarinaConfig", ktypes.NamespacedName{Name: karinaConfig.Name, Namespace: karinaConfig.Namespace})
	ctx := context.Background()

	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-config", name),
			Namespace: karinaConfig.Namespace,
		},
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-files", name),
			Namespace: karinaConfig.Namespace,
		},
	}

	if err := r.Delete(ctx, configMap); err != nil {
		log.Error(err, "failed to delete configmap", "name", configMap.Name)
	}
	if err := r.Delete(ctx, secret); err != nil {
		log.Error(err, "failed to delete secret", "name", secret.Name)
	}
}

func (r *KarinaConfigReconciler) addExtra(karinaConfig *karinav1.KarinaConfig, p *types.PlatformConfig, secret *v1.Secret) error {
	log := r.log.WithValues("KarinaConfig", ktypes.NamespacedName{Name: karinaConfig.Name, Namespace: karinaConfig.Namespace})
	for key, templateFrom := range karinaConfig.Spec.TemplateFrom {
		var value string
		var err error

		if templateFrom.SecretKeyRef != nil {
			value, err = r.getSecretValue(templateFrom.SecretKeyRef.Name, karinaConfig.Namespace, templateFrom.SecretKeyRef.Key)
			if err != nil {
				return errors.Wrapf(err, "failed to get key %s in secret %s in namespace %s", templateFrom.SecretKeyRef.Key, templateFrom.SecretKeyRef.Name, karinaConfig.Namespace)
			}
		}

		if templateFrom.ConfigMapKeyRef != nil {
			value, err = r.getConfigmapValue(templateFrom.ConfigMapKeyRef.Name, karinaConfig.Namespace, templateFrom.ConfigMapKeyRef.Key)
			if err != nil {
				return errors.Wrapf(err, "failed to get key %s in secret %s in namespace %s", templateFrom.ConfigMapKeyRef.Key, templateFrom.ConfigMapKeyRef.Name, karinaConfig.Namespace)
			}
		}

		if templateFrom.Tmpfile {
			secret.Data[key] = []byte(value)
			value = path.Join(secretMountDirectory, key)
		}

		log.Info("Looking up key", "key", key)

		rvalue, err := lookup.LookupString(p, key)
		if err != nil {
			return errors.Wrapf(err, "cannot lookup key %s", key)
		}
		log.Info("Overriding", "key", key)
		switch rvalue.Interface().(type) {
		case string:
			rvalue.SetString(value)
		case int:
			i, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return errors.Wrapf(err, "cannot convert value for key %s to int", key)
			}
			rvalue.SetInt(i)
		case bool:
			b, err := strconv.ParseBool(value)
			if err != nil {
				return errors.Wrapf(err, "cannot convert value for key %s to a boolean", key)
			}
			rvalue.SetBool(b)
		}
	}

	return nil
}

func (r *KarinaConfigReconciler) updateDeployStatus(ctx context.Context, karinaConfig *karinav1.KarinaConfig) error {
	status := *karinaConfig.Status.PodStatus

	pod := &v1.Pod{}
	namespacedNamed := ktypes.NamespacedName{Name: karinaConfig.Status.PodName, Namespace: karinaConfig.Namespace}
	if err := r.Get(ctx, namespacedNamed, pod); err != nil {
		if kerrors.IsNotFound(err) {
			karinaConfig.Status.LastAppliedStatus = "error/missing-pod"
			karinaConfig.Status.PodName = ""
			karinaConfig.Status.PodStatus = nil
		} else {
			return errors.Wrapf(err, "failed to get pod %s", karinaConfig.Status.PodName)
		}
	} else {
		newStatus := pod.Status.Phase
		if newStatus == v1.PodSucceeded {
			karinaConfig.Status.LastAppliedStatus = "succeeded"
			karinaConfig.Status.PodName = ""
			karinaConfig.Status.PodStatus = nil
			r.cleanup(karinaConfig, karinaConfig.Status.PodName)
		} else if newStatus == v1.PodFailed {
			karinaConfig.Status.LastAppliedStatus = "failed"
			karinaConfig.Status.PodName = ""
			karinaConfig.Status.PodStatus = nil
			r.cleanup(karinaConfig, karinaConfig.Status.PodName)
		} else if newStatus != status {
			karinaConfig.Status.PodStatus = &newStatus
		} else {
			return nil
		}
	}

	if err := r.Status().Update(ctx, karinaConfig); err != nil {
		return errors.Wrapf(err, "failed to update status for karina config %s in namespace %s", karinaConfig.Name, karinaConfig.Namespace)
	}

	return nil
}

func (r *KarinaConfigReconciler) getSecretValue(name, namespace, key string) (string, error) {
	secret := &v1.Secret{}
	namespacedName := ktypes.NamespacedName{Name: name, Namespace: namespace}

	if err := r.Get(context.Background(), namespacedName, secret); err != nil {
		return "", errors.Wrapf(err, "failed to get secret %s in namespace %s", name, namespace)
	}

	data, found := secret.Data[key]
	if !found {
		return "", errors.Errorf("failed to find key %s in secret %s in namespace %s", key, name, namespace)
	}

	return string(data), nil
}

func (r *KarinaConfigReconciler) getConfigmapValue(name, namespace, key string) (string, error) {
	cm := &v1.ConfigMap{}
	namespacedName := ktypes.NamespacedName{Name: name, Namespace: namespace}

	if err := r.Get(context.Background(), namespacedName, cm); err != nil {
		return "", errors.Wrapf(err, "failed to get configmap %s in namespace %s", name, namespace)
	}

	data, found := cm.Data[key]
	if !found {
		return "", errors.Errorf("failed to find key %s in configmap %s in namespace %s", key, name, namespace)
	}

	return data, nil
}

func addDefaults(p *types.PlatformConfig) {
	ldap := p.Ldap
	if ldap.Port == "" {
		ldap.Port = "636"
	}
	if ldap != nil {
		p.Ldap = ldap
	}

	p.Gatekeeper.WhitelistNamespaces = append(p.Gatekeeper.WhitelistNamespaces, constants.PlatformNamespaces...)
	harbor.Defaults(p)
}

func namespacedName(name, namespace string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}
