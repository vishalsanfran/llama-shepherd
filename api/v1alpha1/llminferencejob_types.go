package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LLMInferenceJobSpec defines the desired state of LLMInferenceJob
type LLMInferenceJobSpec struct {
	Prompt string `json:"prompt,omitempty"`
}

// LLMInferenceJobStatus defines the observed state of LLMInferenceJob.
type LLMInferenceJobStatus struct {
	Completed bool   `json:"completed,omitempty"`
	Output    string `json:"output,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// LLMInferenceJob is the Schema for the llminferencejobs API
type LLMInferenceJob struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of LLMInferenceJob
	// +required
	Spec LLMInferenceJobSpec `json:"spec"`

	// status defines the observed state of LLMInferenceJob
	// +optional
	Status LLMInferenceJobStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// LLMInferenceJobList contains a list of LLMInferenceJob
type LLMInferenceJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []LLMInferenceJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LLMInferenceJob{}, &LLMInferenceJobList{})
}
