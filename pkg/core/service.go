/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"errors"

	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	istioNamespace     = "istio-system"
	ingressServiceName = "knative-ingressgateway"
)

type ListServiceOptions struct {
	Namespaced
}

func (c *client) ListServices(options ListServiceOptions) (*v1alpha1.ServiceList, error) {
	ns := c.explicitOrConfigNamespace(options.Namespaced)
	return c.serving.ServingV1alpha1().Services(ns).List(meta_v1.ListOptions{})
}

type CreateServiceOptions struct {
	Namespaced
	Name string

	Image string
}

func (c *client) CreateService(options CreateServiceOptions) (*v1alpha1.Service, error) {
	ns := c.explicitOrConfigNamespace(options.Namespaced)

	s := newService(options)

	_, err := c.serving.ServingV1alpha1().Services(ns).Create(&s)

	return &s, err
}

func newService(options CreateServiceOptions) v1alpha1.Service {
	return v1alpha1.Service{
		TypeMeta: meta_v1.TypeMeta{
			APIVersion: "serving.knative.dev/v1alpha1",
			Kind:       "Service",
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name: options.Name,
		},
		Spec: v1alpha1.ServiceSpec{
			RunLatest: &v1alpha1.RunLatestType{
				Configuration: v1alpha1.ConfigurationSpec{
					RevisionTemplate: v1alpha1.RevisionTemplateSpec{
						Spec: v1alpha1.RevisionSpec{
							Container: core_v1.Container{
								Image: options.Image,
							},
						},
					},
				},
			},
		},
	}
}

type ServiceStatusOptions struct {
	Namespaced
	Name string
}

func (c *client) ServiceStatus(options ServiceStatusOptions) (*v1alpha1.ServiceCondition, error) {

	s, err := c.service(options.Namespaced, options.Name)
	if err != nil {
		return nil, err
	}

	for _, cond := range s.Status.Conditions {
		if cond.Type == v1alpha1.ServiceConditionReady {
			return &cond, nil
		}
	}

	return nil, errors.New("No condition of type ServiceConditionReady found for the service")
}

type ServiceInvokeOptions struct {
	Namespaced
	Name string
}

func (c *client) ServiceCoordinates(options ServiceInvokeOptions) (string, string, error) {

	ksvc, err := c.kubeClient.CoreV1().Services(istioNamespace).Get(ingressServiceName, meta_v1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	ingresses := ksvc.Status.LoadBalancer.Ingress
	if len(ingresses) == 0 {
		return "", "", errors.New("Ingress not available")
	}
	ingressIP := ingresses[0].IP

	s, err := c.service(options.Namespaced, options.Name)
	if err != nil {
		return "", "", err
	}

	return ingressIP, s.Status.Domain, nil
}

func (c *client) service(namespace Namespaced, name string) (*v1alpha1.Service, error) {

	ns := c.explicitOrConfigNamespace(namespace)

	return c.serving.ServingV1alpha1().Services(ns).Get(name, meta_v1.GetOptions{})
}

type DeleteServiceOptions struct {
	Namespaced
	Name string
}

func (c *client) DeleteService(options DeleteServiceOptions) error {

	ns := c.explicitOrConfigNamespace(options.Namespaced)

	return c.serving.ServingV1alpha1().Services(ns).Delete(options.Name, nil)
}
