package postgresoperator

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/flanksource/karina/pkg/api/postgres"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
	"github.com/flanksource/kommons/proxy"
	"github.com/pkg/errors"

	pv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const exporterPort = 9187

func GetOrCreateDB(p *platform.Platform, config postgres.ClusterConfig) (*types.DB, error) {
	clusterName := "postgres-" + config.Name
	databases := make(map[string]string)
	appUsername := "app"
	ns := config.Namespace
	secretName := fmt.Sprintf("%s.%s.credentials", appUsername, clusterName)
	backupBucket := getBackupBucket(p)

	db := &postgres.Postgresql{}
	if err := p.Get(ns, clusterName, db); err != nil {
		log.Infof("Creating new cluster: %s", clusterName)
		for _, db := range config.Databases {
			databases[db] = appUsername
		}
		db = postgres.NewPostgresql(clusterName)
		db.Spec.Databases = databases
		db.Spec.Users = map[string]postgres.UserFlags{
			appUsername: {
				"createdb",
				"superuser",
			},
		}

		db.Spec.Sidecars = []postgres.Sidecar{
			patroniExporterSidecar(clusterName),
		}
		db.Spec.AdditionalVolumes = []postgres.AdditionalVolume{
			{
				Name:             "exporter-extra-queries",
				MountPath:        "/opt/extra-queries",
				TargetContainers: []string{"exporter"},
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "postgres-exporter-config",
						},
					},
				},
			},
		}

		envVarsList := []v1.EnvVar{}
		if config.EnableWalArchiving {
			db.Spec.Parameters = map[string]string{
				"archive_mode":    "on",
				"archive_timeout": "60s",
			}

			walEnvVars := getWalArchivingEnvVars(config, clusterName, backupBucket)
			envVarsList = append(envVarsList, walEnvVars...)
		}
		if config.Clone != nil {
			cloneEnvVars := cloneDatabaseEnv(p, config)
			envVarsList = append(envVarsList, cloneEnvVars...)
		}
		db.Spec.Env = envVarsList

		if err := p.Apply(ns, db); err != nil {
			return nil, err
		}
	}

	if err := createExporterServiceMonitor(p, clusterName, ns, exporterPort); err != nil {
		log.Errorf("Failed to create prometheus service monitor: %v", err)
	}

	doUntil(func() bool {
		if err := p.Get(ns, clusterName, db); err != nil {
			return true
		}
		log.Infof("Waiting for %s to be running, is: %s", clusterName, db.Status.PostgresClusterStatus)
		return db.Status.PostgresClusterStatus == postgres.ClusterStatusRunning
	})
	if db.Status.PostgresClusterStatus != postgres.ClusterStatusRunning {
		return nil, fmt.Errorf("postgres cluster failed to start: %s", db.Status.PostgresClusterStatus)
	}
	secret := p.GetSecret("postgres-operator", secretName)
	if secret == nil {
		return nil, fmt.Errorf("%s not found", secretName)
	}

	return &types.DB{
		Host:     fmt.Sprintf("%s.%s.svc.cluster.local", clusterName, ns),
		Username: string((*secret)["username"]),
		Port:     5432,
		Password: string((*secret)["password"]),
	}, nil
}

func cloneDatabaseEnv(p *platform.Platform, config postgres.ClusterConfig) []v1.EnvVar {
	waleS3Prefix := fmt.Sprintf("s3://%s/%s/wal", getBackupBucket(p), config.Clone.ClusterName)
	if config.EnableWalClusterID {
		waleS3Prefix = fmt.Sprintf("s3://%s/spilo/%s/%s/wal", getBackupBucket(p), config.Clone.ClusterName, config.Clone.ClusterID)
	}
	envVars := []v1.EnvVar{
		{
			Name:  "CLONE_METHOD",
			Value: "CLONE_WITH_WALE",
		},
		{
			Name:  "CLONE_USE_WALG_RESTORE",
			Value: strconv.FormatBool(config.UseWalgRestore),
		},
		{
			Name:  "CLONE_TARGET_TIME",
			Value: config.Clone.Timestamp,
		},
	}
	if !config.EnableWalClusterID {
		envVars = append(envVars, v1.EnvVar{
			Name:  "CLONE_WALG_S3_PREFIX",
			Value: waleS3Prefix,
		})
	}
	awsEnvVars := []string{
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_ENDPOINT",
		"AWS_S3_FORCE_PATH_STYLE",
	}
	for _, k := range awsEnvVars {
		envVar := v1.EnvVar{
			Name: fmt.Sprintf("CLONE_%s", k),
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{Name: config.AwsCredentialsSecret},
					Key:                  k,
				},
			},
		}
		envVars = append(envVars, envVar)
	}

	return envVars
}

