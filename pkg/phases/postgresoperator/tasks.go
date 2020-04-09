package postgresoperator

import (
	"fmt"
	"strconv"
	"time"

	"github.com/moshloop/platform-cli/pkg/api/postgres"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/types"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

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
		db.Spec.Parameters = map[string]string{
			"archive_mode":    "on",
			"archive_timeout": "60s",
		}

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
		if config.Clone != nil {
			cloneEnvVars := cloneDatabaseEnv(p, config)
			envVarsList = append(envVarsList, cloneEnvVars...)
		}
		db.Spec.Env = envVarsList

		if err := p.Apply(ns, db); err != nil {
			return nil, err
		}
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
