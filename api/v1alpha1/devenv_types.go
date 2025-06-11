package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=devenv;devs
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type DevEnv struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DevEnvSpec   `json:"spec,omitempty"`
	Status DevEnvStatus `json:"status,omitempty"`
}

// DevEnvSpec defines the spec for DevEnv
type DevEnvSpec struct {
	RepoURL string          `json:"repoURL"`
	Branch  string          `json:"branch"`
	TTL     metav1.Duration `json:"ttl,omitempty"`
}

// DevEnvStatus defines the observed state of DevEnv
type DevEnvStatus struct {
	Phase     string      `json:"phase,omitempty"`
	Message   string      `json:"message,omitempty"`
	StartedAt metav1.Time `json:"startedAt,omitempty"`
}

// +kubebuilder:object:root=true
type DevEnvList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DevEnv `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DevEnv{}, &DevEnvList{})
}
