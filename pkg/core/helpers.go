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

package core

import (
	"fmt"

	"github.com/projectriff/riff/pkg/env"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	buildConfigMap        = "riff-build"
	defaultImagePrefixKey = "default-image-prefix"
)

func (c *client) DefaultBuildImagePrefix(namespace string) (string, error) {
	cm, err := c.buildConfigMap(namespace)
	if err != nil {
		return "", err
	}
	return cm.Data[defaultImagePrefixKey], nil
}

func (c *client) SetDefaultBuildImagePrefix(namespace, prefix string) error {
	cm, err := c.buildConfigMap(namespace)
	if err != nil {
		return err
	}
	cm.Data[defaultImagePrefixKey] = prefix
	fmt.Printf("Setting default image prefix to %q for namespace %q\n", prefix, namespace)
	return c.saveBuildConfigMap(cm)
}

func (c *client) buildConfigMap(namespace string) (*core_v1.ConfigMap, error) {
	ns := c.explicitOrConfigNamespace(namespace)
	cm, err := c.kubeClient.CoreV1().ConfigMaps(ns).Get(buildConfigMap, meta_v1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}
		cm = &core_v1.ConfigMap{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      buildConfigMap,
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
