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
	KafkaGatewayConditionReady                           = apis.ConditionReady
	KafkaGatewayConditionGatewayReady apis.ConditionType = "GatewayReady"
)

var kafkaGatewayCondSet = apis.NewLivingConditionSet(
	KafkaGatewayConditionGatewayReady,
)

func (s *KafkaGatewayStatus) GetObservedGeneration() int64 {
	return s.ObservedGeneration
}

func (s *KafkaGatewayStatus) IsReady() bool {
	return kafkaGatewayCondSet.Manage(s).IsHappy()
}

func (*KafkaGatewayStatus) GetReadyConditionType() apis.ConditionType {
	return KafkaGatewayConditionReady
}

func (s *KafkaGatewayStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return kafkaGatewayCondSet.Manage(s).GetCondition(t)
}

func (s *KafkaGatewayStatus) InitializeConditions() {
	kafkaGatewayCondSet.Manage(s).InitializeConditions()
}

func (s *KafkaGatewayStatus) PropagateGatewayStatus(gs *GatewayStatus) {
	sc := gs.GetCondition(GatewayConditionReady)
	if sc == nil {
		return
	}
	switch {
	case sc.Status == corev1.ConditionUnknown:
		kafkaGatewayCondSet.Manage(s).MarkUnknown(KafkaGatewayConditionGatewayReady, sc.Reason, sc.Message)
	case sc.Status == corev1.ConditionTrue:
		kafkaGatewayCondSet.Manage(s).MarkTrue(KafkaGatewayConditionGatewayReady)
	case sc.Status == corev1.ConditionFalse:
		kafkaGatewayCondSet.Manage(s).MarkFalse(KafkaGatewayConditionGatewayReady, sc.Reason, sc.Message)
	}
}

func (s *KafkaGatewayStatus) MarkGatewayNotOwned(name string) {
	kafkaGatewayCondSet.Manage(s).MarkFalse(KafkaGatewayConditionGatewayReady, "NotOwned", "There is an existing Gateway %q that the KafkaGateway does not own.", name)
}
