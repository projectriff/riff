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
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const (
	HandlerConditionReady                                         = duckv1alpha1.ConditionReady
	HandlerConditionConfigurationReady duckv1alpha1.ConditionType = "ConfigurationReady"
	HandlerConditionRouteReady         duckv1alpha1.ConditionType = "RouteReady"
)

var handlerCondSet = duckv1alpha1.NewLivingConditionSet(
	HandlerConditionConfigurationReady,
	HandlerConditionRouteReady,
)

func (hs *HandlerStatus) GetObservedGeneration() int64 {
	return hs.ObservedGeneration
}

func (hs *HandlerStatus) IsReady() bool {
	return handlerCondSet.Manage(hs).IsHappy()
}

func (*HandlerStatus) GetReadyConditionType() duckv1alpha1.ConditionType {
	return HandlerConditionReady
}

func (hs *HandlerStatus) GetCondition(t duckv1alpha1.ConditionType) *duckv1alpha1.Condition {
	return handlerCondSet.Manage(hs).GetCondition(t)
}

func (hs *HandlerStatus) InitializeConditions() {
	handlerCondSet.Manage(hs).InitializeConditions()
}

func (hs *HandlerStatus) MarkConfigurationNotOwned(name string) {
	handlerCondSet.Manage(hs).MarkFalse(HandlerConditionConfigurationReady, "NotOwned",
		"There is an existing Configuration %q that we do not own.", name)
}

func (hs *HandlerStatus) PropagateConfigurationStatus(cs *servingv1alpha1.ConfigurationStatus) {
	sc := cs.GetCondition(servingv1alpha1.ConfigurationConditionReady)
	if sc == nil {
		return
	}
	switch {
	case sc.Status == corev1.ConditionUnknown:
		handlerCondSet.Manage(hs).MarkUnknown(HandlerConditionConfigurationReady, sc.Reason, sc.Message)
	case sc.Status == corev1.ConditionTrue:
		handlerCondSet.Manage(hs).MarkTrue(HandlerConditionConfigurationReady)
	case sc.Status == corev1.ConditionFalse:
		handlerCondSet.Manage(hs).MarkFalse(HandlerConditionConfigurationReady, sc.Reason, sc.Message)
	}
}

func (hs *HandlerStatus) MarkRouteNotOwned(name string) {
	handlerCondSet.Manage(hs).MarkFalse(HandlerConditionRouteReady, "NotOwned",
		"There is an existing Route %q that we do not own.", name)
}

func (hs *HandlerStatus) PropagateRouteStatus(rs *servingv1alpha1.RouteStatus) {
	hs.Address = rs.Address
	hs.URL = rs.URL

	sc := rs.GetCondition(servingv1alpha1.RouteConditionReady)
	if sc == nil {
		return
	}
	switch {
	case sc.Status == corev1.ConditionUnknown:
		handlerCondSet.Manage(hs).MarkUnknown(HandlerConditionRouteReady, sc.Reason, sc.Message)
	case sc.Status == corev1.ConditionTrue:
		handlerCondSet.Manage(hs).MarkTrue(HandlerConditionRouteReady)
	case sc.Status == corev1.ConditionFalse:
		handlerCondSet.Manage(hs).MarkFalse(HandlerConditionRouteReady, sc.Reason, sc.Message)
	}
}
