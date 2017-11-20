/*
 * Copyright 2017 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kapi "k8s.io/kubernetes/pkg/api/v1"
)

// Represents the functions.extensions.sk8s.io CRD
type Function struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               FunctionSpec   `json:"spec"`
	Status             FunctionStatus `json:"status,omitempty"`
}

// Spec (what the user wants) for a function
type FunctionSpec struct {

	Protocol string `json:"protocol"`

	// +optional
	Input string `json:"input,omitempty"`

	// +optional
	Output string `json:"output,omitempty"`

	// +optional
	IdleTimeoutMs *int32 `json:"idleTimeoutMs,omitempty"`

	Container kapi.Container `json:"container"`

}

// Status (computed) for a function
type FunctionStatus struct {
}

// Returned in list operations
type FunctionList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []Function `json:"items"`
}
