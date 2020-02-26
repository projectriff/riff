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

package v1alpha1

import (
	kpackv1alpha1 "github.com/projectriff/riff/system/pkg/apis/thirdparty/kpack/build/v1alpha1"
)

// +k8s:deepcopy-gen=false
type Source = kpackv1alpha1.SourceConfig

// +k8s:deepcopy-gen=false
type Git = kpackv1alpha1.Git

// +k8s:deepcopy-gen=false
type Blob = kpackv1alpha1.Blob

// +k8s:deepcopy-gen=false
type Registry = kpackv1alpha1.Registry

type ImageTaggingStrategy = kpackv1alpha1.ImageTaggingStrategy

const None = kpackv1alpha1.None
const BuildNumber = kpackv1alpha1.BuildNumber

// +k8s:deepcopy-gen=false
type ImageBuild = kpackv1alpha1.ImageBuild
