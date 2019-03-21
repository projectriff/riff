/*
 * Copyright 2017 the original author or authors.
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

// Represents the invokers.projectriff.io CRD
type Invoker struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               InvokerSpec    `json:"spec"`
	Status             *InvokerStatus `json:"status,omitempty"`
}

// Spec (what the user wants) for an invoker.
type InvokerSpec struct {

	// Invoker version
	Version string `json:"version"`

	// File patterns to match this invoker with for an artifact
	Matchers []string `json:"matchers"`

	// Default function properties
	FunctionTemplate Function `json:"functionTemplate,omitempty"`

	// Default topic properties
	TopicTemplate Topic `json:"topicTemplate,omitempty"`

	// Default link properties
	LinkTemplate Link `json:"linkTemplate,omitempty"`

	// Handler function, if needed
	Handler InvokerHandler `json:"handler,omitempty"`

	// Files to generate for this invoker
	Files []InvokerFile `json:"files"`

	// Longform documentation for the `riff init` and `riff create` commands
	Doc string `json:"doc,omitempty"`
}

// Handler (handler function) for an invoker
type InvokerHandler struct {

	// Default handler function name
	Default string `json:"default"`

	// CLI help description for the handler
	Description string `json:"description"`
}

// File (generated files) for an invoker
type InvokerFile struct {

	// Path to generated file
	Path string `json:"path"`

	// File content to generate
	Template string `json:"template"`
}

// Status (computed) for an invoker
type InvokerStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Returned in list operations
type InvokerList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []Invoker `json:"items"`
}

func SetDefaults_InvokerSpec(obj *InvokerSpec) {
	if obj.Version == "" {
		obj.Version = "latest"
	}
}
