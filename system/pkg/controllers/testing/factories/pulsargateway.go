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

	"github.com/projectriff/riff/system/pkg/apis"
	streamingv1alpha1 "github.com/projectriff/riff/system/pkg/apis/streaming/v1alpha1"
	rtesting "github.com/projectriff/riff/system/pkg/controllers/testing"
	"github.com/projectriff/riff/system/pkg/refs"
)

type pulsarGateway struct {
	target *streamingv1alpha1.PulsarGateway
}

var (
	_ rtesting.Factory = (*pulsarGateway)(nil)
)

func PulsarGateway(seed ...*streamingv1alpha1.PulsarGateway) *pulsarGateway {
	var target *streamingv1alpha1.PulsarGateway
	switch len(seed) {
	case 0:
		target = &streamingv1alpha1.PulsarGateway{}
	case 1:
		target = seed[0]
	default:
		panic(fmt.Errorf("expected exactly zero or one seed, got %v", seed))
	}
	return &pulsarGateway{
		target: target,
	}
}

func (f *pulsarGateway) deepCopy() *pulsarGateway {
	return PulsarGateway(f.target.DeepCopy())
}

func (f *pulsarGateway) Create() *streamingv1alpha1.PulsarGateway {
	return f.deepCopy().target
}

func (f *pulsarGateway) CreateObject() apis.Object {
	return f.Create()
}

func (f *pulsarGateway) mutation(m func(*streamingv1alpha1.PulsarGateway)) *pulsarGateway {
	f = f.deepCopy()
	m(f.target)
	return f
}

func (f *pulsarGateway) NamespaceName(namespace, name string) *pulsarGateway {
	return f.mutation(func(g *streamingv1alpha1.PulsarGateway) {
		g.ObjectMeta.Namespace = namespace
		g.ObjectMeta.Name = name
	})
}

func (f *pulsarGateway) ObjectMeta(nf func(ObjectMeta)) *pulsarGateway {
	return f.mutation(func(g *streamingv1alpha1.PulsarGateway) {
		omf := objectMeta(g.ObjectMeta)
		nf(omf)
		g.ObjectMeta = omf.Create()
	})
}

func (f *pulsarGateway) ServiceURL(url string) *pulsarGateway {
	return f.mutation(func(g *streamingv1alpha1.PulsarGateway) {
		g.Spec.ServiceURL = url
	})
}

func (f *pulsarGateway) StatusConditions(conditions ...*condition) *pulsarGateway {
	return f.mutation(func(g *streamingv1alpha1.PulsarGateway) {
		c := make([]apis.Condition, len(conditions))
		for i, cg := range conditions {
			c[i] = cg.Create()
		}
		g.Status.Conditions = c
	})
}

func (f *pulsarGateway) StatusReady() *pulsarGateway {
	return f.StatusConditions(
		Condition().Type(streamingv1alpha1.PulsarGatewayConditionReady).True(),
	)
}

func (f *pulsarGateway) StatusObservedGeneration(generation int64) *pulsarGateway {
	return f.mutation(func(g *streamingv1alpha1.PulsarGateway) {
		g.Status.ObservedGeneration = generation
	})
}

func (f *pulsarGateway) StatusGatewayRef(format string, a ...interface{}) *pulsarGateway {
	return f.mutation(func(g *streamingv1alpha1.PulsarGateway) {
		g.Status.GatewayRef = &refs.TypedLocalObjectReference{
			APIGroup: rtesting.StringPtr("streaming.projectriff.io"),
			Kind:     "Gateway",
			Name:     fmt.Sprintf(format, a...),
		}
	})
}

func (f *pulsarGateway) StatusAddress(format string, a ...interface{}) *pulsarGateway {
	return f.mutation(func(g *streamingv1alpha1.PulsarGateway) {
		g.Status.Address = &apis.Addressable{
			URL: fmt.Sprintf(format, a...),
		}
	})
}

func (f *pulsarGateway) StatusGatewayImage(image string) *pulsarGateway {
	return f.mutation(func(g *streamingv1alpha1.PulsarGateway) {
		g.Status.GatewayImage = image
	})
}

func (f *pulsarGateway) StatusProvisionerImage(image string) *pulsarGateway {
	return f.mutation(func(g *streamingv1alpha1.PulsarGateway) {
		g.Status.ProvisionerImage = image
	})
}
