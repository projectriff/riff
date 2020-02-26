/*
Copyright 2020 the original author or authors.

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

	"github.com/projectriff/riff/system/pkg/apis"
)

const (
	GatewayConditionReady                              = apis.ConditionReady
	GatewayConditionDeploymentReady apis.ConditionType = "DeploymentReady"
	GatewayConditionServiceReady    apis.ConditionType = "ServiceReady"
)

var gatewayCondSet = apis.NewLivingConditionSet(
	GatewayConditionDeploymentReady,
	GatewayConditionServiceReady,
)

func (s *GatewayStatus) GetObservedGeneration() int64 {
	return s.ObservedGeneration
}

func (s *GatewayStatus) IsReady() bool {
	return gatewayCondSet.Manage(s).IsHappy()
}

func (*GatewayStatus) GetReadyConditionType() apis.ConditionType {
	return GatewayConditionReady
}

func (s *GatewayStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return gatewayCondSet.Manage(s).GetCondition(t)
}

func (s *GatewayStatus) InitializeConditions() {
	gatewayCondSet.Manage(s).InitializeConditions()
}

func (s *GatewayStatus) PropagateDeploymentStatus(cds *appsv1.DeploymentStatus) {
	var available, progressing *appsv1.DeploymentCondition
	for i := range cds.Conditions {
		switch cds.Conditions[i].Type {
		case appsv1.DeploymentAvailable:
			available = &cds.Conditions[i]
		case appsv1.DeploymentProgressing:
			progressing = &cds.Conditions[i]
		}
	}
	if available == nil || progressing == nil {
		return
	}
	if progressing.Status == corev1.ConditionTrue && available.Status == corev1.ConditionFalse {
		// DeploymentAvailable is False while progressing, avoid reporting GatewayConditionReady as False
		gatewayCondSet.Manage(s).MarkUnknown(GatewayConditionDeploymentReady, progressing.Reason, progressing.Message)
		return
	}
	switch {
	case available.Status == corev1.ConditionUnknown:
		gatewayCondSet.Manage(s).MarkUnknown(GatewayConditionDeploymentReady, available.Reason, available.Message)
	case available.Status == corev1.ConditionTrue:
		gatewayCondSet.Manage(s).MarkTrue(GatewayConditionDeploymentReady)
	case available.Status == corev1.ConditionFalse:
		gatewayCondSet.Manage(s).MarkFalse(GatewayConditionDeploymentReady, available.Reason, available.Message)
	}
}

func (s *GatewayStatus) PropagateServiceStatus(ss *corev1.ServiceStatus) {
	// services don't have meaningful status
	gatewayCondSet.Manage(s).MarkTrue(GatewayConditionServiceReady)
}
