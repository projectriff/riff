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
	"strconv"

	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/knative/serving/pkg/reconciler/v1alpha1/configuration/resources"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (c *client) bumpNonceAnnotationForRevision(s *v1alpha1.Service) {
	annotations := s.Spec.RunLatest.Configuration.RevisionTemplate.Annotations
	if annotations == nil {
		annotations = map[string]string{}
	}
	build := annotations[nonceAnnotation]
	i, err := strconv.Atoi(build)
	if err != nil {
		i = 0
	}
	annotations[nonceAnnotation] = strconv.Itoa(i + 1)
	s.Spec.RunLatest.Configuration.RevisionTemplate.SetAnnotations(annotations)
}

func (c *client) bumpNonceAnnotationForBuild(s *v1alpha1.Service) {
	configurationSpec := s.Spec.RunLatest.Configuration
	build := getBuild(configurationSpec)
	if build != nil {
		annotations := build.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}
		nonce := annotations[nonceAnnotation]
		i, err := strconv.Atoi(nonce)
		if err != nil {
			i = 0
		}
		annotations[nonceAnnotation] = strconv.Itoa(i + 1)
		build.SetAnnotations(annotations)
		s.Spec.RunLatest.Configuration.Build = &v1alpha1.RawExtension{
			Object: build,
		}
	}
}

func getBuild(configSpec v1alpha1.ConfigurationSpec) *unstructured.Unstructured {
	if configSpec.Build == nil {
		return nil
	}
	u := resources.GetBuild(&configSpec)
	return u
}
