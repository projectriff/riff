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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	kedav1alpha1 "github.com/projectriff/riff/system/pkg/apis/thirdparty/keda/v1alpha1"

	"github.com/projectriff/riff/system/pkg/apis"
)

const (
	ProcessorConditionReady                                = apis.ConditionReady
	ProcessorConditionStreamsReady      apis.ConditionType = "StreamsReady"
	ProcessorConditionDeploymentReady   apis.ConditionType = "DeploymentReady"
	ProcessorConditionScaledObjectReady apis.ConditionType = "ScaledObjectReady"
)

var processorCondSet = apis.NewLivingConditionSet(
	ProcessorConditionStreamsReady,
	ProcessorConditionDeploymentReady,
	ProcessorConditionScaledObjectReady,
)

func (ps *ProcessorStatus) GetObservedGeneration() int64 {
	return ps.ObservedGeneration
}

func (ps *ProcessorStatus) IsReady() bool {
	return processorCondSet.Manage(ps).IsHappy()
}

func (*ProcessorStatus) GetReadyConditionType() apis.ConditionType {
	return ProcessorConditionReady
}

func (ps *ProcessorStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return processorCondSet.Manage(ps).GetCondition(t)
}

func (ps *ProcessorStatus) InitializeConditions() {
	processorCondSet.Manage(ps).InitializeConditions()
}

func (ps *ProcessorStatus) MarkStreamsReady() {
	processorCondSet.Manage(ps).MarkTrue(ProcessorConditionStreamsReady)
}

func (ps *ProcessorStatus) MarkStreamsNotReady(message string) {
	processorCondSet.Manage(ps).MarkFalse(ProcessorConditionStreamsReady, "StreamNotReady", message)
}

func (ps *ProcessorStatus) PropagateDeploymentStatus(ds *appsv1.DeploymentStatus) {
	var available, progressing *appsv1.DeploymentCondition
	for i := range ds.Conditions {
		switch ds.Conditions[i].Type {
		case appsv1.DeploymentAvailable:
			available = &ds.Conditions[i]
		case appsv1.DeploymentProgressing:
			progressing = &ds.Conditions[i]
		}
	}
	if available == nil || progressing == nil {
		return
	}
	if progressing.Status == corev1.ConditionTrue && available.Status == corev1.ConditionFalse {
		// DeploymentAvailable is False while progressing, avoid reporting DeployerConditionReady as False
		processorCondSet.Manage(ps).MarkUnknown(ProcessorConditionDeploymentReady, progressing.Reason, progressing.Message)
		return
	}
	switch {
	case available.Status == corev1.ConditionUnknown:
		processorCondSet.Manage(ps).MarkUnknown(ProcessorConditionDeploymentReady, available.Reason, available.Message)
	case available.Status == corev1.ConditionTrue:
		processorCondSet.Manage(ps).MarkTrue(ProcessorConditionDeploymentReady)
	case available.Status == corev1.ConditionFalse:
		processorCondSet.Manage(ps).MarkFalse(ProcessorConditionDeploymentReady, available.Reason, available.Message)
	}
}

func (ps *ProcessorStatus) PropagateScaledObjectStatus(sos *kedav1alpha1.ScaledObjectStatus) {
	// TODO: ScaledObject does not report much atm
	processorCondSet.Manage(ps).MarkTrue(ProcessorConditionScaledObjectReady)
}
