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
)

const (
	InMemoryGatewayConditionReady                           = apis.ConditionReady
	InMemoryGatewayConditionGatewayReady apis.ConditionType = "GatewayReady"
)

var inmemoryGatewayCondSet = apis.NewLivingConditionSet(
	InMemoryGatewayConditionGatewayReady,
)

func (s *InMemoryGatewayStatus) GetObservedGeneration() int64 {
	return s.ObservedGeneration
}

func (s *InMemoryGatewayStatus) IsReady() bool {
	return inmemoryGatewayCondSet.Manage(s).IsHappy()
}

func (*InMemoryGatewayStatus) GetReadyConditionType() apis.ConditionType {
	return InMemoryGatewayConditionReady
}

func (s *InMemoryGatewayStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return inmemoryGatewayCondSet.Manage(s).GetCondition(t)
}

func (s *InMemoryGatewayStatus) InitializeConditions() {
	inmemoryGatewayCondSet.Manage(s).InitializeConditions()
}

func (s *InMemoryGatewayStatus) PropagateGatewayStatus(gs *GatewayStatus) {
	sc := gs.GetCondition(GatewayConditionReady)
	if sc == nil {
		return
	}
	switch {
	case sc.Status == corev1.ConditionUnknown:
		inmemoryGatewayCondSet.Manage(s).MarkUnknown(InMemoryGatewayConditionGatewayReady, sc.Reason, sc.Message)
	case sc.Status == corev1.ConditionTrue:
		inmemoryGatewayCondSet.Manage(s).MarkTrue(InMemoryGatewayConditionGatewayReady)
	case sc.Status == corev1.ConditionFalse:
		inmemoryGatewayCondSet.Manage(s).MarkFalse(InMemoryGatewayConditionGatewayReady, sc.Reason, sc.Message)
	}
}

func (s *InMemoryGatewayStatus) MarkGatewayNotOwned(name string) {
	inmemoryGatewayCondSet.Manage(s).MarkFalse(InMemoryGatewayConditionGatewayReady, "NotOwned", "There is an existing Gateway %q that the InMemoryGateway does not own.", name)
}
