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

	"github.com/projectriff/riff/system/pkg/apis"
	knativev1alpha1 "github.com/projectriff/riff/system/pkg/apis/knative/v1alpha1"
	rtesting "github.com/projectriff/riff/system/pkg/controllers/testing"
	"github.com/projectriff/riff/system/pkg/refs"
)

type deployerKnative struct {
	target *knativev1alpha1.Deployer
}

var (
	_ rtesting.Factory = (*deployerKnative)(nil)
)

func DeployerKnative(seed ...*knativev1alpha1.Deployer) *deployerKnative {
	var target *knativev1alpha1.Deployer
	switch len(seed) {
	case 0:
		target = &knativev1alpha1.Deployer{}
	case 1:
		target = seed[0]
	default:
		panic(fmt.Errorf("expected exactly zero or one seed, got %v", seed))
	}
	return &deployerKnative{
		target: target,
	}
}

func (f *deployerKnative) deepCopy() *deployerKnative {
	return DeployerKnative(f.target.DeepCopy())
}

func (f *deployerKnative) Create() *knativev1alpha1.Deployer {
	return f.deepCopy().target
}

func (f *deployerKnative) CreateObject() apis.Object {
	return f.Create()
}

func (f *deployerKnative) mutation(m func(*knativev1alpha1.Deployer)) *deployerKnative {
	f = f.deepCopy()
	m(f.target)
	return f
}

func (f *deployerKnative) NamespaceName(namespace, name string) *deployerKnative {
	return f.mutation(func(deployer *knativev1alpha1.Deployer) {
		deployer.ObjectMeta.Namespace = namespace
		deployer.ObjectMeta.Name = name
	})
}

func (f *deployerKnative) ObjectMeta(nf func(ObjectMeta)) *deployerKnative {
	return f.mutation(func(deployer *knativev1alpha1.Deployer) {
		omf := objectMeta(deployer.ObjectMeta)
		nf(omf)
		deployer.ObjectMeta = omf.Create()
	})
}

func (f *deployerKnative) PodTemplateSpec(nf func(PodTemplateSpec)) *deployerKnative {
	return f.mutation(func(deployer *knativev1alpha1.Deployer) {
		if deployer.Spec.Template == nil {
			deployer.Spec.Template = &corev1.PodTemplateSpec{}
		}
		ptsf := podTemplateSpec(*deployer.Spec.Template)
		nf(ptsf)
		pts := ptsf.Create()
		deployer.Spec.Template = &pts
	})
}

func (f *deployerKnative) ApplicationRef(format string, a ...interface{}) *deployerKnative {
	return f.mutation(func(deployer *knativev1alpha1.Deployer) {
		deployer.Spec.Build = &knativev1alpha1.Build{
			ApplicationRef: fmt.Sprintf(format, a...),
		}
	})
}

func (f *deployerKnative) ContainerRef(format string, a ...interface{}) *deployerKnative {
	return f.mutation(func(deployer *knativev1alpha1.Deployer) {
		deployer.Spec.Build = &knativev1alpha1.Build{
			ContainerRef: fmt.Sprintf(format, a...),
		}
	})
}

func (f *deployerKnative) FunctionRef(format string, a ...interface{}) *deployerKnative {
	return f.mutation(func(deployer *knativev1alpha1.Deployer) {
		deployer.Spec.Build = &knativev1alpha1.Build{
			FunctionRef: fmt.Sprintf(format, a...),
		}
	})
}

func (f *deployerKnative) Image(format string, a ...interface{}) *deployerKnative {
	return f.PodTemplateSpec(func(ptsf PodTemplateSpec) {
		ptsf.ContainerNamed("user-container", func(container *corev1.Container) {
			container.Image = fmt.Sprintf(format, a...)
		})
	})
}

func (f *deployerKnative) IngressPolicy(policy knativev1alpha1.IngressPolicy) *deployerKnative {
	return f.mutation(func(deployer *knativev1alpha1.Deployer) {
		deployer.Spec.IngressPolicy = policy
	})
}

func (f *deployerKnative) ContainerConcurrency(cc int64) *deployerKnative {
	return f.mutation(func(deployer *knativev1alpha1.Deployer) {
		deployer.Spec.ContainerConcurrency = &cc
	})
}

func (f *deployerKnative) MinScale(scale int32) *deployerKnative {
	return f.mutation(func(deployer *knativev1alpha1.Deployer) {
		deployer.Spec.Scale.Min = &scale
	})
}

func (f *deployerKnative) MaxScale(scale int32) *deployerKnative {
	return f.mutation(func(deployer *knativev1alpha1.Deployer) {
		deployer.Spec.Scale.Max = &scale
	})
}

func (f *deployerKnative) StatusConditions(conditions ...*condition) *deployerKnative {
	return f.mutation(func(deployer *knativev1alpha1.Deployer) {
		c := make([]apis.Condition, len(conditions))
		for i, cg := range conditions {
			c[i] = cg.Create()
		}
		deployer.Status.Conditions = c
	})
}

func (f *deployerKnative) StatusObservedGeneration(generation int64) *deployerKnative {
	return f.mutation(func(deployer *knativev1alpha1.Deployer) {
		deployer.Status.ObservedGeneration = generation
	})
}

func (f *deployerKnative) StatusLatestImage(format string, a ...interface{}) *deployerKnative {
	return f.mutation(func(deployer *knativev1alpha1.Deployer) {
		deployer.Status.LatestImage = fmt.Sprintf(format, a...)
	})
}

func (f *deployerKnative) StatusConfigurationRef(format string, a ...interface{}) *deployerKnative {
	return f.mutation(func(deployer *knativev1alpha1.Deployer) {
		deployer.Status.ConfigurationRef = &refs.TypedLocalObjectReference{
			APIGroup: rtesting.StringPtr("serving.knative.dev"),
			Kind:     "Configuration",
			Name:     fmt.Sprintf(format, a...),
		}
	})
}

func (f *deployerKnative) StatusRouteRef(format string, a ...interface{}) *deployerKnative {
	return f.mutation(func(deployer *knativev1alpha1.Deployer) {
		deployer.Status.RouteRef = &refs.TypedLocalObjectReference{
			APIGroup: rtesting.StringPtr("serving.knative.dev"),
			Kind:     "Route",
			Name:     fmt.Sprintf(format, a...),
		}
	})
}

func (f *deployerKnative) StatusAddressURL(url string) *deployerKnative {
	return f.mutation(func(deployer *knativev1alpha1.Deployer) {
		deployer.Status.Address = &apis.Addressable{
			URL: url,
		}
	})
}

func (f *deployerKnative) StatusURL(url string) *deployerKnative {
	return f.mutation(func(deployer *knativev1alpha1.Deployer) {
		deployer.Status.URL = url
	})
}
