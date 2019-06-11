/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

import (
	knapis "github.com/knative/pkg/apis"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"github.com/knative/pkg/kmeta"
	"github.com/projectriff/system/pkg/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Processor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProcessorSpec   `json:"spec"`
	Status ProcessorStatus `json:"status"`
}

var (
	_ knapis.Validatable = (*Processor)(nil)
	_ knapis.Defaultable = (*Processor)(nil)
	_ kmeta.OwnerRefable = (*Processor)(nil)
	_ apis.Object        = (*Processor)(nil)
)

type ProcessorSpec struct {
	FunctionRef string   `json:"functionRef"`
	Inputs      []string `json:"inputs"`
	Outputs     []string `json:"outputs"`
}

type ProcessorStatus struct {
	duckv1alpha1.Status `json:",inline"`

	InputAddresses     []string `json:"inputAddresses,omitempty"`
	OutputAddresses    []string `json:"outputAddresses,omitempty"`
	OutputContentTypes []string `json:"outputContentTypes,omitempty"`
	DeploymentName     string   `json:"deploymentName,omitempty"`
	FunctionImage      string   `json:"functionImage,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ProcessorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Processor `json:"items"`
}

func (*Processor) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Processor")
}

func (p *Processor) GetStatus() apis.Status {
	return &p.Status
}
