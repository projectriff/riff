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

type serviceAccount struct {
	target *corev1.ServiceAccount
}

var (
	_ rtesting.Factory = (*serviceAccount)(nil)
)

func ServiceAccount(seed ...*corev1.ServiceAccount) *serviceAccount {
	var target *corev1.ServiceAccount
	switch len(seed) {
	case 0:
		target = &corev1.ServiceAccount{}
	case 1:
		target = seed[0]
	default:
		panic(fmt.Errorf("expected exactly zero or one seed, got %v", seed))
	}
	return &serviceAccount{
		target: target,
	}
}

func (f *serviceAccount) deepCopy() *serviceAccount {
	return ServiceAccount(f.target.DeepCopy())
}

func (f *serviceAccount) Create() *corev1.ServiceAccount {
	return f.deepCopy().target
}

func (f *serviceAccount) CreateObject() apis.Object {
	return f.Create()
}

func (f *serviceAccount) mutation(m func(*corev1.ServiceAccount)) *serviceAccount {
	f = f.deepCopy()
	m(f.target)
	return f
}

func (f *serviceAccount) NamespaceName(namespace, name string) *serviceAccount {
	return f.mutation(func(sa *corev1.ServiceAccount) {
		sa.ObjectMeta.Namespace = namespace
		sa.ObjectMeta.Name = name
	})
}

func (f *serviceAccount) ObjectMeta(nf func(ObjectMeta)) *serviceAccount {
	return f.mutation(func(sa *corev1.ServiceAccount) {
		omf := objectMeta(sa.ObjectMeta)
		nf(omf)
		sa.ObjectMeta = omf.Create()
	})
}

func (f *serviceAccount) Secrets(secrets ...string) *serviceAccount {
	return f.mutation(func(sa *corev1.ServiceAccount) {
		sa.Secrets = make([]corev1.ObjectReference, len(secrets))
		for i, secret := range secrets {
			sa.Secrets[i] = corev1.ObjectReference{Name: secret}
		}
	})
}
