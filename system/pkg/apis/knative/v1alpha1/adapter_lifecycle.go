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
	apis "github.com/projectriff/system/pkg/apis"
)

const (
	AdapterConditionReady                          = apis.ConditionReady
	AdapterConditionBuildReady  apis.ConditionType = "BuildReady"
	AdapterConditionTargetFound apis.ConditionType = "TargetFound"
)

var adapterCondSet = apis.NewLivingConditionSet(
	AdapterConditionBuildReady,
	AdapterConditionTargetFound,
)

func (as *AdapterStatus) GetObservedGeneration() int64 {
	return as.ObservedGeneration
}

func (as *AdapterStatus) IsReady() bool {
	return adapterCondSet.Manage(as).IsHappy()
}

func (*AdapterStatus) GetReadyConditionType() apis.ConditionType {
	return AdapterConditionReady
}

func (as *AdapterStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return adapterCondSet.Manage(as).GetCondition(t)
}

func (as *AdapterStatus) InitializeConditions() {
	adapterCondSet.Manage(as).InitializeConditions()
}

func (as *AdapterStatus) MarkBuildNotFound(kind, name string) {
	adapterCondSet.Manage(as).MarkFalse(AdapterConditionBuildReady, "NotFound",
		"The %s %q was not found.", kind, name)
}

func (as *AdapterStatus) MarkBuildLatestImageMissing(kind, name string) {
	adapterCondSet.Manage(as).MarkFalse(AdapterConditionBuildReady, "LatestImageMissing",
		"The %s %q is missing the latest image.", kind, name)
}

func (as *AdapterStatus) MarkBuildReady() {
	adapterCondSet.Manage(as).MarkTrue(AdapterConditionBuildReady)
}

func (as *AdapterStatus) MarkTargetNotFound(kind, name string) {
	adapterCondSet.Manage(as).MarkFalse(AdapterConditionTargetFound, "NotFound",
		"The %s %q was not found.", kind, name)
}

func (as *AdapterStatus) MarkTargetInvalid(kind, name string, err error) {
	adapterCondSet.Manage(as).MarkFalse(AdapterConditionTargetFound, "Invalid",
		"The %s %q errored: %s", kind, name, err)
}

func (as *AdapterStatus) MarkTargetFound() {
	adapterCondSet.Manage(as).MarkTrue(AdapterConditionTargetFound)
}
