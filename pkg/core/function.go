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
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
)

type CreateFunctionOptions struct {
	CreateServiceOptions

	GitRepo     string
	GitRevision string

	InvokerURL string
	Handler    string
	Artifact   string
}

func (c *client) CreateFunction(options CreateFunctionOptions) (*v1alpha1.Service, error) {
	ns := c.explicitOrConfigNamespace(options.Namespaced)

	s := newService(options.CreateServiceOptions)

	s.Spec.RunLatest.Configuration.Build = &build.BuildSpec{}

	s.Spec.RunLatest.Configuration.Build = &build.BuildSpec{
		ServiceAccountName: "riff-build",
		Source: &build.SourceSpec{
			Git: &build.GitSourceSpec{
				Url:      options.GitRepo,
				Revision: options.GitRevision,
			},
		},
		Template: &build.TemplateInstantiationSpec{
			Name: "riff",
			Arguments: []build.ArgumentSpec{
				{Name: "IMAGE", Value: options.Image},
				{Name: "INVOKER_PATH", Value: options.InvokerURL},
				{Name: "FUNCTION_ARTIFACT", Value: options.Artifact},
				{Name: "FUNCTION_HANDLER", Value: options.Handler},
				{Name: "FUNCTION_NAME", Value: options.Name},
			},
		},
	}

	_, err := c.serving.ServingV1alpha1().Services(ns).Create(&s)

	return &s, err
}
