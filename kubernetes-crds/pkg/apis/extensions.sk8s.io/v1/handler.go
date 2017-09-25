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
)

// Represents the handlers.extensions.sk8s.io CRD
type Handler struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               HandlerSpec   `json:"spec"`
	Status             HandlerStatus `json:"status,omitempty"`
}

// Spec (what the user wants) for a handler
type HandlerSpec struct {
	// The dispatcher to use
	Dispatcher string `json:"dispatcher" description:"The Dispatcher strategy to use with this handler (name of a spring bean)"`

	// The container image to use
	Image string `json:"image"`

	// The command used to run the container
	// +optional
	Command string `json:"command,omitempty"`

	// The args used to run the container
	// +optional
	Args []string `json:"args,omitempty"`

	// The number of replicas to create.
	// Defaults to 1
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
}

// Status (computed) for a handler
type HandlerStatus struct {
}

// Returned in list operations
type HandlerList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []Handler `json:"items"`
}
