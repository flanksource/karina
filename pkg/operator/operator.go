package operator

import (
	"time"

	karinav1 "github.com/flanksource/karina/pkg/api/operator/v1"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	zapu "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// +kubebuilder:rbac:groups="*",resources="*",verbs="*"

type Operator struct {
	log logr.Logger
	mgr manager.Manager
}

type Config struct {
	MetricsAddr          string
	EnableLeaderElection bool
	SyncPeriod           time.Duration
	LogLevel             string
	Port                 int
}

func New(config Config) (*Operator, error) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = karinav1.AddToScheme(scheme)

	ctrl.SetLogger(zap.New(zap.UseDevMode(true), zap.Level(logLevelFromString(config.LogLevel))))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: config.MetricsAddr,
		Port:               config.Port,
		LeaderElection:     config.EnableLeaderElection,
		LeaderElectionID:   "bc12345d.flanksource.com",
		SyncPeriod:         &config.SyncPeriod,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to start manager")
	}

	if err = (NewKarinaConfigReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("controllers").WithName("KarinaConfig"),
		mgr.GetScheme(),
	)).SetupWithManager(mgr); err != nil {
		return nil, errors.Wrap(err, "failed to add KarinaConfigReconciler")
	}

	operator := &Operator{
		log: ctrl.Log.WithName("operator").WithName("KarinaOperator"),
		mgr: mgr,
	}
	return operator, nil
}

func (o *Operator) Run() error {
	// +kubebuilder:scaffold:builder

	o.log.Info("starting manager")
	if err := o.mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return errors.Wrap(err, "failed to start manager")
	}

	return nil
}

func logLevelFromString(logLevel string) *zapu.AtomicLevel {
	var level zapcore.Level

	switch logLevel {
	case "debug":
		level = zapu.DebugLevel
	case "info":
		level = zapu.InfoLevel
	case "error":
		level = zapu.ErrorLevel
	default:
		level = zapu.ErrorLevel
	}

	ll := zapu.NewAtomicLevelAt(level)
	return &ll
}
