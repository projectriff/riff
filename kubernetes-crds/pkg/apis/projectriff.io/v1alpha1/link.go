/*
 * Copyright 2018 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true

// Represents the links.projectriff.io CRD
type Link struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               LinkSpec    `json:"spec"`
	Status             *LinkStatus `json:"status,omitempty"`
}

// Spec (what the user wants) for a link.
type LinkSpec struct {

	// The name of the function to bind
	Function string `json:"function"`

	// The name of the topic the function is monitoring for input messages.
	// +optional
	Input string `json:"input,omitempty"`

	// The name of the topic the function is writing its results to.
	// +optional
	Output string `json:"output,omitempty"`

	// How to slice streams of incoming messages to the function.
	Windowing Windowing `json:"windowing"`
}

// Status (computed) for a link
type LinkStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Returned in list operations
type LinkList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []Link `json:"items"`
}

func SetDefaults_LinkSpec(obj *LinkSpec) {
}
