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

package v1alpha1

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ImageResource interface {
	metav1.ObjectMetaAccessor
	GetImage() string
}

func ResolveDefaultImage(resource ImageResource, registry string) (string, error) {
	if registry == "" {
		return "", fmt.Errorf("invalid registry %q", registry)
	}
	image := resource.GetImage()
	if image == "_" {
		// combine registry prefix and application name
		image = fmt.Sprintf("%s/%s", registry, resource.GetObjectMeta().GetName())
	} else if strings.HasPrefix(image, "_/") {
		// add the prefix to the specified image name
		image = strings.Replace(image, "_", registry, 1)
	} else {
		return "", fmt.Errorf("unable to default registry")
	}
	return image, nil
}

type BuildArgument struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Source struct {
	Git     *GitSource `json:"git"`
	SubPath string     `json:"subPath,omitempty"`
}

type GitSource struct {
	Revision string `json:"revision"`
	URL      string `json:"url"`
}

type BuildStatus struct {
	BuildCacheName string `json:"buildCacheName,omitempty"`
	BuildName      string `json:"buildName,omitempty"`
	LatestImage    string `json:"latestImage,omitempty"`
}
