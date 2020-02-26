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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/projectriff/riff/system/pkg/apis"
	"github.com/projectriff/riff/system/pkg/refs"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

var (
	DeployerLabelKey = GroupVersion.Group + "/deployer"
)

var (
	_ apis.Resource = (*Deployer)(nil)
)

// DeployerSpec defines the desired state of Deployer
type DeployerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Build resolves the image from a build resource. As the target build
	// produces new images, they will be automatically rolled out to the
	// deployer.
	// +optional
	Build *Build `json:"build,omitempty"`

	// +optional
	ContainerConcurrency *int64 `json:"containerConcurrency,omitempty"`

	Scale Scale `json:"scale,omitempty"`

	// Template pod
	// +optional
	Template *corev1.PodTemplateSpec `json:"template,omitempty"`

	// IngressPolicy defines whether the workload should be reachable from
	// outside the cluster
	IngressPolicy IngressPolicy `json:"ingressPolicy,omitempty"`
}

// IngressPolicy describes whether the container should be exposed via
// ingress. Only one of the following ingress policies may be specified.
// If none of the following policies is specified, the default one is
// IngressPolicyClusterLocal.
type IngressPolicy string

const (
	IngressPolicyClusterLocal IngressPolicy = "ClusterLocal"
	IngressPolicyExternal     IngressPolicy = "External"
)

type Scale struct {
	Min *int32 `json:"min,omitempty"`
	Max *int32 `json:"max,omitempty"`
}

// DeployerStatus defines the observed state of Deployer
type DeployerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	apis.Status `json:",inline"`

	// LatestImage is the most recent image resolved from the build
	LatestImage string `json:"latestImage,omitempty"`

	// ConfigurationRef is a reference to the Knative Serving configuration
	// backing this deployer.
	ConfigurationRef *refs.TypedLocalObjectReference `json:"configurationRef,omitempty"`

	// RouteRef is a reference to the Knative Serving route backing this
	// deployer.
	RouteRef *refs.TypedLocalObjectReference `json:"routeRef,omitempty"`

	// Address to target this deployer internally
	Address *apis.Addressable `json:"address,omitempty"`

	// URL to target this deployer publicly
	URL string `json:"url,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="riff"
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.status.url`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].reason`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +genclient

// Deployer is the Schema for the deployers API
type Deployer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeployerSpec   `json:"spec,omitempty"`
	Status DeployerStatus `json:"status,omitempty"`
}

func (*Deployer) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Deployer")
}

func (d *Deployer) GetStatus() apis.ResourceStatus {
	return &d.Status
}

// +kubebuilder:object:root=true

// DeployerList contains a list of Deployer
type DeployerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Deployer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Deployer{}, &DeployerList{})
}
