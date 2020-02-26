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
	rtesting "github.com/projectriff/system/pkg/controllers/testing"
)

type secret struct {
	target *corev1.Secret
}

var (
	_ rtesting.Factory = (*secret)(nil)
)

func Secret(seed ...*corev1.Secret) *secret {
	var target *corev1.Secret
	switch len(seed) {
	case 0:
		target = &corev1.Secret{}
	case 1:
		target = seed[0]
	default:
		panic(fmt.Errorf("expected exactly zero or one seed, got %v", seed))
	}
	return &secret{
		target: target,
	}
}

func (f *secret) deepCopy() *secret {
	return Secret(f.target.DeepCopy())
}

func (f *secret) Create() *corev1.Secret {
	return f.deepCopy().target
}

func (f *secret) CreateObject() apis.Object {
	return f.Create()
}

func (f *secret) mutation(m func(*corev1.Secret)) *secret {
	f = f.deepCopy()
	m(f.target)
	return f
}

func (f *secret) NamespaceName(namespace, name string) *secret {
	return f.mutation(func(s *corev1.Secret) {
		s.ObjectMeta.Namespace = namespace
		s.ObjectMeta.Name = name
	})
}

func (f *secret) ObjectMeta(nf func(ObjectMeta)) *secret {
	return f.mutation(func(s *corev1.Secret) {
		omf := objectMeta(s.ObjectMeta)
		nf(omf)
		s.ObjectMeta = omf.Create()
	})
}

func (f *secret) Type(t corev1.SecretType) *secret {
	return f.mutation(func(s *corev1.Secret) {
		s.Type = t
	})
}

func (f *secret) AddData(key, value string) *secret {
	return f.mutation(func(s *corev1.Secret) {
		if s.Data == nil {
			s.Data = map[string][]byte{}
		}
		s.Data[key] = []byte(value)
	})
}
