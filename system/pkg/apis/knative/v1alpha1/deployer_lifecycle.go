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

	apis "github.com/projectriff/riff/system/pkg/apis"
	servingv1 "github.com/projectriff/riff/system/pkg/apis/thirdparty/knative/serving/v1"
)

const (
	DeployerConditionReady                                 = apis.ConditionReady
	DeployerConditionConfigurationReady apis.ConditionType = "ConfigurationReady"
	DeployerConditionRouteReady         apis.ConditionType = "RouteReady"
)

var deployerCondSet = apis.NewLivingConditionSet(
	DeployerConditionConfigurationReady,
	DeployerConditionRouteReady,
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

func (ds *DeployerStatus) PropagateConfigurationStatus(kcs *servingv1.ConfigurationStatus) {
	sc := kcs.GetCondition(servingv1.ConfigurationConditionReady)
	if sc == nil {
		return
	}
	switch {
	case sc.Status == corev1.ConditionUnknown:
		deployerCondSet.Manage(ds).MarkUnknown(DeployerConditionConfigurationReady, sc.Reason, sc.Message)
	case sc.Status == corev1.ConditionTrue:
		deployerCondSet.Manage(ds).MarkTrue(DeployerConditionConfigurationReady)
	case sc.Status == corev1.ConditionFalse:
		deployerCondSet.Manage(ds).MarkFalse(DeployerConditionConfigurationReady, sc.Reason, sc.Message)
	}
}

func (ds *DeployerStatus) PropagateRouteStatus(rs *servingv1.RouteStatus) {
	ds.Address = rs.Address
	ds.URL = rs.URL

	sc := rs.GetCondition(servingv1.RouteConditionReady)
	if sc == nil {
		return
	}
	switch {
	case sc.Status == corev1.ConditionUnknown:
		deployerCondSet.Manage(ds).MarkUnknown(DeployerConditionRouteReady, sc.Reason, sc.Message)
	case sc.Status == corev1.ConditionTrue:
		deployerCondSet.Manage(ds).MarkTrue(DeployerConditionRouteReady)
	case sc.Status == corev1.ConditionFalse:
		deployerCondSet.Manage(ds).MarkFalse(DeployerConditionRouteReady, sc.Reason, sc.Message)
	}
}

func (ds *DeployerStatus) MarkRouteNotOwned(name string) {
	deployerCondSet.Manage(ds).MarkFalse(DeployerConditionRouteReady, "NotOwned", "There is an existing Route %q that the Deployer does not own.", name)
}
