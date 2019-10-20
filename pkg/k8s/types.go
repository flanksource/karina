package k8s

type CRD struct {
	Kind       string                 `yaml:"kind,omitempty"`
	ApiVersion string                 `yaml:"apiVersion,omitempty"`
	Metadata   Metadata               `yaml:"metadata,omitempty"`
	Spec       map[string]interface{} `yaml:"spec,omitempty"`
}

type Metadata struct {
	Name        string            `yaml:"name,omitempty"`
	Namespace   string            `yaml:"namespace,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}
