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

package factories

import (
	"fmt"

	"github.com/projectriff/system/pkg/apis"
	streamingv1alpha1 "github.com/projectriff/system/pkg/apis/streaming/v1alpha1"
	rtesting "github.com/projectriff/system/pkg/controllers/testing"
	"github.com/projectriff/system/pkg/refs"
)

type kafkaGateway struct {
	target *streamingv1alpha1.KafkaGateway
}

var (
	_ rtesting.Factory = (*kafkaGateway)(nil)
)

func KafkaGateway(seed ...*streamingv1alpha1.KafkaGateway) *kafkaGateway {
	var target *streamingv1alpha1.KafkaGateway
	switch len(seed) {
	case 0:
		target = &streamingv1alpha1.KafkaGateway{}
	case 1:
		target = seed[0]
	default:
		panic(fmt.Errorf("expected exactly zero or one seed, got %v", seed))
	}
	return &kafkaGateway{
		target: target,
	}
}

func (f *kafkaGateway) deepCopy() *kafkaGateway {
	return KafkaGateway(f.target.DeepCopy())
}

func (f *kafkaGateway) Create() *streamingv1alpha1.KafkaGateway {
	return f.deepCopy().target
}

func (f *kafkaGateway) CreateObject() apis.Object {
	return f.Create()
}

func (f *kafkaGateway) mutation(m func(*streamingv1alpha1.KafkaGateway)) *kafkaGateway {
	f = f.deepCopy()
	m(f.target)
	return f
}

func (f *kafkaGateway) NamespaceName(namespace, name string) *kafkaGateway {
	return f.mutation(func(g *streamingv1alpha1.KafkaGateway) {
		g.ObjectMeta.Namespace = namespace
		g.ObjectMeta.Name = name
	})
}

func (f *kafkaGateway) ObjectMeta(nf func(ObjectMeta)) *kafkaGateway {
	return f.mutation(func(g *streamingv1alpha1.KafkaGateway) {
		omf := objectMeta(g.ObjectMeta)
		nf(omf)
		g.ObjectMeta = omf.Create()
	})
}

func (f *kafkaGateway) BootstrapServers(bootstrapServers string) *kafkaGateway {
	return f.mutation(func(g *streamingv1alpha1.KafkaGateway) {
		g.Spec.BootstrapServers = bootstrapServers
	})
}

func (f *kafkaGateway) StatusConditions(conditions ...*condition) *kafkaGateway {
	return f.mutation(func(g *streamingv1alpha1.KafkaGateway) {
		c := make([]apis.Condition, len(conditions))
		for i, cg := range conditions {
			c[i] = cg.Create()
		}
		g.Status.Conditions = c
	})
}

func (f *kafkaGateway) StatusReady() *kafkaGateway {
	return f.StatusConditions(
		Condition().Type(streamingv1alpha1.KafkaGatewayConditionReady).True(),
	)
}

func (f *kafkaGateway) StatusObservedGeneration(generation int64) *kafkaGateway {
	return f.mutation(func(g *streamingv1alpha1.KafkaGateway) {
		g.Status.ObservedGeneration = generation
	})
}

func (f *kafkaGateway) StatusGatewayRef(format string, a ...interface{}) *kafkaGateway {
	return f.mutation(func(g *streamingv1alpha1.KafkaGateway) {
		g.Status.GatewayRef = &refs.TypedLocalObjectReference{
			APIGroup: rtesting.StringPtr("streaming.projectriff.io"),
			Kind:     "Gateway",
			Name:     fmt.Sprintf(format, a...),
		}
	})
}

func (f *kafkaGateway) StatusAddress(format string, a ...interface{}) *kafkaGateway {
	return f.mutation(func(g *streamingv1alpha1.KafkaGateway) {
		g.Status.Address = &apis.Addressable{
			URL: fmt.Sprintf(format, a...),
		}
	})
}

func (f *kafkaGateway) StatusGatewayImage(image string) *kafkaGateway {
	return f.mutation(func(g *streamingv1alpha1.KafkaGateway) {
		g.Status.GatewayImage = image
	})
}

func (f *kafkaGateway) StatusProvisionerImage(image string) *kafkaGateway {
	return f.mutation(func(g *streamingv1alpha1.KafkaGateway) {
		g.Status.ProvisionerImage = image
	})
}
