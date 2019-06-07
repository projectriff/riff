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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Handler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HandlerSpec   `json:"spec"`
	Status HandlerStatus `json:"status"`
}

var (
	_ knapis.Validatable = (*Handler)(nil)
	_ knapis.Defaultable = (*Handler)(nil)
	_ kmeta.OwnerRefable = (*Handler)(nil)
	_ apis.Object        = (*Handler)(nil)
)

type HandlerSpec struct {
	Build    *Build          `json:"build,omitempty"`
	Template *corev1.PodSpec `json:"template,omitempty"`
}

type Build struct {
	ApplicationRef string `json:"applicationRef,omitempty"`
	FunctionRef    string `json:"functionRef,omitempty"`
}

type HandlerStatus struct {
	duckv1alpha1.Status `json:",inline"`

	ConfigurationName string `json:"configurationName,omitempty"`
	RouteName         string `json:"routeName,omitempty"`

	Address *duckv1alpha1.Addressable `json:"address,omitempty"`
	URL     *knapis.URL               `json:"url,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type HandlerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Handler `json:"items"`
}

func (*Handler) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Handler")
}

func (h *Handler) GetStatus() apis.Status {
	return &h.Status
}
