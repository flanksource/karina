/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package platformoperator

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterResourceQuotaSpec defines the desired state of ClusterResourceQuota
type ClusterResourceQuotaSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// Quota sets aggregate quota restrictions enforced across all namespaces
	Quota corev1.ResourceQuotaSpec `json:"quota,omitempty"`
}

// ClusterResourceQuotaStatus defines the observed state of ClusterResourceQuota
type ClusterResourceQuotaStatus struct {
	// Important: Run "make" to regenerate code after modifying this file

	// Total defines the actual enforced quota and its current usage across all namespaces
	Total corev1.ResourceQuotaStatus `json:"total,omitempty"`

	// Slices the quota used per namespace
	Namespaces ResourceQuotasStatusByNamespace `json:"namespaces"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// ClusterResourceQuota is the Schema for the clusterresourcequotas API
type ClusterResourceQuota struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired quota
	Spec ClusterResourceQuotaSpec `json:"spec,omitempty"`

	// Status defines the actual enforced quota and its current usage
	Status ClusterResourceQuotaStatus `json:"status,omitempty"`
}

// ResourceQuotasStatusByNamespace bundles multiple ResourceQuotaStatusByNamespace
type ResourceQuotasStatusByNamespace []ResourceQuotaStatusByNamespace

// ResourceQuotaStatusByNamespace gives status for a particular name
type ResourceQuotaStatusByNamespace struct {
	// Namespace the project this status applies to
	Namespace string `json:"namespace"`

	// Status indicates how many resources have been consumed by this project
	Status corev1.ResourceQuotaStatus `json:"status"`
}

// +kubebuilder:object:root=true

// ClusterResourceQuotaList contains a list of ClusterResourceQuota
type ClusterResourceQuotaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterResourceQuota `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterResourceQuota{}, &ClusterResourceQuotaList{})
}
