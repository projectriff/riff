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

package v1alpha1

import (
	kapi "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true

// Represents the functions.projectriff.io CRD
type Function struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               FunctionSpec    `json:"spec"`
	Status             *FunctionStatus `json:"status,omitempty"`
}

// Spec (what the user wants) for a function.
type FunctionSpec struct {

	// Protocol used to communicate between the sidecar and the invoker (eg http, grpc).
	Protocol string `json:"protocol"`

	// The name of the topic the function is monitoring for input messages.
	// +optional
	Input string `json:"input,omitempty"`

	// The name of the topic the function is writing its results to.
	// +optional
	Output string `json:"output,omitempty"`

	// The maximum number of replicas to use. Defaults to the number of partitions of the input topic.
	// +optional
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`

	// How long to wait (in milliseconds) before scaling the function down from 1 to 0 replicas.
	// +optional
	IdleTimeoutMs *int32 `json:"idleTimeoutMs,omitempty"`

	// Container definition to use for the function.
	Container kapi.Container `json:"container"`

	// How to slice streams of incoming messages to the function.
	Windowing Windowing `json:"windowing"`
}

// Status (computed) for a function
type FunctionStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Returned in list operations
type FunctionList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []Function `json:"items"`
}

var defaultIdleTimeout = int32(10000)

func SetDefaults_FunctionSpec(obj *FunctionSpec) {
	if obj.IdleTimeoutMs == nil {
		obj.IdleTimeoutMs = &defaultIdleTimeout
	}
}
