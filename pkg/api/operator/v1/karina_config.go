package v1

import (
	"github.com/flanksource/karina/pkg/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KarinaConfigSpec defines the desired state of KarinaConfig
type KarinaConfigSpec struct {
	Config       types.PlatformConfig        `json:"config,omitempty"`
	EnvFrom      map[string]v1.EnvVarSource  `json:"envFrom,omitmepty"`
	TemplateFrom map[string]v1.EnvFromSource `json:"templateFrom,omitempty"`
}

// KarinaConfigStatus defines the observed state of KarinaConfig
type KarinaConfigStatus struct {
	LastApplied metav1.Time `json:"lastApplied,omitempty"`
}

// +kubebuilder:object:root=true

// KarinaConfig is the Schema for the KarinaConfiges API
type KarinaConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KarinaConfigSpec   `json:"spec,omitempty"`
	Status KarinaConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KarinaConfigList contains a list of KarinaConfig
type KarinaConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KarinaConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KarinaConfig{}, &KarinaConfigList{})
}
