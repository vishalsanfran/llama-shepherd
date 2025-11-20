package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InferenceServiceSpec defines the desired state of InferenceService
type InferenceServiceSpec struct {

	// logical name of the model service routes to
	ModelRef string `json:"modelRef"`

	// number of router pods
	// +kubebuilder:default=1
	Replicas *int32 `json:"replicas,omitempty"`

	// maximum number of concurrent requests the router
	// should accept per pod
	// +kubebuilder:default=4
	MaxConcurrency int32 `json:"maxConcurrency,omitempty"`

	// CachePoolRef points to a KVCachePool the router should use.
	CachePoolRef string `json:"cachePoolRef,omitempty"`
}

// InferenceServiceStatus defines the observed state of InferenceService.
type InferenceServiceStatus struct {
	// how many router pods are actually ready.
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// InferenceService is the Schema for the inferenceservices API
type InferenceService struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of InferenceService
	// +required
	Spec InferenceServiceSpec `json:"spec"`

	// status defines the observed state of InferenceService
	// +optional
	Status InferenceServiceStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// InferenceServiceList contains a list of InferenceService
type InferenceServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []InferenceService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InferenceService{}, &InferenceServiceList{})
}
