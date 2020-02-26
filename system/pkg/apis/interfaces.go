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

package apis

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Object interface {
	runtime.Object
	metav1.Object
}

type Resource interface {
	GetGroupVersionKind() schema.GroupVersionKind
	GetObjectMeta() metav1.Object
	GetStatus() ResourceStatus
}

type ResourceStatus interface {
	IsReady() bool
	GetCondition(t ConditionType) *Condition
	GetReadyConditionType() ConditionType
	GetObservedGeneration() int64
}
