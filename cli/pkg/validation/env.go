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

	"github.com/projectriff/cli/pkg/cli"
)

func EnvVar(env, field string) cli.FieldErrors {
	errs := cli.FieldErrors{}

	if strings.HasPrefix(env, "=") || !strings.Contains(env, "=") {
		errs = errs.Also(cli.ErrInvalidValue(env, field))
	}

	return errs
}

func EnvVars(envs []string, field string) cli.FieldErrors {
	errs := cli.FieldErrors{}

	for i, env := range envs {
		errs = errs.Also(EnvVar(env, cli.CurrentField).ViaFieldIndex(field, i))
	}

	return errs
}

func EnvVarFrom(env, field string) cli.FieldErrors {
	errs := cli.FieldErrors{}

	parts := strings.SplitN(env, "=", 2)
	if len(parts) != 2 || parts[0] == "" {
		errs = errs.Also(cli.ErrInvalidValue(env, field))
	} else {
		value := strings.SplitN(parts[1], ":", 3)
		if len(value) != 3 {
			errs = errs.Also(cli.ErrInvalidValue(env, field))
		} else if value[0] != "configMapKeyRef" && value[0] != "secretKeyRef" {
			errs = errs.Also(cli.ErrInvalidValue(env, field))
		} else if value[1] == "" {
			errs = errs.Also(cli.ErrInvalidValue(env, field))
		}
	}

	return errs
}

func EnvVarFroms(envs []string, field string) cli.FieldErrors {
	errs := cli.FieldErrors{}

	for i, env := range envs {
		errs = errs.Also(EnvVarFrom(env, cli.CurrentField).ViaFieldIndex(field, i))
	}

	return errs
}
