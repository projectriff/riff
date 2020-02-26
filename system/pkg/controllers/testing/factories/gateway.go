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

	corev1 "k8s.io/api/core/v1"

	"github.com/projectriff/riff/system/pkg/apis"
	streamingv1alpha1 "github.com/projectriff/riff/system/pkg/apis/streaming/v1alpha1"
	rtesting "github.com/projectriff/riff/system/pkg/controllers/testing"
	"github.com/projectriff/riff/system/pkg/refs"
)

type gateway struct {
	target *streamingv1alpha1.Gateway
}

var (
	_ rtesting.Factory = (*gateway)(nil)
)

func Gateway(seed ...*streamingv1alpha1.Gateway) *gateway {
	var target *streamingv1alpha1.Gateway
	switch len(seed) {
	case 0:
		target = &streamingv1alpha1.Gateway{}
	case 1:
		target = seed[0]
	default:
		panic(fmt.Errorf("expected exactly zero or one seed, got %v", seed))
	}
	return &gateway{
		target: target,
	}
}

func (f *gateway) deepCopy() *gateway {
	return Gateway(f.target.DeepCopy())
}

func (f *gateway) Create() *streamingv1alpha1.Gateway {
	return f.deepCopy().target
}

func (f *gateway) CreateObject() apis.Object {
	return f.Create()
}

func (f *gateway) mutation(m func(*streamingv1alpha1.Gateway)) *gateway {
	f = f.deepCopy()
	m(f.target)
	return f
}

func (f *gateway) NamespaceName(namespace, name string) *gateway {
	return f.mutation(func(g *streamingv1alpha1.Gateway) {
		g.ObjectMeta.Namespace = namespace
		g.ObjectMeta.Name = name
	})
}

func (f *gateway) ObjectMeta(nf func(ObjectMeta)) *gateway {
	return f.mutation(func(g *streamingv1alpha1.Gateway) {
		omf := objectMeta(g.ObjectMeta)
		nf(omf)
		g.ObjectMeta = omf.Create()
	})
}

func (f *gateway) PodTemplateSpec(nf func(PodTemplateSpec)) *gateway {
	return f.mutation(func(g *streamingv1alpha1.Gateway) {
		if g.Spec.Template == nil {
			g.Spec.Template = &corev1.PodTemplateSpec{}
		}
		ptsf := podTemplateSpec(*g.Spec.Template)
		nf(ptsf)
		template := ptsf.Create()
		g.Spec.Template = &template
	})
}

func (f *gateway) Ports(ports ...corev1.ServicePort) *gateway {
	return f.mutation(func(g *streamingv1alpha1.Gateway) {
		g.Spec.Ports = ports
	})
}

func (f *gateway) StatusConditions(conditions ...*condition) *gateway {
	return f.mutation(func(g *streamingv1alpha1.Gateway) {
		c := make([]apis.Condition, len(conditions))
		for i, cg := range conditions {
			c[i] = cg.Create()
		}
		g.Status.Conditions = c
	})
}

func (f *gateway) StatusReady() *gateway {
	return f.StatusConditions(
		Condition().Type(streamingv1alpha1.GatewayConditionReady).True(),
	)
}

func (f *gateway) StatusObservedGeneration(generation int64) *gateway {
	return f.mutation(func(g *streamingv1alpha1.Gateway) {
		g.Status.ObservedGeneration = generation
	})
}

func (f *gateway) StatusDeploymentRef(format string, a ...interface{}) *gateway {
	return f.mutation(func(g *streamingv1alpha1.Gateway) {
		g.Status.DeploymentRef = &refs.TypedLocalObjectReference{
			APIGroup: rtesting.StringPtr("apps"),
			Kind:     "Deployment",
			Name:     fmt.Sprintf(format, a...),
		}
	})
}

func (f *gateway) StatusServiceRef(format string, a ...interface{}) *gateway {
	return f.mutation(func(g *streamingv1alpha1.Gateway) {
		g.Status.ServiceRef = &refs.TypedLocalObjectReference{
			Kind: "Service",
			Name: fmt.Sprintf(format, a...),
		}
	})
}

func (f *gateway) StatusAddress(format string, a ...interface{}) *gateway {
	return f.mutation(func(g *streamingv1alpha1.Gateway) {
		g.Status.Address = &apis.Addressable{
			URL: fmt.Sprintf(format, a...),
		}
	})
}
