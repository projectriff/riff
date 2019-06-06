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
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
)

const (
	StreamConditionReady                                        = duckv1alpha1.ConditionReady
	StreamConditionResourceAvailable duckv1alpha1.ConditionType = "ResourceAvailable"
)

var streamCondSet = duckv1alpha1.NewLivingConditionSet(
	StreamConditionResourceAvailable,
)

func (ss *StreamStatus) GetObservedGeneration() int64 {
	return ss.ObservedGeneration
}

func (ss *StreamStatus) IsReady() bool {
	return streamCondSet.Manage(ss).IsHappy()
}

func (*StreamStatus) GetReadyConditionType() duckv1alpha1.ConditionType {
	return StreamConditionReady
}

func (ss *StreamStatus) GetCondition(t duckv1alpha1.ConditionType) *duckv1alpha1.Condition {
	return streamCondSet.Manage(ss).GetCondition(t)
}

func (ss *StreamStatus) InitializeConditions() {
	streamCondSet.Manage(ss).InitializeConditions()
}

func (ss *StreamStatus) MarkStreamProvisioned() {
	streamCondSet.Manage(ss).MarkTrue(StreamConditionReady)
}

func (ss *StreamStatus) MarkStreamProvisionFailed(message string) {
	streamCondSet.Manage(ss).MarkFalse(StreamConditionReady, "ProvisionFailed", message)
}
