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
	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/testing"
)

func TestListOptions(t *testing.T) {
	table := testing.OptionsTable{
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
			ExpectFieldError: cli.ErrMissingOneOf("namespace", "all-namespaces"),
		},
		{
			Name: "both",
			Options: &cli.ListOptions{
				Namespace:     "default",
				AllNamespaces: true,
			},
			ExpectFieldError: cli.ErrMultipleOneOf("namespace", "all-namespaces"),
		},
	}

	table.Run(t)
}

func TestResourceOptions(t *testing.T) {
	table := testing.OptionsTable{
		{
			Name:    "default",
			Options: &cli.ResourceOptions{},
			ExpectFieldError: (&cli.FieldError{}).Also(
				cli.ErrMissingField("namespace"),
				cli.ErrMissingField("name"),
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
			ExpectFieldError: cli.ErrMissingField("namespace"),
		},
		{
			Name: "missing name",
			Options: &cli.ResourceOptions{
				Namespace: "default",
			},
			ExpectFieldError: cli.ErrMissingField("name"),
		},
	}

	table.Run(t)
}

func TestDeleteOptions(t *testing.T) {
	table := testing.OptionsTable{
		{
			Name: "default",
			Options: &cli.DeleteOptions{
				Namespace: "default",
			},
			ExpectFieldError: cli.ErrMissingOneOf("all", "names"),
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
			ExpectFieldError: cli.ErrInvalidValue("my.function", cli.CurrentField).ViaFieldIndex("names", 0),
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
			ExpectFieldError: cli.ErrMultipleOneOf("all", "names"),
		},
	}

	table.Run(t)
}
