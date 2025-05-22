// Package v1alpha1 contains API Schema definitions for the ANARETA operator
// +kubebuilder:object:generate=true
// +groupName=anareta.dev
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	scheme "sigs.k8s.io/controller-runtime/pkg/scheme"
)

// GroupVersion is group version used to register these objects
var GroupVersion = schema.GroupVersion{Group: "anareta.dev", Version: "v1alpha1"}

// SchemeBuilder is used to add go types to the GroupVersionKind scheme
var SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

// AddToScheme AddToGroupVersion registers the types with the given scheme
var AddToScheme = SchemeBuilder.AddToScheme

// AddToScheme adds all types to the scheme
func init() {
	SchemeBuilder.Register(&DevEnv{}, &DevEnvList{})
}
