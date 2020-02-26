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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/projectriff/system/pkg/apis"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ImageSpec is the spec for a Image resource.
type ImageSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Tag                      string               `json:"tag"`
	Builder                  ImageBuilder         `json:"builder"`
	ServiceAccount           string               `json:"serviceAccount"`
	Source                   SourceConfig         `json:"source"`
	CacheSize                *resource.Quantity   `json:"cacheSize,omitempty"`
	FailedBuildHistoryLimit  *int64               `json:"failedBuildHistoryLimit"`
	SuccessBuildHistoryLimit *int64               `json:"successBuildHistoryLimit"`
	ImageTaggingStrategy     ImageTaggingStrategy `json:"imageTaggingStrategy"`
	Build                    ImageBuild           `json:"build"`
}

// ImageStatus is the status for a Image resource
type ImageStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	apis.Status    `json:",inline"`
	LatestBuildRef string `json:"latestBuildRef"`
	LatestImage    string `json:"latestImage"`
	BuildCounter   int64  `json:"buildCounter"`
	BuildCacheName string `json:"buildCacheName"`
}

type ImageBuilder struct {
	metav1.TypeMeta `json:",inline"`
	Name            string `json:"name"`
}

type ImageTaggingStrategy string

const (
	None        ImageTaggingStrategy = "None"
	BuildNumber ImageTaggingStrategy = "BuildNumber"
)

type ImageBuild struct {
	Env       []corev1.EnvVar             `json:"env,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// +kubebuilder:object:root=true

type Image struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImageSpec   `json:"spec,omitempty"`
	Status ImageStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ImageList contains a list of Image
type ImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Image `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Image{}, &ImageList{})
}
