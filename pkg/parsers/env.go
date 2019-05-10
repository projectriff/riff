/*
 * Copyright 2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package parsers

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func EnvVar(str string) corev1.EnvVar {
	parts := strings.SplitN(str, "=", 2)

	return corev1.EnvVar{
		Name:  parts[0],
		Value: parts[1],
	}
}
