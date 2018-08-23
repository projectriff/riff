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
	"errors"
	"fmt"
	"k8s.io/api/core/v1"
	"strings"
)

func ParseEnvVar(envVars []string) ([]v1.EnvVar, error) {
	var results []v1.EnvVar
	for _, env := range envVars {
		envEntry, err := splitEnvVarEntry(env)
		if err != nil {
			return []v1.EnvVar{}, err
		}
		results = append(results, v1.EnvVar{Name: envEntry[0], Value: envEntry[1]})
	}
	return results, nil
}

func ParseEnvVarSource(envVarsFrom []string) ([]v1.EnvVar, error) {
	var results []v1.EnvVar
	for _, env := range envVarsFrom {
		envEntry, err := splitEnvVarEntry(env)
		if err != nil {
			return []v1.EnvVar{}, err
		}
		source := strings.Split(envEntry[1], ":")
		sourceType := strings.TrimSpace(source[0])
		switch sourceType {
		case "secretKeyRef":
			if len(source) != 3 {
				return []v1.EnvVar{}, errors.New(fmt.Sprintf("unable to parse 'env-from' entry '%s', it should be provided as secretKeyRef:{secret-name}:{key-to-select}", env))
			}
			results = append(results, v1.EnvVar{
				Name: envEntry[0],
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: source[1],
						},
						Key: source[2],
					},
				},
			})
		case "configMapKeyRef":
			if len(source) != 3 {
				return []v1.EnvVar{}, errors.New(fmt.Sprintf("unable to parse 'env-from' entry '%s', it should be provided as configMapKeyRef:{config-map-name}:{key-to-select}", env))
			}
			results = append(results, v1.EnvVar{
				Name: envEntry[0],
				ValueFrom: &v1.EnvVarSource{
					ConfigMapKeyRef: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: source[1],
						},
						Key: source[2],
					},
				},
			})
		default:
			return []v1.EnvVar{}, errors.New(fmt.Sprintf("unable to parse 'env-from' entry '%s', the only accepted source types are secretKeyRef and configMapKeyRef", env))
		}
	}
	return results, nil
}

func splitEnvVarEntry(env string) ([]string, error) {
	envEntry := strings.SplitN(env, "=", 2)
	if len(envEntry) != 2 {
		return nil, errors.New(fmt.Sprintf("unable to parse '%s', environment variables must be provided as 'key=value'", env))
	}
	if len(envEntry[0]) < 1 {
		return nil, errors.New(fmt.Sprintf("unable to parse '%s', the key part is missing", env))
	}
	return envEntry, nil
}
