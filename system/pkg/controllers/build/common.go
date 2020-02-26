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

package build

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
)

const riffBuildServiceAccount = "riff-build"

var errMissingDefaultPrefix = fmt.Errorf("missing default image prefix")

// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch

func resolveTargetImage(ctx context.Context, client client.Client, build buildv1alpha1.ImageResource) (string, error) {
	if !strings.HasPrefix(build.GetImage(), "_") {
		return build.GetImage(), nil
	}

	var riffBuildConfig corev1.ConfigMap
	if err := client.Get(ctx, types.NamespacedName{Namespace: build.GetNamespace(), Name: riffBuildServiceAccount}, &riffBuildConfig); err != nil {
		if apierrs.IsNotFound(err) {
			return "", errMissingDefaultPrefix
		}
		return "", err
	}
	defaultPrefix := riffBuildConfig.Data["default-image-prefix"]
	if defaultPrefix == "" {
		return "", errMissingDefaultPrefix
	}
	image, err := buildv1alpha1.ResolveDefaultImage(build, defaultPrefix)
	if err != nil {
		return "", err
	}
	return image, nil
}
