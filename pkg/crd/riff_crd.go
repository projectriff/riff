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

package crd

import (
	extClientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	extApi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func CreateCRD(clientset extClientset.Interface) error {
	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(
		&extApi.CustomResourceDefinition{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: "riffsystem.projectriff.io",
			},
			Spec: extApi.CustomResourceDefinitionSpec{
				Group: "projectriff.io",
				Versions: []extApi.CustomResourceDefinitionVersion {
					{
						Name:    "v1alpha1",
						Served:  true,
						Storage: true,
					},
				},
				Scope: extApi.ClusterScoped,
				Names: extApi.CustomResourceDefinitionNames{
					Singular: "riff-system",
					Kind: "RiffSystem",
					ShortNames: []string{
						"riff",
					},
				},
			},
		})

	if err != nil && apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}
