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
	"fmt"
	"strings"

	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ingressServiceName = "knative-ingressgateway"
)

type ListServiceOptions struct {
	Namespace string
}

func (c *client) ListServices(options ListServiceOptions) (*v1alpha1.ServiceList, error) {
	ns := c.explicitOrConfigNamespace(options.Namespace)
	return c.serving.ServingV1alpha1().Services(ns).List(meta_v1.ListOptions{})
}

type CreateOrReviseServiceOptions struct {
	Namespace string
	Name      string
	Image     string
	Env       []string
	EnvFrom   []string
	DryRun    bool
	Verbose   bool
	Wait      bool
}

func (c *client) CreateService(options CreateOrReviseServiceOptions) (*v1alpha1.Service, error) {
	ns := c.explicitOrConfigNamespace(options.Namespace)

	s, err := newService(options)
	if err != nil {
		return nil, err
	}

	if !options.DryRun {
		_, err := c.serving.ServingV1alpha1().Services(ns).Create(s)
		return s, err
	} else {
		return s, nil
	}

}

func (c *client) ReviseService(options CreateOrReviseServiceOptions) (*v1alpha1.Service, error) {
	ns := c.explicitOrConfigNamespace(options.Namespace)

	existingSvc, err := c.serving.ServingV1alpha1().Services(ns).Get(options.Name, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if options.Image != "" {
		existingSvc.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Image = options.Image
	}
	envVars, err := ParseEnvVar(options.Env)
	if err != nil {
		return nil, err
	}
	envVarsFrom, err := ParseEnvVarSource(options.EnvFrom)
	if err != nil {
		return nil, err
	}
	existingSvc.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Env = append(existingSvc.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Env, envVars...)
	existingSvc.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Env = append(existingSvc.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Env, envVarsFrom...)

	c.bumpNonceAnnotation(existingSvc)

	if !options.DryRun {
		_, err := c.serving.ServingV1alpha1().Services(ns).Update(existingSvc)
		return existingSvc, err
	} else {
		existingSvc.Status = v1alpha1.ServiceStatus{}
		return existingSvc, nil
	}

}

func newService(options CreateOrReviseServiceOptions) (*v1alpha1.Service, error) {
	envVars, err := ParseEnvVar(options.Env)
	if err != nil {
		return nil, err
	}
	envVarsFrom, err := ParseEnvVarSource(options.EnvFrom)
	if err != nil {
		return nil, err
	}
	envVars = append(envVars, envVarsFrom...)

	s := v1alpha1.Service{
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
								Env:   envVars,
								Image: options.Image,
							},
						},
					},
				},
			},
		},
	}

	return &s, nil
}

type ServiceStatusOptions struct {
	Namespace string
	Name      string
}

func (c *client) ServiceStatus(options ServiceStatusOptions) (*v1alpha1.ServiceCondition, error) {

	conds, err := c.ServiceConditions(options)
	if err != nil {
		return nil, err
	}

	for _, cond := range conds {
		if cond.Type == v1alpha1.ServiceConditionReady {
			return &cond, nil
		}
	}

	return nil, errors.New("No condition of type ServiceConditionReady found for the service")
}

func (c *client) ServiceConditions(options ServiceStatusOptions) ([]v1alpha1.ServiceCondition, error) {

	s, err := c.service(options.Namespace, options.Name)
	if err != nil {
		return nil, err
	}
	return s.Status.Conditions, nil
}

type ServiceInvokeOptions struct {
	Namespace       string
	Name            string
	ContentTypeText bool
	ContentTypeJson bool
}

func (c *client) ServiceCoordinates(options ServiceInvokeOptions) (string, string, error) {

	ksvc, err := c.kubeClient.CoreV1().Services(istioNamespace).Get(ingressServiceName, meta_v1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	var ingress string
	if ksvc.Spec.Type == "LoadBalancer" {
		ingresses := ksvc.Status.LoadBalancer.Ingress
		if len(ingresses) > 0 {
			ingress = ingresses[0].IP
			if ingress == "" {
				ingress = ingresses[0].Hostname
			}
		}
	}
	if ingress == "" {
		for _, port := range ksvc.Spec.Ports {
			if port.Name == "http" || port.Name == "http2" {
				config, err := c.clientConfig.ClientConfig()
				if err != nil {
					return "", "", err
				}
				host := config.Host[0:strings.LastIndex(config.Host, ":")]
				host = strings.Replace(host, "https", "http", 1)
				ingress = fmt.Sprintf("%s:%d", host, port.NodePort)
			}
		}
		if ingress == "" {
			return "", "", errors.New("Ingress not available")
		}
	}

	s, err := c.service(options.Namespace, options.Name)
	if err != nil {
		return "", "", err
	}

	return ingress, s.Status.Domain, nil
}

func (c *client) service(namespace string, name string) (*v1alpha1.Service, error) {
	ns := c.explicitOrConfigNamespace(namespace)
	return c.serving.ServingV1alpha1().Services(ns).Get(name, meta_v1.GetOptions{})
}

type DeleteServiceOptions struct {
	Namespace string
	Name      string
}

func (c *client) DeleteService(options DeleteServiceOptions) error {
	ns := c.explicitOrConfigNamespace(options.Namespace)
	return c.serving.ServingV1alpha1().Services(ns).Delete(options.Name, nil)
}
