/*
Copyright 2019 The Knative Authors.

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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	apis "github.com/projectriff/system/pkg/apis"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

var (
	_ apis.Resource = (*Revision)(nil)
)

// RevisionSpec holds the desired state of the Revision (from the client).
type RevisionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	corev1.PodSpec `json:",inline"`

	// ContainerConcurrency specifies the maximum allowed in-flight (concurrent)
	// requests per container of the Revision.  Defaults to `0` which means
	// concurrency to the application is not limited, and the system decides the
	// target concurrency for the autoscaler.
	// +optional
	ContainerConcurrency *int64 `json:"containerConcurrency,omitempty"`

	// TimeoutSeconds holds the max duration the instance is allowed for
	// responding to a request.  If unspecified, a system default will
	// be provided.
	// +optional
	TimeoutSeconds *int64 `json:"timeoutSeconds,omitempty"`
}

// RevisionStatus communicates the observed state of the Revision (from the controller).
type RevisionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	apis.Status `json:",inline"`

	// ServiceName holds the name of a core Kubernetes Service resource that
	// load balances over the pods backing this Revision.
	// +optional
	ServiceName string `json:"serviceName,omitempty"`

	// LogURL specifies the generated logging url for this particular revision
	// based on the revision url template specified in the controller's config.
	// +optional
	LogURL string `json:"logUrl,omitempty"`

	// ImageDigest holds the resolved digest for the image specified
	// within .Spec.Container.Image. The digest is resolved during the creation
	// of Revision. This field holds the digest value regardless of whether
	// a tag or digest was originally specified in the Container object. It
	// may be empty if the image comes from a registry listed to skip resolution.
	// +optional
	ImageDigest string `json:"imageDigest,omitempty"`
}

const (
	// RevisionConditionReady is set when the revision is starting to materialize
	// runtime resources, and becomes true when those resources are ready.
	RevisionConditionReady = apis.ConditionReady
)

func (rs *RevisionStatus) GetObservedGeneration() int64 {
	return rs.ObservedGeneration
}

func (rs *RevisionStatus) IsReady() bool {
	return rs.GetCondition(rs.GetReadyConditionType()).IsTrue()
}

func (*RevisionStatus) GetReadyConditionType() apis.ConditionType {
	return RevisionConditionReady
}

func (rs *RevisionStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return rs.Status.GetCondition(t)
}

// RevisionTemplateSpec describes the data a revision should have when created from a template.
// Based on: https://github.com/kubernetes/api/blob/e771f807/core/v1/types.go#L3179-L3190
type RevisionTemplateSpec struct {
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec RevisionSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// Revision is an immutable snapshot of code and configuration.  A revision
// references a container image. Revisions are created by updates to a
// Configuration.
//
// See also: https://knative.dev/serving/blob/master/docs/spec/overview.md#revision
type Revision struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RevisionSpec   `json:"spec,omitempty"`
	Status RevisionStatus `json:"status,omitempty"`
}

func (*Revision) GetGroupVersionKind() schema.GroupVersionKind {
	return GroupVersion.WithKind("Revision")
}

func (r *Revision) GetStatus() apis.ResourceStatus {
	return &r.Status
}

// +kubebuilder:object:root=true

// RevisionList contains a list of Revision
type RevisionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Revision `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Revision{}, &RevisionList{})
}
