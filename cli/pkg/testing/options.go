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

package testing

import (
	"github.com/projectriff/riff/cli/pkg/cli"
	"github.com/projectriff/riff/cli/pkg/cli/options"
)

const TestField = "test-field"

var (
	ValidListOptions = options.ListOptions{
		Namespace: "default",
	}
	InvalidListOptions           = options.ListOptions{}
	InvalidListOptionsFieldError = cli.ErrMissingOneOf(cli.NamespaceFlagName, cli.AllNamespacesFlagName)
)

var (
	ValidResourceOptions = options.ResourceOptions{
		Namespace: "default",
		Name:      "push-credentials",
	}
	InvalidResourceOptions           = options.ResourceOptions{}
	InvalidResourceOptionsFieldError = cli.FieldErrors{}.Also(
		cli.ErrMissingField(cli.NamespaceFlagName),
		cli.ErrMissingField(cli.NameArgumentName),
	)
)

var (
	ValidDeleteOptions = options.DeleteOptions{
		Namespace: "default",
		Names:     []string{"my-resource"},
	}
	InvalidDeleteOptions = options.DeleteOptions{
		Namespace: "default",
	}
	InvalidDeleteOptionsFieldError = cli.ErrMissingOneOf(cli.AllFlagName, cli.NamesArgumentName)
)
