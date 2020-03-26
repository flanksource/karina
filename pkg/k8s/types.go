package k8s

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type CRD struct {
	Kind       string                 `yaml:"kind,omitempty"`
	APIVersion string                 `yaml:"apiVersion,omitempty"`
	Metadata   Metadata               `yaml:"metadata,omitempty"`
	Spec       map[string]interface{} `yaml:"spec,omitempty"`
}

type Metadata struct {
	Name        string            `yaml:"name,omitempty"`
	Namespace   string            `yaml:"namespace,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type DynamicKind struct {
	APIVersion, Kind string
}

func (dk DynamicKind) SetGroupVersionKind(gvk schema.GroupVersionKind) {}

func (dk DynamicKind) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(dk.APIVersion, dk.Kind)
}
