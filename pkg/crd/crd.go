/*
 * Copyright 2019 The original author or authors
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
	"fmt"
	extApi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	extClientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	Group = "projectriff.io"
	Version = "v1alpha1"
	Kind = "Manifest"
	Name = "manifest"
)

var schemeGroupVersion = schema.GroupVersion{
	Group:   Group,
	Version: Version,
}

func CreateCRD(clientset extClientset.Interface) error {
	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(
		&extApi.CustomResourceDefinition{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: fmt.Sprintf("%s.%s", Name, Group),
			},
			TypeMeta: meta_v1.TypeMeta{
				APIVersion: "apiextensions.k8s.io/v1beta1",
				Kind: "CustomResourceDefinition",
			},
			Spec: extApi.CustomResourceDefinitionSpec{
				Group: Group,
				Versions: []extApi.CustomResourceDefinitionVersion {
					{
						Name:    Version,
						Served:  true,
						Storage: true,
					},
				},
				Scope: extApi.ClusterScoped,
				Names: extApi.CustomResourceDefinitionNames{
					Singular: Name,
					Plural: Name,
					Kind: Kind,
				},
			},
		})

	if err != nil && apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}
