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

package factories

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/projectriff/system/pkg/apis"
	knativeservingv1 "github.com/projectriff/system/pkg/apis/thirdparty/knative/serving/v1"
	rtesting "github.com/projectriff/system/pkg/controllers/testing"
)

type knativeConfiguration struct {
	target *knativeservingv1.Configuration
}

var (
	_ rtesting.Factory = (*knativeConfiguration)(nil)
)

func KnativeConfiguration(seed ...*knativeservingv1.Configuration) *knativeConfiguration {
	var target *knativeservingv1.Configuration
	switch len(seed) {
	case 0:
		target = &knativeservingv1.Configuration{}
	case 1:
		target = seed[0]
	default:
		panic(fmt.Errorf("expected exactly zero or one seed, got %v", seed))
	}
	return &knativeConfiguration{
		target: target,
	}
}

func (f *knativeConfiguration) deepCopy() *knativeConfiguration {
	return KnativeConfiguration(f.target.DeepCopy())
}

func (f *knativeConfiguration) Create() *knativeservingv1.Configuration {
	return f.deepCopy().target
}

func (f *knativeConfiguration) CreateObject() apis.Object {
	return f.Create()
}

func (f *knativeConfiguration) mutation(m func(*knativeservingv1.Configuration)) *knativeConfiguration {
	f = f.deepCopy()
	m(f.target)
	return f
}

func (f *knativeConfiguration) NamespaceName(namespace, name string) *knativeConfiguration {
	return f.mutation(func(configuration *knativeservingv1.Configuration) {
		configuration.ObjectMeta.Namespace = namespace
		configuration.ObjectMeta.Name = name
	})
}

func (f *knativeConfiguration) ObjectMeta(nf func(ObjectMeta)) *knativeConfiguration {
	return f.mutation(func(configuration *knativeservingv1.Configuration) {
		omf := objectMeta(configuration.ObjectMeta)
		nf(omf)
		configuration.ObjectMeta = omf.Create()
	})
}

func (f *knativeConfiguration) PodTemplateSpec(nf func(PodTemplateSpec)) *knativeConfiguration {
	return f.mutation(func(configuration *knativeservingv1.Configuration) {
		ptsf := podTemplateSpec(
			// convert RevisionTemplateSpec into PodTemplateSpec
			corev1.PodTemplateSpec{
				ObjectMeta: configuration.Spec.Template.ObjectMeta,
				Spec:       configuration.Spec.Template.Spec.PodSpec,
			},
		)
		nf(ptsf)
		template := ptsf.Create()
		// update RevisionTemplateSpec with PodTemplateSpec managed fields
		configuration.Spec.Template.ObjectMeta = template.ObjectMeta
		configuration.Spec.Template.Spec.PodSpec = template.Spec
	})
}

func (f *knativeConfiguration) ContainerConcurrency(cc int64) *knativeConfiguration {
	return f.mutation(func(configuration *knativeservingv1.Configuration) {
		configuration.Spec.Template.Spec.ContainerConcurrency = &cc
	})
}

func (f *knativeConfiguration) UserContainer(cb func(*corev1.Container)) *knativeConfiguration {
	return f.PodTemplateSpec(func(pts PodTemplateSpec) {
		pts.ContainerNamed("user-container", cb)
	})
}

func (f *knativeConfiguration) StatusConditions(conditions ...*condition) *knativeConfiguration {
	return f.mutation(func(configuration *knativeservingv1.Configuration) {
		c := make([]apis.Condition, len(conditions))
		for i, cg := range conditions {
			c[i] = cg.Create()
		}
		configuration.Status.Conditions = c
	})
}

func (f *knativeConfiguration) StatusReady() *knativeConfiguration {
	return f.StatusConditions(
		Condition().Type(knativeservingv1.ConfigurationConditionReady).True(),
	)
}

func (f *knativeConfiguration) StatusObservedGeneration(generation int64) *knativeConfiguration {
	return f.mutation(func(configuration *knativeservingv1.Configuration) {
		configuration.Status.ObservedGeneration = generation
	})
}
