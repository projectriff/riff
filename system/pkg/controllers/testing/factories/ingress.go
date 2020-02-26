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
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/projectriff/system/pkg/apis"
	rtesting "github.com/projectriff/system/pkg/controllers/testing"
)

type ingress struct {
	target *networkingv1beta1.Ingress
}

var (
	_ rtesting.Factory = (*ingress)(nil)
)

func Ingress(seed ...*networkingv1beta1.Ingress) *ingress {
	var target *networkingv1beta1.Ingress
	switch len(seed) {
	case 0:
		target = &networkingv1beta1.Ingress{}
	case 1:
		target = seed[0]
	default:
		panic(fmt.Errorf("expected exactly zero or one seed, got %v", seed))
	}
	return &ingress{
		target: target,
	}
}

func (f *ingress) deepCopy() *ingress {
	return Ingress(f.target.DeepCopy())
}

func (f *ingress) Create() *networkingv1beta1.Ingress {
	return f.deepCopy().target
}

func (f *ingress) CreateObject() apis.Object {
	return f.Create()
}

func (f *ingress) mutation(m func(*networkingv1beta1.Ingress)) *ingress {
	f = f.deepCopy()
	m(f.target)
	return f
}

func (f *ingress) NamespaceName(namespace, name string) *ingress {
	return f.mutation(func(sa *networkingv1beta1.Ingress) {
		sa.ObjectMeta.Namespace = namespace
		sa.ObjectMeta.Name = name
	})
}

func (f *ingress) ObjectMeta(nf func(ObjectMeta)) *ingress {
	return f.mutation(func(sa *networkingv1beta1.Ingress) {
		omf := objectMeta(sa.ObjectMeta)
		nf(omf)
		sa.ObjectMeta = omf.Create()
	})
}

func (f *ingress) HostToService(host, serviceName string) *ingress {
	return f.mutation(func(i *networkingv1beta1.Ingress) {
		i.Spec = networkingv1beta1.IngressSpec{
			Rules: []networkingv1beta1.IngressRule{{
				Host: host,
				IngressRuleValue: networkingv1beta1.IngressRuleValue{
					HTTP: &networkingv1beta1.HTTPIngressRuleValue{
						Paths: []networkingv1beta1.HTTPIngressPath{{
							Path: "/",
							Backend: networkingv1beta1.IngressBackend{
								ServiceName: serviceName,
								ServicePort: intstr.FromInt(80),
							},
						}},
					},
				},
			}},
		}
	})
}

func (f *ingress) StatusLoadBalancer(ingress ...corev1.LoadBalancerIngress) *ingress {
	return f.mutation(func(i *networkingv1beta1.Ingress) {
		i.Status.LoadBalancer.Ingress = ingress
	})
}
