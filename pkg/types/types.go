package types

type PlatformConfig struct {
	Source         string
	Name           string     `yaml:"name,omitempty"`
	PodCidr        string     `yaml:"pod_cidr,omitempty"`
	SubnetCidr     string     `yaml:"subnet_cidr,omitempty"`
	DockerRegistry string     `yaml:"docker_registry,omitempty"`
	Domain         string     `yaml:"domain,omitempty"`
	LdapHost       string     `yaml:"ldap_host,omitempty"`
	LdapUser       string     `yaml:"ldap_user,omitempty"`
	LdapPass       string     `yaml:"ldap_pass,omitempty"`
	LdapDomain     string     `yaml:"ldap_domain,omitempty"`
	LdapAdminGroup string     `yaml:"ldap_admin_group,omitempty"`
	Specs          []string   `yaml:"specs,omitempty"`
	Policies       []string   `yaml:"policies,omitempty"`
	Monitoring     Monitoring `yaml:"monitoring,omitempty"`
	ELK            ELK        `yaml:"elk,omitempty"`
	Versions       Versions   `yaml:"versions,omitempty"`
}

type ObjectStorage struct {
	AccessKey string `yaml:"access_key,omitempty"`
	SecretKey string `yaml:"secret_key,omitempty"`
	Endpoint  string `yaml:"endpoint,omitempty"`
	Bucket    string `yaml:"bucket,omitempty"`
}
type Monitoring struct {
	Prometheus Prometheus `yaml:"prometheus,omitempty"`
	Grafana    Grafana    `yaml:"grafana,omitempty"`
}

type Prometheus struct {
	Version string `yaml:"version,omitempty"`
}

type Grafana struct {
	Version string `yaml:"version,omitempty"`
}

type ELK struct {
	Version      string `yaml:"version,omitempty"`
	Replicas     int    `yaml:"replicas,omitempty"`
	LogRetention string `yaml:"log_retention,omitempty"`
}

type Versions struct {
	Kubernetes       string            `yaml:"kubernetes,omitempty"`
	ContainerRuntime string            `yaml:"container_runtime,omitempty"`
	Dependencies     map[string]string `yaml:"dependencies,omitempty"`
}
