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
	FunctionConditionReady                              = apis.ConditionReady
	FunctionConditionKpackImageReady apis.ConditionType = "KpackImageReady"
	FunctionConditionImageResolved   apis.ConditionType = "ImageResolved"
)

var functionCondSet = apis.NewLivingConditionSet(
	FunctionConditionKpackImageReady,
	FunctionConditionImageResolved,
)

func (fs *FunctionStatus) GetObservedGeneration() int64 {
	return fs.ObservedGeneration
}

func (fs *FunctionStatus) IsReady() bool {
	return functionCondSet.Manage(fs).IsHappy()
}

func (*FunctionStatus) GetReadyConditionType() apis.ConditionType {
	return FunctionConditionReady
}

func (fs *FunctionStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return functionCondSet.Manage(fs).GetCondition(t)
}

func (fs *FunctionStatus) InitializeConditions() {
	functionCondSet.Manage(fs).InitializeConditions()
}

func (fs *FunctionStatus) MarkBuildNotUsed() {
	fs.KpackImageRef = nil
	fs.BuildCacheRef = nil
	functionCondSet.Manage(fs).MarkTrue(FunctionConditionKpackImageReady)
}

func (fs *FunctionStatus) MarkImageDefaultPrefixMissing(message string) {
	functionCondSet.Manage(fs).MarkFalse(FunctionConditionImageResolved, "DefaultImagePrefixMissing", message)
}

func (fs *FunctionStatus) MarkImageInvalid(message string) {
	functionCondSet.Manage(fs).MarkFalse(FunctionConditionImageResolved, "ImageInvalid", message)
}

func (fs *FunctionStatus) MarkImageResolved() {
	functionCondSet.Manage(fs).MarkTrue(FunctionConditionImageResolved)
}

func (fs *FunctionStatus) PropagateKpackImageStatus(is *kpackbuildv1alpha1.ImageStatus) {
	sc := is.GetCondition(apis.ConditionReady)
	if sc == nil {
		return
	}
	switch {
	case sc.Status == corev1.ConditionUnknown:
		functionCondSet.Manage(fs).MarkUnknown(FunctionConditionKpackImageReady, sc.Reason, sc.Message)
	case sc.Status == corev1.ConditionTrue:
		functionCondSet.Manage(fs).MarkTrue(FunctionConditionKpackImageReady)
	case sc.Status == corev1.ConditionFalse:
		functionCondSet.Manage(fs).MarkFalse(FunctionConditionKpackImageReady, sc.Reason, sc.Message)
	}
}
