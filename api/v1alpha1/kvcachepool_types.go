package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KVCachePoolSpec defines the desired state of KVCachePool
type KVCachePoolSpec struct {
	// TotalMemoryGB is the total memory (across all replicas) intended for KV cache.
	TotalMemoryGB int32 `json:"totalMemoryGB"`

	// Replicas is the desired number of cache nodes.
	// If omitted, defaults to 1.
	// +kubebuilder:default=1
	Replicas *int32 `json:"replicas,omitempty"`

	// Strategy is the cache strategy, e.g. "lru", "lfu", "rr".
	// Right now it's just metadata; later it will drive eviction/placement.
	// +kubebuilder:default="lru"
	Strategy string `json:"strategy,omitempty"`
}

// KVCachePoolStatus defines the observed state of KVCachePool.
type KVCachePoolStatus struct {
	// ReadyReplicas is how many cache nodes are currently ready.
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// KVCachePool is the Schema for the kvcachepools API
type KVCachePool struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of KVCachePool
	// +required
	Spec KVCachePoolSpec `json:"spec"`

	// status defines the observed state of KVCachePool
	// +optional
	Status KVCachePoolStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// KVCachePoolList contains a list of KVCachePool
type KVCachePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []KVCachePool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KVCachePool{}, &KVCachePoolList{})
}
