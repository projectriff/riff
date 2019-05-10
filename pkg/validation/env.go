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

package validation

import (
	"strings"

	"github.com/projectriff/riff/pkg/cli"
)

func EnvVar(env, field string) *cli.FieldError {
	errs := &cli.FieldError{}

	if strings.HasPrefix(env, "=") || !strings.Contains(env, "=") {
		errs = errs.Also(cli.ErrInvalidValue(env, field))
	}

	return errs
}

func EnvVars(envs []string, field string) *cli.FieldError {
	errs := &cli.FieldError{}

	for i, env := range envs {
		errs = errs.Also(EnvVar(env, cli.CurrentField).ViaFieldIndex(field, i))
	}

	return errs
}
