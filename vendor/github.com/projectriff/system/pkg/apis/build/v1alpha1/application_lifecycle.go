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
	ApplicationConditionReady                                      = duckv1alpha1.ConditionReady
	ApplicationConditionBuildCacheReady duckv1alpha1.ConditionType = "BuildCacheReady"
	ApplicationConditionBuildSucceeded  duckv1alpha1.ConditionType = "BuildSucceeded"
	ApplicationConditionImageResolved   duckv1alpha1.ConditionType = "ImageResolved"
)

var applicationCondSet = duckv1alpha1.NewLivingConditionSet(
	ApplicationConditionBuildCacheReady,
	ApplicationConditionBuildSucceeded,
	ApplicationConditionImageResolved,
)

func (as *ApplicationStatus) GetObservedGeneration() int64 {
	return as.ObservedGeneration
}

func (as *ApplicationStatus) IsReady() bool {
	return applicationCondSet.Manage(as).IsHappy()
}

func (*ApplicationStatus) GetReadyConditionType() duckv1alpha1.ConditionType {
	return ApplicationConditionReady
}

func (as *ApplicationStatus) GetCondition(t duckv1alpha1.ConditionType) *duckv1alpha1.Condition {
	return applicationCondSet.Manage(as).GetCondition(t)
}

func (as *ApplicationStatus) InitializeConditions() {
	applicationCondSet.Manage(as).InitializeConditions()
}

func (as *ApplicationStatus) MarkBuildCacheNotOwned(name string) {
	applicationCondSet.Manage(as).MarkFalse(ApplicationConditionBuildCacheReady, "NotOwned",
		fmt.Sprintf("There is an existing PersistentVolumeClaim %q that we do not own.", name))
}

func (as *ApplicationStatus) MarkBuildCacheNotUsed() {
	as.BuildCacheName = ""
	applicationCondSet.Manage(as).MarkTrue(ApplicationConditionBuildCacheReady)
}

func (as *ApplicationStatus) MarkBuildNotOwned(name string) {
	applicationCondSet.Manage(as).MarkFalse(ApplicationConditionBuildSucceeded, "NotOwned",
		fmt.Sprintf("There is an existing Build %q that we do not own.", name))
}

func (as *ApplicationStatus) MarkBuildNotUsed() {
	as.BuildName = ""
	applicationCondSet.Manage(as).MarkTrue(ApplicationConditionBuildSucceeded)
}

func (as *ApplicationStatus) MarkImageDefaultPrefixMissing(message string) {
	applicationCondSet.Manage(as).MarkFalse(ApplicationConditionImageResolved, "DefaultImagePrefixMissing", message)
}

func (as *ApplicationStatus) MarkImageInvalid(message string) {
	applicationCondSet.Manage(as).MarkFalse(ApplicationConditionImageResolved, "ImageInvalid", message)
}

func (as *ApplicationStatus) MarkImageMissing(message string) {
	applicationCondSet.Manage(as).MarkFalse(ApplicationConditionImageResolved, "ImageMissing", message)
}

func (as *ApplicationStatus) MarkImageResolved() {
	applicationCondSet.Manage(as).MarkTrue(ApplicationConditionImageResolved)
}

func (as *ApplicationStatus) PropagateBuildCacheStatus(pvcs *corev1.PersistentVolumeClaimStatus) {
	switch pvcs.Phase {
	case corev1.ClaimPending:
		// used for PersistentVolumeClaims that are not yet bound
		applicationCondSet.Manage(as).MarkUnknown(ApplicationConditionBuildCacheReady, string(pvcs.Phase), "volume claim is not yet bound")
	case corev1.ClaimBound:
		// used for PersistentVolumeClaims that are bound
		applicationCondSet.Manage(as).MarkTrue(ApplicationConditionBuildCacheReady)
	case corev1.ClaimLost:
		// used for PersistentVolumeClaims that lost their underlying PersistentVolume
		applicationCondSet.Manage(as).MarkFalse(ApplicationConditionBuildCacheReady, string(pvcs.Phase), "volume claim lost its underlying volume")
	}
}

func (as *ApplicationStatus) PropagateBuildStatus(bs *buildv1alpha1.BuildStatus) {
	sc := bs.GetCondition(buildv1alpha1.BuildSucceeded)
	if sc == nil {
		return
	}
	switch {
	case sc.Status == corev1.ConditionUnknown:
		applicationCondSet.Manage(as).MarkUnknown(ApplicationConditionBuildSucceeded, sc.Reason, sc.Message)
	case sc.Status == corev1.ConditionTrue:
		applicationCondSet.Manage(as).MarkTrue(ApplicationConditionBuildSucceeded)
	case sc.Status == corev1.ConditionFalse:
		applicationCondSet.Manage(as).MarkFalse(ApplicationConditionBuildSucceeded, sc.Reason, sc.Message)
	}
}
