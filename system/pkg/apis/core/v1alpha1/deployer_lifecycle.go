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
	networkingv1beta1 "k8s.io/api/networking/v1beta1"

	apis "github.com/projectriff/riff/system/pkg/apis"
)

const (
	DeployerConditionReady                              = apis.ConditionReady
	DeployerConditionDeploymentReady apis.ConditionType = "DeploymentReady"
	DeployerConditionServiceReady    apis.ConditionType = "ServiceReady"
	DeployerConditionIngressReady    apis.ConditionType = "IngressReady"
)

var deployerCondSet = apis.NewLivingConditionSet(
	DeployerConditionDeploymentReady,
	DeployerConditionServiceReady,
	DeployerConditionIngressReady,
)

func (ds *DeployerStatus) GetObservedGeneration() int64 {
	return ds.ObservedGeneration
}

func (ds *DeployerStatus) IsReady() bool {
	return deployerCondSet.Manage(ds).IsHappy()
}

func (*DeployerStatus) GetReadyConditionType() apis.ConditionType {
	return DeployerConditionReady
}

func (ds *DeployerStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return deployerCondSet.Manage(ds).GetCondition(t)
}

func (ds *DeployerStatus) InitializeConditions() {
	deployerCondSet.Manage(ds).InitializeConditions()
}

func (ds *DeployerStatus) PropagateDeploymentStatus(cds *appsv1.DeploymentStatus) {
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
		// DeploymentAvailable is False while progressing, avoid reporting DeployerConditionReady as False
		deployerCondSet.Manage(ds).MarkUnknown(DeployerConditionDeploymentReady, progressing.Reason, progressing.Message)
		return
	}
	switch {
	case available.Status == corev1.ConditionUnknown:
		deployerCondSet.Manage(ds).MarkUnknown(DeployerConditionDeploymentReady, available.Reason, available.Message)
	case available.Status == corev1.ConditionTrue:
		deployerCondSet.Manage(ds).MarkTrue(DeployerConditionDeploymentReady)
	case available.Status == corev1.ConditionFalse:
		deployerCondSet.Manage(ds).MarkFalse(DeployerConditionDeploymentReady, available.Reason, available.Message)
	}
}

func (ds *DeployerStatus) PropagateServiceStatus(ss *corev1.ServiceStatus) {
	// services don't have meaningful status
	deployerCondSet.Manage(ds).MarkTrue(DeployerConditionServiceReady)
}

func (ds *DeployerStatus) MarkServiceNotOwned(name string) {
	deployerCondSet.Manage(ds).MarkFalse(DeployerConditionServiceReady, "NotOwned", "There is an existing Service %q that the Deployer does not own.", name)
}

// PropagateIngressStatus update DeployerConditionIngressReady condition
// in DeployerStatus according to IngressStatus.
func (ds *DeployerStatus) PropagateIngressStatus(is *networkingv1beta1.IngressStatus) {
	// ingress status is not set reliably
	deployerCondSet.Manage(ds).MarkTrue(DeployerConditionIngressReady)
}

func (ds *DeployerStatus) MarkIngressNotRequired() {
	deployerCondSet.Manage(ds).MarkTrue(DeployerConditionIngressReady)
}
