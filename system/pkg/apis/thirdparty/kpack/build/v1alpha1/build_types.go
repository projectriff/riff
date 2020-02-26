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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/projectriff/riff/system/pkg/apis"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BuildSpec is the spec for a Build resource.
type BuildSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Tags           []string                    `json:"tags"`
	Builder        BuilderImage                `json:"builder"`
	ServiceAccount string                      `json:"serviceAccount"`
	Source         SourceConfig                `json:"source"`
	CacheName      string                      `json:"cacheName"`
	Env            []corev1.EnvVar             `json:"env"`
	Resources      corev1.ResourceRequirements `json:"resources"`
}

// BuildStatus is the status for a Build resource
type BuildStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	apis.Status    `json:",inline"`
	BuildMetadata  BuildpackMetadataList   `json:"buildMetadata"`
	LatestImage    string                  `json:"latestImage"`
	PodName        string                  `json:"podName"`
	StepStates     []corev1.ContainerState `json:"stepStates,omitempty"`
	StepsCompleted []string                `json:"stepsCompleted,omitempty"`
}

type BuilderImage struct {
	Image            string                        `json:"image"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,15,rep,name=imagePullSecrets"`
}

// +kubebuilder:object:root=true

type Build struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BuildSpec   `json:"spec,omitempty"`
	Status BuildStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BuildList contains a list of Build
type BuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Build `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Build{}, &BuildList{})
}
