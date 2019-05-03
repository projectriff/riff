/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"fmt"

	"github.com/projectriff/riff/pkg/env"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	BuildConfigMapName    = "riff-build"
	DefaultImagePrefixKey = "default-image-prefix"
	builderImageParamName = "BUILDER_IMAGE"
)

type PackConfig struct {
	BuilderImage string
}

func (c *client) FetchPackConfig() (*PackConfig, error) {
	template, err := c.build.BuildV1alpha1().ClusterBuildTemplates().Get(buildTemplateName, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	config := &PackConfig{}
	for _, param := range template.Spec.Parameters {
		if param.Default == nil {
			continue
		}
		switch param.Name {
		case builderImageParamName:
			config.BuilderImage = *param.Default
		}
	}
	if config.BuilderImage == "" {
		// should never get here
		return nil, fmt.Errorf("unable to find builder image in cluster")
	}
	return config, nil
}

func (c *client) DefaultBuildImagePrefix(namespace string) (string, error) {
	cm, err := c.buildConfigMap(namespace)
	if err != nil {
		return "", err
	}
	return cm.Data[DefaultImagePrefixKey], nil
}

func (c *client) SetDefaultBuildImagePrefix(namespace, prefix string) error {
	cm, err := c.buildConfigMap(namespace)
	if err != nil {
		return err
	}
	cm.Data[DefaultImagePrefixKey] = prefix
	fmt.Printf("Setting default image prefix to %q for namespace %q\n", prefix, namespace)
	return c.saveBuildConfigMap(cm)
}

func (c *client) buildConfigMap(namespace string) (*core_v1.ConfigMap, error) {
	ns := c.explicitOrConfigNamespace(namespace)
	cm, err := c.kubeClient.CoreV1().ConfigMaps(ns).Get(BuildConfigMapName, meta_v1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}
		cm = &core_v1.ConfigMap{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      BuildConfigMapName,
				Namespace: ns,
				Labels: map[string]string{
					"projectriff.io/installer": env.Cli.Name,
					"projectriff.io/version":   env.Cli.Version,
				},
			},
			Data: map[string]string{},
		}
	}
	return cm, nil
}

func (c *client) saveBuildConfigMap(cm *core_v1.ConfigMap) error {
	if cm.UID == "" {
		_, err := c.kubeClient.CoreV1().ConfigMaps(cm.Namespace).Create(cm)
		return err
	}
	_, err := c.kubeClient.CoreV1().ConfigMaps(cm.Namespace).Update(cm)
	return err
}
