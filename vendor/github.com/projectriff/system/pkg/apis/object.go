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
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
)

type Object interface {
	GetStatus() Status
}

type Status interface {
	IsReady() bool
	GetCondition(t duckv1alpha1.ConditionType) *duckv1alpha1.Condition
	GetReadyConditionType() duckv1alpha1.ConditionType
	GetObservedGeneration() int64
}
