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

	"github.com/projectriff/system/pkg/apis"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	rtesting "github.com/projectriff/system/pkg/controllers/testing"
)

type container struct {
	target *buildv1alpha1.Container
}

var (
	_ rtesting.Factory = (*container)(nil)
)

func Container(seed ...*buildv1alpha1.Container) *container {
	var target *buildv1alpha1.Container
	switch len(seed) {
	case 0:
		target = &buildv1alpha1.Container{}
	case 1:
		target = seed[0]
	default:
		panic(fmt.Errorf("expected exactly zero or one seed, got %v", seed))
	}
	return &container{
		target: target,
	}
}

func (f *container) deepCopy() *container {
	return Container(f.target.DeepCopy())
}

func (f *container) Create() *buildv1alpha1.Container {
	return f.deepCopy().target
}

func (f *container) CreateObject() apis.Object {
	return f.Create()
}

func (f *container) mutation(m func(*buildv1alpha1.Container)) *container {
	f = f.deepCopy()
	m(f.target)
	return f
}

func (f *container) NamespaceName(namespace, name string) *container {
	return f.mutation(func(con *buildv1alpha1.Container) {
		con.ObjectMeta.Namespace = namespace
		con.ObjectMeta.Name = name
	})
}

func (f *container) ObjectMeta(nf func(ObjectMeta)) *container {
	return f.mutation(func(con *buildv1alpha1.Container) {
		omf := objectMeta(con.ObjectMeta)
		nf(omf)
		con.ObjectMeta = omf.Create()
	})
}

func (f *container) Image(format string, a ...interface{}) *container {
	return f.mutation(func(con *buildv1alpha1.Container) {
		con.Spec.Image = fmt.Sprintf(format, a...)
	})
}

func (f *container) StatusConditions(conditions ...*condition) *container {
	return f.mutation(func(con *buildv1alpha1.Container) {
		c := make([]apis.Condition, len(conditions))
		for i, cg := range conditions {
			c[i] = cg.Create()
		}
		con.Status.Conditions = c
	})
}

func (f *container) StatusReady() *container {
	return f.StatusConditions(
		Condition().Type(buildv1alpha1.ContainerConditionReady).True(),
	)
}

func (f *container) StatusObservedGeneration(generation int64) *container {
	return f.mutation(func(con *buildv1alpha1.Container) {
		con.Status.ObservedGeneration = generation
	})
}

func (f *container) StatusTargetImage(format string, a ...interface{}) *container {
	return f.mutation(func(con *buildv1alpha1.Container) {
		con.Status.TargetImage = fmt.Sprintf(format, a...)
	})
}

func (f *container) StatusLatestImage(format string, a ...interface{}) *container {
	return f.mutation(func(con *buildv1alpha1.Container) {
		con.Status.LatestImage = fmt.Sprintf(format, a...)
	})
}
