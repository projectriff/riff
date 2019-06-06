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
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	ProcessorConditionReady = duckv1alpha1.ConditionReady
	// TODO add aggregated streams ready status
	ProcessorConditionFunctionReady   duckv1alpha1.ConditionType = "FunctionReady"
	ProcessorConditionDeploymentReady duckv1alpha1.ConditionType = "DeploymentReady"
)

var processorCondSet = duckv1alpha1.NewLivingConditionSet(
	ProcessorConditionFunctionReady,
	ProcessorConditionDeploymentReady,
)

func (ps *ProcessorStatus) GetObservedGeneration() int64 {
	return ps.ObservedGeneration
}

func (ps *ProcessorStatus) IsReady() bool {
	return processorCondSet.Manage(ps).IsHappy()
}

func (*ProcessorStatus) GetReadyConditionType() duckv1alpha1.ConditionType {
	return ProcessorConditionReady
}

func (ps *ProcessorStatus) GetCondition(t duckv1alpha1.ConditionType) *duckv1alpha1.Condition {
	return processorCondSet.Manage(ps).GetCondition(t)
}

func (ps *ProcessorStatus) InitializeConditions() {
	processorCondSet.Manage(ps).InitializeConditions()
}

func (ps *ProcessorStatus) MarkFunctionNotFound(name string) {
	processorCondSet.Manage(ps).MarkFalse(ProcessorConditionFunctionReady, "NotFound",
		"Unable to find function %q.", name)
}

func (ps *ProcessorStatus) PropagateFunctionStatus(fs *buildv1alpha1.FunctionStatus) {
	ps.FunctionImage = fs.LatestImage

	sc := fs.GetCondition(buildv1alpha1.FunctionConditionReady)
	if sc == nil {
		return
	}
	switch {
	case sc.Status == corev1.ConditionUnknown:
		processorCondSet.Manage(ps).MarkUnknown(ProcessorConditionFunctionReady, sc.Reason, sc.Message)
	case sc.Status == corev1.ConditionTrue:
		processorCondSet.Manage(ps).MarkTrue(ProcessorConditionFunctionReady)
	case sc.Status == corev1.ConditionFalse:
		processorCondSet.Manage(ps).MarkFalse(ProcessorConditionFunctionReady, sc.Reason, sc.Message)
	}
}

func (ps *ProcessorStatus) MarkDeploymentNotOwned(name string) {
	processorCondSet.Manage(ps).MarkFalse(ProcessorConditionDeploymentReady, "NotOwned",
		"There is an existing Deployment %q that we do not own.", name)
}

func (ps *ProcessorStatus) PropagateDeploymentStatus(ds *appsv1.DeploymentStatus) {
	var ac *appsv1.DeploymentCondition
	for _, c := range ds.Conditions {
		if c.Type == appsv1.DeploymentAvailable {
			ac = &c
			break
		}
	}
	if ac == nil {
		return
	}
	switch {
	case ac.Status == corev1.ConditionUnknown:
		processorCondSet.Manage(ps).MarkUnknown(ProcessorConditionDeploymentReady, ac.Reason, ac.Message)
	case ac.Status == corev1.ConditionTrue:
		processorCondSet.Manage(ps).MarkTrue(ProcessorConditionDeploymentReady)
	case ac.Status == corev1.ConditionFalse:
		processorCondSet.Manage(ps).MarkFalse(ProcessorConditionDeploymentReady, ac.Reason, ac.Message)
	}
}
