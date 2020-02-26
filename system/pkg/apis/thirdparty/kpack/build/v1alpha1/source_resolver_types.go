/*
Copyright 2019 The original author or authors

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/projectriff/riff/system/pkg/apis"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SourceResolverSpec is the spec for a SourceResolver resource.
type SourceResolverSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ServiceAccount string       `json:"serviceAccount"`
	Source         SourceConfig `json:"source"`
}

// SourceResolverStatus is the status for a SourceResolver resource
type SourceResolverStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	apis.Status `json:",inline"`
	Source      ResolvedSourceConfig `json:"source"`
}

// +kubebuilder:object:root=true

type SourceResolver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SourceResolverSpec   `json:"spec,omitempty"`
	Status SourceResolverStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SourceResolverList contains a list of SourceResolver
type SourceResolverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SourceResolver `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SourceResolver{}, &SourceResolverList{})
}
