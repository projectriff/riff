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

	"github.com/projectriff/system/pkg/apis"
	streamingv1alpha1 "github.com/projectriff/system/pkg/apis/streaming/v1alpha1"
	rtesting "github.com/projectriff/system/pkg/controllers/testing"
	"github.com/projectriff/system/pkg/refs"
)

type processor struct {
	target *streamingv1alpha1.Processor
}

var (
	_ rtesting.Factory = (*processor)(nil)
)

func Processor(seed ...*streamingv1alpha1.Processor) *processor {
	var target *streamingv1alpha1.Processor
	switch len(seed) {
	case 0:
		target = &streamingv1alpha1.Processor{}
	case 1:
		target = seed[0]
	default:
		panic(fmt.Errorf("expected exactly zero or one seed, got %v", seed))
	}
	return &processor{
		target: target,
	}
}

func (f *processor) deepCopy() *processor {
	return Processor(f.target.DeepCopy())
}

func (f *processor) Create() *streamingv1alpha1.Processor {
	return f.deepCopy().target
}

func (f *processor) CreateObject() apis.Object {
	return f.Create()
}

func (f *processor) mutation(m func(*streamingv1alpha1.Processor)) *processor {
	f = f.deepCopy()
	m(f.target)
	return f
}

func (f *processor) Default() *processor {
	return f.mutation(func(p *streamingv1alpha1.Processor) {
		p.Default()
	})
}

func (f *processor) NamespaceName(namespace, name string) *processor {
	return f.mutation(func(p *streamingv1alpha1.Processor) {
		p.ObjectMeta.Namespace = namespace
		p.ObjectMeta.Name = name
	})
}

func (f *processor) ObjectMeta(nf func(ObjectMeta)) *processor {
	return f.mutation(func(s *streamingv1alpha1.Processor) {
		omf := objectMeta(s.ObjectMeta)
		nf(omf)
		s.ObjectMeta = omf.Create()
	})
}

func (f *processor) BuildFunctionRef(function *function) *processor {
	return f.mutation(func(proc *streamingv1alpha1.Processor) {
		proc.Spec.Build = &streamingv1alpha1.Build{
			FunctionRef: function.Create().Name,
		}
	})
}

func (f *processor) BuildContainerRef(container *container) *processor {
	return f.mutation(func(proc *streamingv1alpha1.Processor) {
		proc.Spec.Build = &streamingv1alpha1.Build{
			ContainerRef: container.Create().Name,
		}
	})
}

func (f *processor) Image(image string) *processor {
	return f.PodTemplateSpec(func(pts PodTemplateSpec) {
		pts.ContainerNamed("function", func(c *corev1.Container) {
			c.Image = image
		})
	})
}

func (f *processor) Inputs(inputs ...streamingv1alpha1.InputStreamBinding) *processor {
	return f.mutation(func(proc *streamingv1alpha1.Processor) {
		proc.Spec.Inputs = inputs
	})
}

func (f *processor) Outputs(outputs ...streamingv1alpha1.OutputStreamBinding) *processor {
	return f.mutation(func(proc *streamingv1alpha1.Processor) {
		proc.Spec.Outputs = outputs
	})
}

func (f *processor) PodTemplateSpec(nf func(PodTemplateSpec)) *processor {
	return f.mutation(func(processor *streamingv1alpha1.Processor) {
		var ptsf *podTemplateSpecImpl
		if processor.Spec.Template != nil {
			ptsf = podTemplateSpec(*processor.Spec.Template)
		} else {
			ptsf = podTemplateSpec(corev1.PodTemplateSpec{})
		}
		nf(ptsf)
		templateSpec := ptsf.Create()
		processor.Spec.Template = &templateSpec
	})
}

func (f *processor) StatusConditions(conditions ...*condition) *processor {
	return f.mutation(func(processor *streamingv1alpha1.Processor) {
		c := make([]apis.Condition, len(conditions))
		for i, cg := range conditions {
			dc := cg.Create()
			c[i] = apis.Condition{
				Type:    apis.ConditionType(dc.Type),
				Status:  dc.Status,
				Reason:  dc.Reason,
				Message: dc.Message,
			}
		}
		processor.Status.Conditions = c
	})
}

func (f *processor) StatusObservedGeneration(generation int64) *processor {
	return f.mutation(func(proc *streamingv1alpha1.Processor) {
		proc.Status.ObservedGeneration = generation
	})
}

func (f *processor) StatusLatestImage(image string) *processor {
	return f.mutation(func(proc *streamingv1alpha1.Processor) {
		proc.Status.LatestImage = image
	})
}

func (f *processor) StatusDeploymentRef(format string, a ...interface{}) *processor {
	return f.mutation(func(proc *streamingv1alpha1.Processor) {
		proc.Status.DeploymentRef = &refs.TypedLocalObjectReference{
			APIGroup: rtesting.StringPtr("apps"),
			Kind:     "Deployment",
			Name:     fmt.Sprintf(format, a...),
		}
	})
}

func (f *processor) StatusScaledObjectRef(format string, a ...interface{}) *processor {
	return f.mutation(func(proc *streamingv1alpha1.Processor) {
		proc.Status.ScaledObjectRef = &refs.TypedLocalObjectReference{
			APIGroup: rtesting.StringPtr("keda.k8s.io"),
			Kind:     "ScaledObject",
			Name:     fmt.Sprintf(format, a...),
		}
	})
}
