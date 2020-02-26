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
	corev1 "k8s.io/api/core/v1"

	"github.com/projectriff/system/pkg/apis"
	kpackbuildv1alpha1 "github.com/projectriff/system/pkg/apis/thirdparty/kpack/build/v1alpha1"
)

const (
	ApplicationConditionReady                              = apis.ConditionReady
	ApplicationConditionKpackImageReady apis.ConditionType = "KpackImageReady"
	ApplicationConditionImageResolved   apis.ConditionType = "ImageResolved"
)

var applicationCondSet = apis.NewLivingConditionSet(
	ApplicationConditionKpackImageReady,
	ApplicationConditionImageResolved,
)

func (as *ApplicationStatus) GetObservedGeneration() int64 {
	return as.ObservedGeneration
}

func (as *ApplicationStatus) IsReady() bool {
	return applicationCondSet.Manage(as).IsHappy()
}

func (*ApplicationStatus) GetReadyConditionType() apis.ConditionType {
	return ApplicationConditionReady
}

func (as *ApplicationStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return applicationCondSet.Manage(as).GetCondition(t)
}

func (as *ApplicationStatus) InitializeConditions() {
	applicationCondSet.Manage(as).InitializeConditions()
}

func (as *ApplicationStatus) MarkBuildNotUsed() {
	as.KpackImageRef = nil
	as.BuildCacheRef = nil
	applicationCondSet.Manage(as).MarkTrue(ApplicationConditionKpackImageReady)
}

func (as *ApplicationStatus) MarkImageDefaultPrefixMissing(message string) {
	applicationCondSet.Manage(as).MarkFalse(ApplicationConditionImageResolved, "DefaultImagePrefixMissing", message)
}

func (as *ApplicationStatus) MarkImageInvalid(message string) {
	applicationCondSet.Manage(as).MarkFalse(ApplicationConditionImageResolved, "ImageInvalid", message)
}

func (as *ApplicationStatus) MarkImageResolved() {
	applicationCondSet.Manage(as).MarkTrue(ApplicationConditionImageResolved)
}

func (as *ApplicationStatus) PropagateKpackImageStatus(is *kpackbuildv1alpha1.ImageStatus) {
	sc := is.GetCondition(apis.ConditionReady)
	if sc == nil {
		return
	}
	switch {
	case sc.Status == corev1.ConditionUnknown:
		applicationCondSet.Manage(as).MarkUnknown(ApplicationConditionKpackImageReady, sc.Reason, sc.Message)
	case sc.Status == corev1.ConditionTrue:
		applicationCondSet.Manage(as).MarkTrue(ApplicationConditionKpackImageReady)
	case sc.Status == corev1.ConditionFalse:
		applicationCondSet.Manage(as).MarkFalse(ApplicationConditionKpackImageReady, sc.Reason, sc.Message)
	}
}
