/*
Copyright 2019 the original author or authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"github.com/projectriff/system/pkg/apis"
)

const (
	StreamConditionReady                                = apis.ConditionReady
	StreamConditionResourceAvailable apis.ConditionType = "ResourceAvailable"
	StreamConditionBindingReady      apis.ConditionType = "BindingReady"
)

var streamCondSet = apis.NewLivingConditionSet(
	StreamConditionResourceAvailable,
	StreamConditionBindingReady,
)

func (ss *StreamStatus) GetObservedGeneration() int64 {
	return ss.ObservedGeneration
}

func (ss *StreamStatus) IsReady() bool {
	return streamCondSet.Manage(ss).IsHappy()
}

func (*StreamStatus) GetReadyConditionType() apis.ConditionType {
	return StreamConditionReady
}

func (ss *StreamStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return streamCondSet.Manage(ss).GetCondition(t)
}

func (ss *StreamStatus) InitializeConditions() {
	streamCondSet.Manage(ss).InitializeConditions()
}

func (ss *StreamStatus) MarkStreamProvisioned() {
	streamCondSet.Manage(ss).MarkTrue(StreamConditionResourceAvailable)
}

func (ss *StreamStatus) MarkStreamProvisionFailed(message string) {
	streamCondSet.Manage(ss).MarkFalse(StreamConditionResourceAvailable, "ProvisionFailed", message)
}

func (ss *StreamStatus) MarkBindingReady() {
	streamCondSet.Manage(ss).MarkTrue(StreamConditionBindingReady)
}

func (ss *StreamStatus) MarkBindingNotReady(message string, a ...interface{}) {
	streamCondSet.Manage(ss).MarkFalse(StreamConditionBindingReady, "BindingFailed", message, a...)
}
