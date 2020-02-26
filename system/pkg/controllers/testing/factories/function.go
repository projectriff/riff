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

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/projectriff/system/pkg/apis"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	rtesting "github.com/projectriff/system/pkg/controllers/testing"
	"github.com/projectriff/system/pkg/refs"
)

type function struct {
	target *buildv1alpha1.Function
}

var (
	_ rtesting.Factory = (*function)(nil)
)

func Function(seed ...*buildv1alpha1.Function) *function {
	var target *buildv1alpha1.Function
	switch len(seed) {
	case 0:
		target = &buildv1alpha1.Function{}
	case 1:
		target = seed[0]
	default:
		panic(fmt.Errorf("expected exactly zero or one seed, got %v", seed))
	}
	return &function{
		target: target,
	}
}

func (f *function) deepCopy() *function {
	return Function(f.target.DeepCopy())
}

func (f *function) Create() *buildv1alpha1.Function {
	return f.deepCopy().target
}

func (f *function) CreateObject() apis.Object {
	return f.Create()
}

func (f *function) mutation(m func(*buildv1alpha1.Function)) *function {
	f = f.deepCopy()
	m(f.target)
	return f
}

func (f *function) NamespaceName(namespace, name string) *function {
	return f.mutation(func(fn *buildv1alpha1.Function) {
		fn.ObjectMeta.Namespace = namespace
		fn.ObjectMeta.Name = name
	})
}

func (f *function) ObjectMeta(nf func(ObjectMeta)) *function {
	return f.mutation(func(fn *buildv1alpha1.Function) {
		omf := objectMeta(fn.ObjectMeta)
		nf(omf)
		fn.ObjectMeta = omf.Create()
	})
}

func (f *function) Image(format string, a ...interface{}) *function {
	return f.mutation(func(fn *buildv1alpha1.Function) {
		fn.Spec.Image = fmt.Sprintf(format, a...)
	})
}

func (f *function) Artifact(artifact string) *function {
	return f.mutation(func(fn *buildv1alpha1.Function) {
		fn.Spec.Artifact = artifact
	})
}

func (f *function) Handler(handler string) *function {
	return f.mutation(func(fn *buildv1alpha1.Function) {
		fn.Spec.Handler = handler
	})
}

func (f *function) Invoker(invoker string) *function {
	return f.mutation(func(fn *buildv1alpha1.Function) {
		fn.Spec.Invoker = invoker
	})
}

func (f *function) SourceGit(url string, revision string) *function {
	return f.mutation(func(fn *buildv1alpha1.Function) {
		if fn.Spec.Source == nil {
			fn.Spec.Source = &buildv1alpha1.Source{}
		}
		fn.Spec.Source = &buildv1alpha1.Source{
			Git: &buildv1alpha1.Git{
				URL:      url,
				Revision: revision,
			},
			SubPath: fn.Spec.Source.SubPath,
		}
	})
}

func (f *function) SourceSubPath(subpath string) *function {
	return f.mutation(func(fn *buildv1alpha1.Function) {
		if fn.Spec.Source == nil {
			fn.Spec.Source = &buildv1alpha1.Source{}
		}
		fn.Spec.Source.SubPath = subpath
	})
}

func (f *function) BuildCache(quantity string) *function {
	return f.mutation(func(fn *buildv1alpha1.Function) {
		size, err := resource.ParseQuantity(quantity)
		if err != nil {
			panic(err)
		}
		fn.Spec.CacheSize = &size
	})
}

func (f *function) StatusConditions(conditions ...*condition) *function {
	return f.mutation(func(fn *buildv1alpha1.Function) {
		c := make([]apis.Condition, len(conditions))
		for i, cg := range conditions {
			c[i] = cg.Create()
		}
		fn.Status.Conditions = c
	})
}

func (f *function) StatusReady() *function {
	return f.StatusConditions(
		Condition().Type(buildv1alpha1.FunctionConditionReady).True(),
	)
}

func (f *function) StatusObservedGeneration(generation int64) *function {
	return f.mutation(func(fn *buildv1alpha1.Function) {
		fn.Status.ObservedGeneration = generation
	})
}

func (f *function) StatusKpackImageRef(format string, a ...interface{}) *function {
	return f.mutation(func(fn *buildv1alpha1.Function) {
		fn.Status.KpackImageRef = &refs.TypedLocalObjectReference{
			APIGroup: rtesting.StringPtr("build.pivotal.io"),
			Kind:     "Image",
			Name:     fmt.Sprintf(format, a...),
		}
	})
}

func (f *function) StatusBuildCacheRef(format string, a ...interface{}) *function {
	return f.mutation(func(fn *buildv1alpha1.Function) {
		fn.Status.BuildCacheRef = &refs.TypedLocalObjectReference{
			Kind: "PersistentVolumeClaim",
			Name: fmt.Sprintf(format, a...),
		}
	})
}

func (f *function) StatusTargetImage(format string, a ...interface{}) *function {
	return f.mutation(func(fn *buildv1alpha1.Function) {
		fn.Status.TargetImage = fmt.Sprintf(format, a...)
	})
}

func (f *function) StatusLatestImage(format string, a ...interface{}) *function {
	return f.mutation(func(fn *buildv1alpha1.Function) {
		fn.Status.LatestImage = fmt.Sprintf(format, a...)
	})
}
