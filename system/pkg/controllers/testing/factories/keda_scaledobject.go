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
	kedav1alpha1 "github.com/projectriff/system/pkg/apis/thirdparty/keda/v1alpha1"
	rtesting "github.com/projectriff/system/pkg/controllers/testing"
)

type kedaScaledObject struct {
	target *kedav1alpha1.ScaledObject
}

var (
	_ rtesting.Factory = (*kedaScaledObject)(nil)
)

func KedaScaledObject(seed ...*kedav1alpha1.ScaledObject) *kedaScaledObject {
	var target *kedav1alpha1.ScaledObject
	switch len(seed) {
	case 0:
		target = &kedav1alpha1.ScaledObject{}
	case 1:
		target = seed[0]
	default:
		panic(fmt.Errorf("expected exactly zero or one seed, got %v", seed))
	}
	return &kedaScaledObject{
		target: target,
	}
}

func (f *kedaScaledObject) deepCopy() *kedaScaledObject {
	return KedaScaledObject(f.target.DeepCopy())
}

func (f *kedaScaledObject) Create() *kedav1alpha1.ScaledObject {
	return f.deepCopy().target
}

func (f *kedaScaledObject) CreateObject() apis.Object {
	return f.Create()
}

func (f *kedaScaledObject) mutation(m func(*kedav1alpha1.ScaledObject)) *kedaScaledObject {
	f = f.deepCopy()
	m(f.target)
	return f
}

func (f *kedaScaledObject) NamespaceName(namespace, name string) *kedaScaledObject {
	return f.mutation(func(s *kedav1alpha1.ScaledObject) {
		s.ObjectMeta.Namespace = namespace
		s.ObjectMeta.Name = name
	})
}

func (f *kedaScaledObject) ObjectMeta(nf func(ObjectMeta)) *kedaScaledObject {
	return f.mutation(func(s *kedav1alpha1.ScaledObject) {
		omf := objectMeta(s.ObjectMeta)
		nf(omf)
		s.ObjectMeta = omf.Create()
	})
}

func (f *kedaScaledObject) Spec(spec *kedav1alpha1.ScaledObjectSpec) *kedaScaledObject {
	return f.mutation(func(s *kedav1alpha1.ScaledObject) {
		s.Spec = *spec
	})
}

func (f *kedaScaledObject) ScaleTargetRefDeployment(format string, a ...interface{}) *kedaScaledObject {
	return f.mutation(func(s *kedav1alpha1.ScaledObject) {
		s.Spec.ScaleTargetRef = &kedav1alpha1.ObjectReference{DeploymentName: fmt.Sprintf(format, a...)}
	})
}

func (f *kedaScaledObject) Triggers(triggers ...kedav1alpha1.ScaleTriggers) *kedaScaledObject {
	return f.mutation(func(s *kedav1alpha1.ScaledObject) {
		s.Spec.Triggers = triggers
	})
}

func (f *kedaScaledObject) PollingInterval(pollingInterval int32) *kedaScaledObject {
	return f.mutation(func(s *kedav1alpha1.ScaledObject) {
		s.Spec.PollingInterval = &pollingInterval
	})
}

func (f *kedaScaledObject) CooldownPeriod(cooldownPeriod int32) *kedaScaledObject {
	return f.mutation(func(s *kedav1alpha1.ScaledObject) {
		s.Spec.CooldownPeriod = &cooldownPeriod
	})
}

func (f *kedaScaledObject) MinReplicaCount(minReplicaCount int32) *kedaScaledObject {
	return f.mutation(func(s *kedav1alpha1.ScaledObject) {
		s.Spec.MinReplicaCount = &minReplicaCount
	})
}

func (f *kedaScaledObject) MaxReplicaCount(maxReplicaCount int32) *kedaScaledObject {
	return f.mutation(func(s *kedav1alpha1.ScaledObject) {
		s.Spec.MaxReplicaCount = &maxReplicaCount
	})
}
