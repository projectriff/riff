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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/projectriff/riff/system/pkg/apis"
	"github.com/projectriff/riff/system/pkg/refs"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

var (
	PulsarGatewayLabelKey = GroupVersion.Group + "/pulsar-gateway"
	PulsarGatewayType     = "pulsar"
)

var (
	_ apis.Resource = (*PulsarGateway)(nil)
)

// PulsarGatewaySpec defines the desired state of PulsarGateway
type PulsarGatewaySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ServiceURL is the Pulsar URL to connect to, in the form pulsar://host:port[,host2:port2].
	ServiceURL string `json:"serviceURL"`
}

// PulsarGatewayStatus defines the observed state of PulsarGateway
type PulsarGatewayStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	apis.Status      `json:",inline"`
	Address          *apis.Addressable               `json:"address,omitempty"`
	GatewayRef       *refs.TypedLocalObjectReference `json:"gatewayRef,omitempty"`
	GatewayImage     string                          `json:"gatewayImage,omitempty"`
	ProvisionerImage string                          `json:"provisionerImage,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories="riff"
// +kubebuilder:printcolumn:name="Service URL",type=string,JSONPath=`.spec.serviceURL`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].reason`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +genclient

// PulsarGateway is the Schema for the gateways API
type PulsarGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PulsarGatewaySpec   `json:"spec,omitempty"`
	Status PulsarGatewayStatus `json:"status,omitempty"`
}

func (*PulsarGateway) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("PulsarGateway")
}

func (p *PulsarGateway) GetStatus() apis.ResourceStatus {
	return &p.Status
}

// +kubebuilder:object:root=true

// PulsarGatewayList contains a list of PulsarGateway
type PulsarGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PulsarGateway `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PulsarGateway{}, &PulsarGatewayList{})
}
