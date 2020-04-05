package postgres

type ClusterConfig struct {
	Name      string
	Databases []string
	Namespace string

	EnableWalClusterId   bool
	UseWalgRestore       bool
	BackupSchedule       string
	AwsCredentialsSecret string

	Clone *CloneConfig
}

type CloneConfig struct {
	ClusterName string
	ClusterID   string
	Timestamp   string
}

func NewClusterConfig(clusterName string, dbNames ...string) ClusterConfig {
	config := ClusterConfig{
		Name:                 clusterName,
		Databases:            dbNames,
		Namespace:            "postgres-operator",
		EnableWalClusterId:   false,
		BackupSchedule:       "*/5 * * * *",
		UseWalgRestore:       true,
		AwsCredentialsSecret: "postgres-operator-cluster-environment",
	}
	return config
}
