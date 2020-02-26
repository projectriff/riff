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

type inmemoryGateway struct {
	target *streamingv1alpha1.InMemoryGateway
}

var (
	_ rtesting.Factory = (*inmemoryGateway)(nil)
)

func InMemoryGateway(seed ...*streamingv1alpha1.InMemoryGateway) *inmemoryGateway {
	var target *streamingv1alpha1.InMemoryGateway
	switch len(seed) {
	case 0:
		target = &streamingv1alpha1.InMemoryGateway{}
	case 1:
		target = seed[0]
	default:
		panic(fmt.Errorf("expected exactly zero or one seed, got %v", seed))
	}
	return &inmemoryGateway{
		target: target,
	}
}

func (f *inmemoryGateway) deepCopy() *inmemoryGateway {
	return InMemoryGateway(f.target.DeepCopy())
}

func (f *inmemoryGateway) Create() *streamingv1alpha1.InMemoryGateway {
	return f.deepCopy().target
}

func (f *inmemoryGateway) CreateObject() apis.Object {
	return f.Create()
}

func (f *inmemoryGateway) mutation(m func(*streamingv1alpha1.InMemoryGateway)) *inmemoryGateway {
	f = f.deepCopy()
	m(f.target)
	return f
}

func (f *inmemoryGateway) NamespaceName(namespace, name string) *inmemoryGateway {
	return f.mutation(func(g *streamingv1alpha1.InMemoryGateway) {
		g.ObjectMeta.Namespace = namespace
		g.ObjectMeta.Name = name
	})
}

func (f *inmemoryGateway) ObjectMeta(nf func(ObjectMeta)) *inmemoryGateway {
	return f.mutation(func(g *streamingv1alpha1.InMemoryGateway) {
		omf := objectMeta(g.ObjectMeta)
		nf(omf)
		g.ObjectMeta = omf.Create()
	})
}

func (f *inmemoryGateway) StatusConditions(conditions ...*condition) *inmemoryGateway {
	return f.mutation(func(g *streamingv1alpha1.InMemoryGateway) {
		c := make([]apis.Condition, len(conditions))
		for i, cg := range conditions {
			c[i] = cg.Create()
		}
		g.Status.Conditions = c
	})
}

func (f *inmemoryGateway) StatusReady() *inmemoryGateway {
	return f.StatusConditions(
		Condition().Type(streamingv1alpha1.InMemoryGatewayConditionReady).True(),
	)
}

func (f *inmemoryGateway) StatusObservedGeneration(generation int64) *inmemoryGateway {
	return f.mutation(func(g *streamingv1alpha1.InMemoryGateway) {
		g.Status.ObservedGeneration = generation
	})
}

func (f *inmemoryGateway) StatusGatewayRef(format string, a ...interface{}) *inmemoryGateway {
	return f.mutation(func(g *streamingv1alpha1.InMemoryGateway) {
		g.Status.GatewayRef = &refs.TypedLocalObjectReference{
			APIGroup: rtesting.StringPtr("streaming.projectriff.io"),
			Kind:     "Gateway",
			Name:     fmt.Sprintf(format, a...),
		}
	})
}

func (f *inmemoryGateway) StatusAddress(format string, a ...interface{}) *inmemoryGateway {
	return f.mutation(func(g *streamingv1alpha1.InMemoryGateway) {
		g.Status.Address = &apis.Addressable{
			URL: fmt.Sprintf(format, a...),
		}
	})
}

func (f *inmemoryGateway) StatusGatewayImage(image string) *inmemoryGateway {
	return f.mutation(func(g *streamingv1alpha1.InMemoryGateway) {
		g.Status.GatewayImage = image
	})
}

func (f *inmemoryGateway) StatusProvisionerImage(image string) *inmemoryGateway {
	return f.mutation(func(g *streamingv1alpha1.InMemoryGateway) {
		g.Status.ProvisionerImage = image
	})
}
