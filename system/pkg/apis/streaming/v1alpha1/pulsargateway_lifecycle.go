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

	"github.com/projectriff/riff/system/pkg/apis"
)

const (
	PulsarGatewayConditionReady                           = apis.ConditionReady
	PulsarGatewayConditionGatewayReady apis.ConditionType = "GatewayReady"
)

var pulsarGatewayCondSet = apis.NewLivingConditionSet(
	PulsarGatewayConditionGatewayReady,
)

func (s *PulsarGatewayStatus) GetObservedGeneration() int64 {
	return s.ObservedGeneration
}

func (s *PulsarGatewayStatus) IsReady() bool {
	return pulsarGatewayCondSet.Manage(s).IsHappy()
}

func (*PulsarGatewayStatus) GetReadyConditionType() apis.ConditionType {
	return PulsarGatewayConditionReady
}

func (s *PulsarGatewayStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return pulsarGatewayCondSet.Manage(s).GetCondition(t)
}

func (s *PulsarGatewayStatus) InitializeConditions() {
	pulsarGatewayCondSet.Manage(s).InitializeConditions()
}

func (s *PulsarGatewayStatus) PropagateGatewayStatus(gs *GatewayStatus) {
	sc := gs.GetCondition(GatewayConditionReady)
	if sc == nil {
		return
	}
	switch {
	case sc.Status == corev1.ConditionUnknown:
		pulsarGatewayCondSet.Manage(s).MarkUnknown(PulsarGatewayConditionGatewayReady, sc.Reason, sc.Message)
	case sc.Status == corev1.ConditionTrue:
		pulsarGatewayCondSet.Manage(s).MarkTrue(PulsarGatewayConditionGatewayReady)
	case sc.Status == corev1.ConditionFalse:
		pulsarGatewayCondSet.Manage(s).MarkFalse(PulsarGatewayConditionGatewayReady, sc.Reason, sc.Message)
	}
}

func (s *PulsarGatewayStatus) MarkGatewayNotOwned(name string) {
	pulsarGatewayCondSet.Manage(s).MarkFalse(PulsarGatewayConditionGatewayReady, "NotOwned", "There is an existing Gateway %q that the PulsarGateway does not own.", name)
}
