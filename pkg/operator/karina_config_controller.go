package operator

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/flanksource/commons/lookup"
	karinav1 "github.com/flanksource/karina/pkg/api/operator/v1"
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/phases/harbor"
	"github.com/flanksource/karina/pkg/phases/order"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
	"github.com/go-logr/logr"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
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

// KarinaConfigReconciler reconciles a KarinaConfig object
type KarinaConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *KarinaConfigReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("KarinaConfig", req.NamespacedName)

	karinaConfig := &karinav1.KarinaConfig{}
	if err := r.Get(ctx, req.NamespacedName, karinaConfig); err != nil {
		if kerrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		log.Error(err, "failed to get KarinaConfig")
		return ctrl.Result{}, err
	}

	platformConfig := karinaConfig.Spec.Config
	defaultConfig := types.DefaultPlatformConfig()
	if err := r.addExtra(karinaConfig, &platformConfig); err != nil {
		log.Error(err, "failed to add extra variables")
		return ctrl.Result{}, err
	}

	if err := mergo.Merge(&platformConfig, defaultConfig); err != nil {
		log.Error(err, "failed to merge default config")
		return ctrl.Result{}, err
	}

	platformConfig.DryRun = karinaConfig.Spec.DryRun

	yml, err := yaml.Marshal(platformConfig)
	if err != nil {
		log.Error(err, "failed to marshal config")
		return ctrl.Result{}, err
	}

	h := sha1.New()
	io.WriteString(h, string(yml))
	checksum := string(h.Sum(nil))

	if karinaConfig.Status.LastAppliedChecksum != nil && *karinaConfig.Status.LastAppliedChecksum != checksum {
		log.Info("Karina config already deployed", "checksum", checksum, "name", karinaConfig.Name, "namespace", karinaConfig.Namespace)
		return ctrl.Result{}, nil
	}

	platform := platform.Platform{
		PlatformConfig: platformConfig,
	}
	platform.InClusterConfig = true
	if err := platform.Init(); err != nil {
		log.Error(err, "failed to initialise platform")
		return ctrl.Result{}, err
	}

	if err := r.deploy(karinaConfig, &platform); err != nil {
		log.Error(err, "failed to deploy config")
		return ctrl.Result{}, err
	}

	karinaConfig.Status.LastAppliedChecksum = &checksum
	karinaConfig.Status.LastApplied = metav1.Now()
	if err := r.Status().Update(ctx, karinaConfig); err != nil {
		log.Error(err, "failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *KarinaConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&karinav1.KarinaConfig{}).
		Complete(r)
}

func (r *KarinaConfigReconciler) deploy(karinaConfig *karinav1.KarinaConfig, p *platform.Platform) error {
	log := r.Log.WithValues("KarinaConfig", ktypes.NamespacedName{Name: karinaConfig.Name, Namespace: karinaConfig.Namespace})
	phases := order.GetPhases()
	// we track the failure status, and continue on failure to allow degraded operations
	failed := false

	errorMessage := ""

	// first deploy strictly ordered phases, these phases are often dependencies for other phases
	for _, name := range order.PhaseOrder {
		if err := phases[name](p); err != nil {
			log.Error(err, "failed to deploy", "phase", name)
			errorMessage += fmt.Sprintf("failed to deploy phase=%s error: %v\n", name, err)
			failed = true
		}
		// remove the phase from the map so it isn't run again
		delete(phases, name)
	}

	for name, fn := range phases {
		if err := fn(p); err != nil {
			log.Error(err, "failed to deploy", "phase", name)
			errorMessage += fmt.Sprintf("failed to deploy phase=%s error: %v\n", name, err)
			failed = true
		}
	}
	if failed {
		return errors.Errorf(errorMessage)
	}
	return nil
}

func (r *KarinaConfigReconciler) addExtra(karinaConfig *karinav1.KarinaConfig, p *types.PlatformConfig) error {
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
			tempfile, err := ioutil.TempFile("", "")
			if err != nil {
				return errors.Wrapf(err, "failed to create tempfile")
			}
			if err := ioutil.WriteFile(tempfile.Name(), []byte(value), 0600); err != nil {
				return errors.Wrapf(err, "failed to write tempfile for key %s", key)
			}

			if p.TmpFiles == nil {
				p.TmpFiles = []string{}
			}

			p.TmpFiles = append(p.TmpFiles, tempfile.Name())
			value = tempfile.Name()
		}

		r.Log.Info("Looking up key", "key", key)

		rvalue, err := lookup.LookupString(p, key)
		if err != nil {
			return errors.Wrapf(err, "cannot lookup key %s", key)
		}
		r.Log.Info("Overriding", "key", key)
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
