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
	rtesting "github.com/projectriff/riff/system/pkg/controllers/testing"
)

type service struct {
	target *corev1.Service
}

var (
	_ rtesting.Factory = (*service)(nil)
)

func Service(seed ...*corev1.Service) *service {
	var target *corev1.Service
	switch len(seed) {
	case 0:
		target = &corev1.Service{}
	case 1:
		target = seed[0]
	default:
		panic(fmt.Errorf("expected exactly zero or one seed, got %v", seed))
	}
	return &service{
		target: target,
	}
}

func (f *service) deepCopy() *service {
	return Service(f.target.DeepCopy())
}

func (f *service) Create() *corev1.Service {
	return f.deepCopy().target
}

func (f *service) CreateObject() apis.Object {
	return f.Create()
}

func (f *service) mutation(m func(*corev1.Service)) *service {
	f = f.deepCopy()
	m(f.target)
	return f
}

func (f *service) NamespaceName(namespace, name string) *service {
	return f.mutation(func(sa *corev1.Service) {
		sa.ObjectMeta.Namespace = namespace
		sa.ObjectMeta.Name = name
	})
}

func (f *service) ObjectMeta(nf func(ObjectMeta)) *service {
	return f.mutation(func(sa *corev1.Service) {
		omf := objectMeta(sa.ObjectMeta)
		nf(omf)
		sa.ObjectMeta = omf.Create()
	})
}

func (f *service) AddSelectorLabel(key, value string) *service {
	return f.mutation(func(service *corev1.Service) {
		if service.Spec.Selector == nil {
			service.Spec.Selector = map[string]string{}
		}
		service.Spec.Selector[key] = value
	})
}

func (f *service) Ports(ports ...corev1.ServicePort) *service {
	return f.mutation(func(service *corev1.Service) {
		service.Spec.Ports = ports
	})
}

func (f *service) ClusterIP(ip string) *service {
	return f.mutation(func(service *corev1.Service) {
		service.Spec.ClusterIP = ip
	})
}
