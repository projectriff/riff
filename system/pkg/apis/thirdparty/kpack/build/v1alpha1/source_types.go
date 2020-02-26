/*
Copyright 2019 The original author or authors

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
	corev1 "k8s.io/api/core/v1"
)

type SourceConfig struct {
	Git      *Git      `json:"git,omitempty"`
	Blob     *Blob     `json:"blob,omitempty"`
	Registry *Registry `json:"registry,omitempty"`
	SubPath  string    `json:"subPath,omitempty"`
}

type Git struct {
	URL      string `json:"url"`
	Revision string `json:"revision"`
}

type Blob struct {
	URL string `json:"url"`
}

type Registry struct {
	Image            string                        `json:"image"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,15,rep,name=imagePullSecrets"`
}

type ResolvedSourceConfig struct {
	Git      *ResolvedGitSource      `json:"git,omitempty"`
	Blob     *ResolvedBlobSource     `json:"blob,omitempty"`
	Registry *ResolvedRegistrySource `json:"registry,omitempty"`
}

type GitSourceKind string

const (
	Unknown GitSourceKind = "Unknown"
	Branch  GitSourceKind = "Branch"
	Tag     GitSourceKind = "Tag"
	Commit  GitSourceKind = "Commit"
)

type ResolvedGitSource struct {
	URL      string        `json:"url"`
	Revision string        `json:"commit"`
	SubPath  string        `json:"subPath,omitempty"`
	Type     GitSourceKind `json:"type"`
}

type ResolvedBlobSource struct {
	URL     string `json:"url"`
	SubPath string `json:"subPath,omitempty"`
}

type ResolvedRegistrySource struct {
	Image            string                        `json:"image"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,15,rep,name=imagePullSecrets"`
	SubPath          string                        `json:"subPath,omitempty"`
}