func getWalArchivingEnvVars(config postgres.ClusterConfig, clusterName, backupBucket string) []v1.EnvVar {
	envVars := []string{
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_ENDPOINT",
		"AWS_S3_FORCE_PATH_STYLE",
	}
	envVarsList := []v1.EnvVar{
		{
			Name:  "BACKUP_SCHEDULE",
			Value: config.BackupSchedule,
		},
		{
			Name:  "USE_WALG_RESTORE",
			Value: strconv.FormatBool(config.UseWalgRestore),
		},
		{
			Name:  "USE_WALG_BACKUP",
			Value: strconv.FormatBool(config.UseWalgRestore),
		},
	}

	for _, k := range envVars {
		envVar := v1.EnvVar{
			Name: k,
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{Name: config.AwsCredentialsSecret},
					Key:                  k,
				},
			},
		}
		envVarsList = append(envVarsList, envVar)
	}
	if !config.EnableWalClusterID {
		envVarsList = append(envVarsList, v1.EnvVar{
			Name:  "WAL_BUCKET_SCOPE_SUFFIX",
			Value: "",
		})
		envVarsList = append(envVarsList, v1.EnvVar{
			Name:  "WALG_S3_PREFIX",
			Value: fmt.Sprintf("s3://%s/%s/wal/", backupBucket, clusterName),
		})
		envVarsList = append(envVarsList, v1.EnvVar{
			Name:  "CLONE_WAL_BUCKET_SCOPE_SUFFIX",
			Value: "/",
		})
	}
	return envVarsList
}

func getBackupBucket(p *platform.Platform) string {
	backupBucket := p.PostgresOperator.BackupBucket

	if backupBucket == "" {
		backupBucket = "postgres-backups-" + p.Name
	}

	return backupBucket
}

func doUntil(fn func() bool) bool {
	start := time.Now()

	for {
		if fn() {
			return true
		}
		if time.Now().After(start.Add(5 * time.Minute)) {
			return false
		}
		time.Sleep(5 * time.Second)
	}
}

func GetPatroniClient(p *platform.Platform, namespace, clusterName string) (*http.Client, error) {
	client, _ := p.GetClientset()
	opts := metav1.ListOptions{LabelSelector: fmt.Sprintf("cluster-name=%s,spilo-role=master", clusterName)}
	pods, err := client.CoreV1().Pods(namespace).List(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get master pod for cluster %s: %v", clusterName, err)
	}

	if len(pods.Items) != 1 {
		return nil, fmt.Errorf("expected 1 pod for spilo-role=master got %d", len(pods.Items))
	}

	dialer, err := p.GetProxyDialer(proxy.Proxy{
		Namespace:    namespace,
		Kind:         "pods",
		ResourceName: pods.Items[0].Name,
		Port:         8008,
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to get proxy dialer")
	}

	tr := &http.Transport{
		DialContext: dialer.DialContext,
	}

	httpClient := &http.Client{Transport: tr}

	return httpClient, nil
}

func createExporterServiceMonitor(p *platform.Platform, clusterName, ns string, port int32) error {
	service := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-exporter", clusterName),
			Namespace: ns,
			Labels: map[string]string{
				"application":  "spilo",
				"cluster-name": clusterName,
				"role":         "exporter",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "exporter",
					Port:       port,
					TargetPort: intstr.FromInt(int(port)),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"application":  "spilo",
				"cluster-name": clusterName,
			},
		},
	}
	if err := p.Apply(ns, service); err != nil {
		return errors.Wrap(err, "failed to create exporter service")
	}

	serviceMonitor := &pv1.ServiceMonitor{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "monitoring.coreos.com/v1",
			Kind:       "ServiceMonitor",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: ns,
		},
		Spec: pv1.ServiceMonitorSpec{
			Endpoints: []pv1.Endpoint{
				{
					Interval: "30s",
					Port:     "exporter",
					Path:     "/metrics",
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"application":  "spilo",
					"cluster-name": clusterName,
					"role":         "exporter",
				},
			},
		},
	}
	return p.Apply(ns, serviceMonitor)
}

func patroniExporterSidecar(clusterName string) postgres.Sidecar {
	sidecar := postgres.Sidecar{
		Name:        "exporter",
		DockerImage: "docker.io/bitnami/postgres-exporter:0.8.0",
		Ports: []v1.ContainerPort{
			{
				Name:          "exporter",
				ContainerPort: exporterPort,
				Protocol:      v1.ProtocolTCP,
			},
		},
		Resources: postgres.Resources{
			ResourceRequests: postgres.ResourceDescription{
				CPU:    "10m",
				Memory: "128Mi",
			},
			ResourceLimits: postgres.ResourceDescription{
				CPU:    "200m",
				Memory: "128Mi",
			},
		},
		Env: []v1.EnvVar{
			{
				Name:  "PG_EXPORTER_WEB_LISTEN_ADDRESS",
				Value: fmt.Sprintf(":%d", exporterPort),
			},
			{
				Name:  "DATA_SOURCE_URI",
				Value: "localhost?sslmode=disable",
			},
			{
				Name: "DATA_SOURCE_USER",
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: fmt.Sprintf("postgres.%s.credentials", clusterName),
						},
						Key: "username",
					},
				},
			},
			{
				Name: "DATA_SOURCE_PASS",
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: fmt.Sprintf("postgres.%s.credentials", clusterName),
						},
						Key: "password",
					},
				},
			},
			{
				Name:  "PG_EXPORTER_EXTEND_QUERY_PATH",
				Value: "/opt/extra-queries/queries.yaml",
			},
		},
	}
	return sidecar
}
