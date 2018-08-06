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
	"k8s.io/apimachinery/pkg/api/resource"
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

	Env []string

	DryRun bool
}

func (c *client) CreateService(options CreateServiceOptions) (*v1alpha1.Service, error) {
	ns := c.explicitOrConfigNamespace(options.Namespaced)

	s, err := newService(options)
	if err != nil {
		return nil, err
	}

	if !options.DryRun {
		_, err := c.serving.ServingV1alpha1().Services(ns).Create(&s)
		return &s, err
	} else {
		return &s, nil
	}

}

func newService(options CreateServiceOptions) (v1alpha1.Service, error) {

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
								Image: options.Image,
							},
						},
					},
				},
			},
		},
	}

	if len(options.Env) > 0 {
		var envVars []core_v1.EnvVar
		for _, env := range options.Env {
			envEntry := strings.SplitN(env, "=", 2)
			if len(envEntry) != 2 {
				return v1alpha1.Service{}, errors.New(fmt.Sprintf("unable to parse '%v', environment variables must provided as 'name=value'", env))
			}
			envKey := strings.TrimSpace(envEntry[0])
			envValue := strings.TrimSpace(envEntry[1])
			if strings.HasPrefix(envValue, "valueFrom:") {
				source := strings.Split(envValue, ":")
				if len(source) < 3 {
					return v1alpha1.Service{}, errors.New(fmt.Sprintf("unable to parse 'valueFrom' entry '%v', it should be provided as a secretKeyRef, configMapKeyRef, fieldRef or resourceFieldRef", env))
				}
				sourceType := strings.TrimSpace(source[1])
				switch sourceType {
				case "secretKeyRef":
					{
						if len(source) != 4 {
							return v1alpha1.Service{}, errors.New(fmt.Sprintf("unable to parse 'valueFrom' entry '%v', it should be provided as valueFrom:secretKeyRef:{name}:{key}", env))
						}
						envVars = append(envVars, core_v1.EnvVar{
							Name: envKey,
							ValueFrom: &core_v1.EnvVarSource{
								SecretKeyRef: &core_v1.SecretKeySelector{
									LocalObjectReference: core_v1.LocalObjectReference{
										Name: strings.TrimSpace(source[2]),
									},
									Key: strings.TrimSpace(source[3]),
								},
							},
						})
					}
				case "configMapKeyRef":
					{
						if len(source) != 4 {
							return v1alpha1.Service{}, errors.New(fmt.Sprintf("unable to parse 'valueFrom' entry '%v', it should be provided as valueFrom:configMapKeyRef:{name}:{key}", env))
						}
						envVars = append(envVars, core_v1.EnvVar{
							Name: envKey,
							ValueFrom: &core_v1.EnvVarSource{
								ConfigMapKeyRef: &core_v1.ConfigMapKeySelector{
									LocalObjectReference: core_v1.LocalObjectReference{
										Name: strings.TrimSpace(source[2]),
									},
									Key: strings.TrimSpace(source[3]),
								},
							},
						})
					}
				case "fieldRef":
					{
						if len(source) != 3 {
							return v1alpha1.Service{}, errors.New(fmt.Sprintf("unable to parse 'valueFrom' entry '%v', it should be provided as valueFrom:fieldRef:{field-path}", env))
						}
						envVars = append(envVars, core_v1.EnvVar{
							Name: envKey,
							ValueFrom: &core_v1.EnvVarSource{
								FieldRef: &core_v1.ObjectFieldSelector{
										FieldPath: strings.TrimSpace(source[2]),
								},
							},
						})
					}
				case "resourceFieldRef":
					{
						if len(source) < 4 || len(source) > 5 {
							return v1alpha1.Service{}, errors.New(fmt.Sprintf("unable to parse 'valueFrom' entry '%v', it should be provided as valueFrom:resourceFieldRef:{containerName}:{resource}:[{divisor}]", env))
						}
						divisor := resource.Quantity{}
						if len(source) == 5 {
							err := divisor.UnmarshalJSON([]byte(source[4]))
							if err != nil {
								return v1alpha1.Service{}, errors.New(fmt.Sprintf("failed parsing the divisor in 'valueFrom' entry '%v'\n  %s", env, err.Error()))
							}
						}
						envVars = append(envVars, core_v1.EnvVar{
							Name: envKey,
							ValueFrom: &core_v1.EnvVarSource{
								ResourceFieldRef: &core_v1.ResourceFieldSelector{
										ContainerName: strings.TrimSpace(source[2]),
										Resource: strings.TrimSpace(source[3]),
										Divisor: divisor,
								},
							},
						})
					}
				default:
					return v1alpha1.Service{}, errors.New(fmt.Sprintf("unable to parse 'valueFrom' entry '%v', the only accepted source types are secretKeyRef, configMapKeyRef, fieldRef and resourceFieldRef", env))
				}
			} else {
				envVars = append(envVars, core_v1.EnvVar{Name: envEntry[0], Value: envEntry[1]})
			}
		}
		s.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Env = envVars
	}

	return s, nil
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
	var ingress string
	ingresses := ksvc.Status.LoadBalancer.Ingress
	if len(ingresses) > 0 {
		ingress = ingresses[0].IP
		if ingress == "" {
			ingress = ingresses[0].Hostname
		}
	}
	if ingress == "" {
		for _, port := range ksvc.Spec.Ports {
			if port.Name == "http" {
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

	s, err := c.service(options.Namespaced, options.Name)
	if err != nil {
		return "", "", err
	}

	return ingress, s.Status.Domain, nil
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
