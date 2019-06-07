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

package cli_test

import (
	"testing"

	"github.com/projectriff/riff/pkg/cli"
	rifftesting "github.com/projectriff/riff/pkg/testing"
)

func TestListOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "default",
			Options: &cli.ListOptions{
				Namespace: "default",
			},
			ShouldValidate: true,
		},
		{
			Name: "all namespaces",
			Options: &cli.ListOptions{
				AllNamespaces: true,
			},
			ShouldValidate: true,
		},
		{
			Name:             "neither",
			Options:          &cli.ListOptions{},
			ExpectFieldError: cli.ErrMissingOneOf(cli.NamespaceFlagName, cli.AllNamespacesFlagName),
		},
		{
			Name: "both",
			Options: &cli.ListOptions{
				Namespace:     "default",
				AllNamespaces: true,
			},
			ExpectFieldError: cli.ErrMultipleOneOf(cli.NamespaceFlagName, cli.AllNamespacesFlagName),
		},
	}

	table.Run(t)
}

func TestResourceOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name:    "default",
			Options: &cli.ResourceOptions{},
			ExpectFieldError: cli.EmptyFieldError.Also(
				cli.ErrMissingField(cli.NamespaceFlagName),
				cli.ErrMissingField(cli.NameArgumentName),
			),
		},
		{
			Name: "has both",
			Options: &cli.ResourceOptions{
				Namespace: "default",
				Name:      "push-credentials",
			},
			ShouldValidate: true,
		},
		{
			Name: "missing namespace",
			Options: &cli.ResourceOptions{
				Name: "push-credentials",
			},
			ExpectFieldError: cli.ErrMissingField(cli.NamespaceFlagName),
		},
		{
			Name: "missing name",
			Options: &cli.ResourceOptions{
				Namespace: "default",
			},
			ExpectFieldError: cli.ErrMissingField(cli.NameArgumentName),
		},
	}

	table.Run(t)
}

func TestDeleteOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "default",
			Options: &cli.DeleteOptions{
				Namespace: "default",
			},
			ExpectFieldError: cli.ErrMissingOneOf(cli.AllFlagName, cli.NamesArgumentName),
		},
		{
			Name: "single name",
			Options: &cli.DeleteOptions{
				Namespace: "default",
				Names:     []string{"my-function"},
			},
			ShouldValidate: true,
		},
		{
			Name: "multiple names",
			Options: &cli.DeleteOptions{
				Namespace: "default",
				Names:     []string{"my-function", "my-other-function"},
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid name",
			Options: &cli.DeleteOptions{
				Namespace: "default",
				Names:     []string{"my.function"},
			},
			ExpectFieldError: cli.ErrInvalidValue("my.function", cli.CurrentField).ViaFieldIndex(cli.NamesArgumentName, 0),
		},
		{
			Name: "all",
			Options: &cli.DeleteOptions{
				Namespace: "default",
				All:       true,
			},
			ShouldValidate: true,
		},
		{
			Name: "all with name",
			Options: &cli.DeleteOptions{
				Namespace: "default",
				Names:     []string{"my-function"},
				All:       true,
			},
			ExpectFieldError: cli.ErrMultipleOneOf(cli.AllFlagName, cli.NamesArgumentName),
		},
	}

	table.Run(t)
}
