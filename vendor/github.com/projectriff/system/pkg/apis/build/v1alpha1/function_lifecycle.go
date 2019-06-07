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
	"fmt"

	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const (
	FunctionConditionReady                                      = duckv1alpha1.ConditionReady
	FunctionConditionBuildCacheReady duckv1alpha1.ConditionType = "BuildCacheReady"
	FunctionConditionBuildSucceeded  duckv1alpha1.ConditionType = "BuildSucceeded"
	FunctionConditionImageResolved   duckv1alpha1.ConditionType = "ImageResolved"
)

var functionCondSet = duckv1alpha1.NewLivingConditionSet(
	FunctionConditionBuildCacheReady,
	FunctionConditionBuildSucceeded,
	FunctionConditionImageResolved,
)

func (fs *FunctionStatus) GetObservedGeneration() int64 {
	return fs.ObservedGeneration
}

func (fs *FunctionStatus) IsReady() bool {
	return functionCondSet.Manage(fs).IsHappy()
}

func (*FunctionStatus) GetReadyConditionType() duckv1alpha1.ConditionType {
	return FunctionConditionReady
}

func (fs *FunctionStatus) GetCondition(t duckv1alpha1.ConditionType) *duckv1alpha1.Condition {
	return functionCondSet.Manage(fs).GetCondition(t)
}

func (fs *FunctionStatus) InitializeConditions() {
	functionCondSet.Manage(fs).InitializeConditions()
}

func (fs *FunctionStatus) MarkBuildCacheNotOwned(name string) {
	functionCondSet.Manage(fs).MarkFalse(FunctionConditionBuildCacheReady, "NotOwned",
		fmt.Sprintf("There is an existing PersistentVolumeClaim %q that we do not own.", name))
}

func (fs *FunctionStatus) MarkBuildCacheNotUsed() {
	fs.BuildCacheName = ""
	functionCondSet.Manage(fs).MarkTrue(FunctionConditionBuildCacheReady)
}

func (fs *FunctionStatus) MarkBuildNotOwned(name string) {
	functionCondSet.Manage(fs).MarkFalse(FunctionConditionBuildSucceeded, "NotOwned",
		fmt.Sprintf("There is an existing Build %q that we do not own.", name))
}

func (fs *FunctionStatus) MarkBuildNotUsed() {
	fs.BuildName = ""
	functionCondSet.Manage(fs).MarkTrue(FunctionConditionBuildSucceeded)
}

func (fs *FunctionStatus) MarkImageDefaultPrefixMissing(message string) {
	functionCondSet.Manage(fs).MarkFalse(FunctionConditionImageResolved, "DefaultImagePrefixMissing", message)
}

func (fs *FunctionStatus) MarkImageInvalid(message string) {
	functionCondSet.Manage(fs).MarkFalse(FunctionConditionImageResolved, "ImageInvalid", message)
}

func (fs *FunctionStatus) MarkImageMissing(message string) {
	functionCondSet.Manage(fs).MarkFalse(FunctionConditionImageResolved, "ImageMissing", message)
}

func (fs *FunctionStatus) MarkImageResolved() {
	functionCondSet.Manage(fs).MarkTrue(FunctionConditionImageResolved)
}

func (fs *FunctionStatus) PropagateBuildCacheStatus(pvcs *corev1.PersistentVolumeClaimStatus) {
	switch pvcs.Phase {
	case corev1.ClaimPending:
		// used for PersistentVolumeClaims that are not yet bound
		functionCondSet.Manage(fs).MarkUnknown(FunctionConditionBuildCacheReady, string(pvcs.Phase), "volume claim is not yet bound")
	case corev1.ClaimBound:
		// used for PersistentVolumeClaims that are bound
		functionCondSet.Manage(fs).MarkTrue(FunctionConditionBuildCacheReady)
	case corev1.ClaimLost:
		// used for PersistentVolumeClaims that lost their underlying PersistentVolume
		functionCondSet.Manage(fs).MarkFalse(FunctionConditionBuildCacheReady, string(pvcs.Phase), "volume claim lost its underlying volume")
	}
}

func (fs *FunctionStatus) PropagateBuildStatus(bs *buildv1alpha1.BuildStatus) {
	sc := bs.GetCondition(buildv1alpha1.BuildSucceeded)
	if sc == nil {
		return
	}
	switch {
	case sc.Status == corev1.ConditionUnknown:
		functionCondSet.Manage(fs).MarkUnknown(FunctionConditionBuildSucceeded, sc.Reason, sc.Message)
	case sc.Status == corev1.ConditionTrue:
		functionCondSet.Manage(fs).MarkTrue(FunctionConditionBuildSucceeded)
	case sc.Status == corev1.ConditionFalse:
		functionCondSet.Manage(fs).MarkFalse(FunctionConditionBuildSucceeded, sc.Reason, sc.Message)
	}
}
