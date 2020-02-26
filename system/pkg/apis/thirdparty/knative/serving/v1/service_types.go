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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	apis "github.com/projectriff/system/pkg/apis"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

var (
	_ apis.Resource = (*Service)(nil)
)

// ServiceSpec represents the configuration for the Service object.
// A Service's specification is the union of the specifications for a Route
// and Configuration.  The Service restricts what can be expressed in these
// fields, e.g. the Route must reference the provided Configuration;
// however, these limitations also enable friendlier defaulting,
// e.g. Route never needs a Configuration name, and may be defaulted to
// the appropriate "run latest" spec.
type ServiceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ServiceSpec inlines an unrestricted ConfigurationSpec.
	ConfigurationSpec `json:",inline"`

	// ServiceSpec inlines RouteSpec and restricts/defaults its fields
	// via webhook.  In particular, this spec can only reference this
	// Service's configuration and revisions (which also influences
	// defaults).
	RouteSpec `json:",inline"`
}

// ServiceStatus represents the Status stanza of the Service resource.
type ServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	apis.Status `json:",inline"`

	// In addition to inlining ConfigurationSpec, we also inline the fields
	// specific to ConfigurationStatus.
	ConfigurationStatusFields `json:",inline"`

	// In addition to inlining RouteSpec, we also inline the fields
	// specific to RouteStatus.
	RouteStatusFields `json:",inline"`
}

const (
	// ServiceConditionReady is set when the service is configured
	// and has available backends ready to receive traffic.
	ServiceConditionReady = apis.ConditionReady
)

func (ss *ServiceStatus) GetObservedGeneration() int64 {
	return ss.ObservedGeneration
}

func (ss *ServiceStatus) IsReady() bool {
	return ss.GetCondition(ss.GetReadyConditionType()).IsTrue()
}

func (*ServiceStatus) GetReadyConditionType() apis.ConditionType {
	return ServiceConditionReady
}

func (ss *ServiceStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return ss.Status.GetCondition(t)
}

// +kubebuilder:object:root=true

// Service acts as a top-level container that manages a Route and Configuration
// which implement a network service. Service exists to provide a singular
// abstraction which can be access controlled, reasoned about, and which
// encapsulates software lifecycle decisions such as rollout policy and
// team resource ownership. Service acts only as an orchestrator of the
// underlying Routes and Configurations (much as a kubernetes Deployment
// orchestrates ReplicaSets), and its usage is optional but recommended.
//
// The Service's controller will track the statuses of its owned Configuration
// and Route, reflecting their statuses and conditions as its own.
//
// See also: https://knative.dev/serving/blob/master/docs/spec/overview.md#service
type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceSpec   `json:"spec,omitempty"`
	Status ServiceStatus `json:"status,omitempty"`
}

func (*Service) GetGroupVersionKind() schema.GroupVersionKind {
	return GroupVersion.WithKind("Service")
}

func (s *Service) GetStatus() apis.ResourceStatus {
	return &s.Status
}

// +kubebuilder:object:root=true

// ServiceList contains a list of Service
type ServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Service `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Service{}, &ServiceList{})
}
