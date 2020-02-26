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

package options_test

import (
	"testing"

	"github.com/projectriff/cli/pkg/cli"
	"github.com/projectriff/cli/pkg/cli/options"
	rifftesting "github.com/projectriff/cli/pkg/testing"
)

func TestListOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "default",
			Options: &options.ListOptions{
				Namespace: "default",
			},
			ShouldValidate: true,
		},
		{
			Name: "all namespaces",
			Options: &options.ListOptions{
				AllNamespaces: true,
			},
			ShouldValidate: true,
		},
		{
			Name:              "neither",
			Options:           &options.ListOptions{},
			ExpectFieldErrors: cli.ErrMissingOneOf(cli.NamespaceFlagName, cli.AllNamespacesFlagName),
		},
		{
			Name: "both",
			Options: &options.ListOptions{
				Namespace:     "default",
				AllNamespaces: true,
			},
			ExpectFieldErrors: cli.ErrMultipleOneOf(cli.NamespaceFlagName, cli.AllNamespacesFlagName),
		},
	}

	table.Run(t)
}

func TestResourceOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name:    "default",
			Options: &options.ResourceOptions{},
			ExpectFieldErrors: cli.FieldErrors{}.Also(
				cli.ErrMissingField(cli.NamespaceFlagName),
				cli.ErrMissingField(cli.NameArgumentName),
			),
		},
		{
			Name: "has both",
			Options: &options.ResourceOptions{
				Namespace: "default",
				Name:      "push-credentials",
			},
			ShouldValidate: true,
		},
		{
			Name: "missing namespace",
			Options: &options.ResourceOptions{
				Name: "push-credentials",
			},
			ExpectFieldErrors: cli.ErrMissingField(cli.NamespaceFlagName),
		},
		{
			Name: "missing name",
			Options: &options.ResourceOptions{
				Namespace: "default",
			},
			ExpectFieldErrors: cli.ErrMissingField(cli.NameArgumentName),
		},
	}

	table.Run(t)
}

func TestDeleteOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "default",
			Options: &options.DeleteOptions{
				Namespace: "default",
			},
			ExpectFieldErrors: cli.ErrMissingOneOf(cli.AllFlagName, cli.NamesArgumentName),
		},
		{
			Name: "single name",
			Options: &options.DeleteOptions{
				Namespace: "default",
				Names:     []string{"my-function"},
			},
			ShouldValidate: true,
		},
		{
			Name: "multiple names",
			Options: &options.DeleteOptions{
				Namespace: "default",
				Names:     []string{"my-function", "my-other-function"},
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid name",
			Options: &options.DeleteOptions{
				Namespace: "default",
				Names:     []string{"my.function"},
			},
			ExpectFieldErrors: cli.ErrInvalidValue("my.function", cli.CurrentField).ViaFieldIndex(cli.NamesArgumentName, 0),
		},
		{
			Name: "all",
			Options: &options.DeleteOptions{
				Namespace: "default",
				All:       true,
			},
			ShouldValidate: true,
		},
		{
			Name: "all with name",
			Options: &options.DeleteOptions{
				Namespace: "default",
				Names:     []string{"my-function"},
				All:       true,
			},
			ExpectFieldErrors: cli.ErrMultipleOneOf(cli.AllFlagName, cli.NamesArgumentName),
		},
		{
			Name: "missing namespace",
			Options: &options.DeleteOptions{
				Names: []string{"my-function"},
			},
			ExpectFieldErrors: cli.ErrMissingField(cli.NamespaceFlagName),
		},
	}

	table.Run(t)
}
