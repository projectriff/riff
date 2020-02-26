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

	"github.com/projectriff/riff/system/pkg/apis"
	kpackbuildv1alpha1 "github.com/projectriff/riff/system/pkg/apis/thirdparty/kpack/build/v1alpha1"
	rtesting "github.com/projectriff/riff/system/pkg/controllers/testing"
)

type kpackClusterBuilder struct {
	target *kpackbuildv1alpha1.ClusterBuilder
}

var (
	_ rtesting.Factory = (*kpackClusterBuilder)(nil)
)

func KpackClusterBuilder(seed ...*kpackbuildv1alpha1.ClusterBuilder) *kpackClusterBuilder {
	var target *kpackbuildv1alpha1.ClusterBuilder
	switch len(seed) {
	case 0:
		target = &kpackbuildv1alpha1.ClusterBuilder{}
	case 1:
		target = seed[0]
	default:
		panic(fmt.Errorf("expected exactly zero or one seed, got %v", seed))
	}
	return &kpackClusterBuilder{
		target: target,
	}
}

func (f *kpackClusterBuilder) deepCopy() *kpackClusterBuilder {
	return KpackClusterBuilder(f.target.DeepCopy())
}

func (f *kpackClusterBuilder) Create() *kpackbuildv1alpha1.ClusterBuilder {
	return f.deepCopy().target
}

func (f *kpackClusterBuilder) CreateObject() apis.Object {
	return f.Create()
}

func (f *kpackClusterBuilder) mutation(m func(*kpackbuildv1alpha1.ClusterBuilder)) *kpackClusterBuilder {
	f = f.deepCopy()
	m(f.target)
	return f
}

func (f *kpackClusterBuilder) NamespaceName(namespace, name string) *kpackClusterBuilder {
	return f.mutation(func(cb *kpackbuildv1alpha1.ClusterBuilder) {
		cb.ObjectMeta.Namespace = namespace
		cb.ObjectMeta.Name = name
	})
}

func (f *kpackClusterBuilder) ObjectMeta(nf func(ObjectMeta)) *kpackClusterBuilder {
	return f.mutation(func(cb *kpackbuildv1alpha1.ClusterBuilder) {
		omf := objectMeta(cb.ObjectMeta)
		nf(omf)
		cb.ObjectMeta = omf.Create()
	})
}

func (f *kpackClusterBuilder) Image(format string, a ...interface{}) *kpackClusterBuilder {
	return f.mutation(func(cb *kpackbuildv1alpha1.ClusterBuilder) {
		cb.Spec.Image = fmt.Sprintf(format, a...)
	})
}

func (f *kpackClusterBuilder) StatusConditions(conditions ...*condition) *kpackClusterBuilder {
	return f.mutation(func(cb *kpackbuildv1alpha1.ClusterBuilder) {
		c := make([]apis.Condition, len(conditions))
		for i, cg := range conditions {
			c[i] = cg.Create()
		}
		cb.Status.Conditions = c
	})
}

func (f *kpackClusterBuilder) StatusReady() *kpackClusterBuilder {
	return f.StatusConditions(
		Condition().Type(apis.ConditionReady).True(),
	)
}

func (f *kpackClusterBuilder) StatusObservedGeneration(generation int64) *kpackClusterBuilder {
	return f.mutation(func(cb *kpackbuildv1alpha1.ClusterBuilder) {
		cb.Status.ObservedGeneration = generation
	})
}

func (f *kpackClusterBuilder) StatusLatestImage(format string, a ...interface{}) *kpackClusterBuilder {
	return f.mutation(func(cb *kpackbuildv1alpha1.ClusterBuilder) {
		cb.Status.LatestImage = fmt.Sprintf(format, a...)
	})
}
