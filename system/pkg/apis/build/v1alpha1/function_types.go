/*
Copyright 2019 the original author or authors.

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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	apis "github.com/projectriff/riff/system/pkg/apis"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

var (
	FunctionLabelKey = GroupVersion.Group + "/function"
)

var (
	_ apis.Resource = (*Function)(nil)
)

// FunctionSpec defines the desired state of Function
type FunctionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Image repository to push built images. May contain a leading underscore
	// to have the default image prefix applied, or be `_` to combine the default
	// image prefix with the resource's name as a default value.
	Image string `json:"image"`

	// CacheSize of persistent volume to store resources between builds
	CacheSize *resource.Quantity `json:"cacheSize,omitempty"`

	// Source location. Required for on cluster builds.
	Source *Source `json:"source,omitempty"`

	// +optional
	// +nullable
	FailedBuildHistoryLimit *int64 `json:"failedBuildHistoryLimit,omitempty"`
	// +optional
	// +nullable
	SuccessBuildHistoryLimit *int64 `json:"successBuildHistoryLimit,omitempty"`
	// +optional
	ImageTaggingStrategy ImageTaggingStrategy `json:"imageTaggingStrategy,omitempty"`
	// +optional
	Build ImageBuild `json:"build,omitempty"`

	// Artifact file containing the function within the build workspace.
	Artifact string `json:"artifact,omitempty"`

	// Handler name of the method or class to invoke. The value depends on the
	// invoker.
	Handler string `json:"handler,omitempty"`

	// Invoker language runtime name. Detected by default.
	Invoker string `json:"invoker,omitempty"`
}

// FunctionStatus defines the observed state of Function
type FunctionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	apis.Status `json:",inline"`
	BuildStatus `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="riff"
// +kubebuilder:printcolumn:name="Latest Image",type=string,JSONPath=`.status.latestImage`
// +kubebuilder:printcolumn:name="Artifact",type=string,JSONPath=`.spec.artifact`
// +kubebuilder:printcolumn:name="Handler",type=string,JSONPath=`.spec.handler`
// +kubebuilder:printcolumn:name="Invoker",type=string,JSONPath=`.spec.invoker`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].reason`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +genclient

// Function is the Schema for the functions API
type Function struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FunctionSpec   `json:"spec,omitempty"`
	Status FunctionStatus `json:"status,omitempty"`
}

func (f *Function) GetImage() string {
	return f.Spec.Image
}

func (*Function) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Function")
}

func (f *Function) GetStatus() apis.ResourceStatus {
	return &f.Status
}

// +kubebuilder:object:root=true

// FunctionList contains a list of Function
type FunctionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Function `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Function{}, &FunctionList{})
}
