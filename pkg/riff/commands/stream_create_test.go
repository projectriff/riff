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

package commands_test

import (
	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/riff/commands"
	"github.com/projectriff/riff/pkg/testing"
	streamv1alpha1 "github.com/projectriff/system/pkg/apis/stream/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestStreamCreateOptions(t *testing.T) {
	table := testing.OptionsTable{
		{
			Name: "invalid resource",
			Options: &commands.StreamCreateOptions{
				ResourceOptions: testing.InvalidResourceOptions,
			},
			ExpectFieldError: testing.InvalidResourceOptionsFieldError.Also(
				cli.ErrMissingField(cli.ProviderFlagName),
			),
		},
		{
			Name: "valid provider",
			Options: &commands.StreamCreateOptions{
				ResourceOptions: testing.ValidResourceOptions,
				Provider:        "test-provider",
			},
			ShouldValidate: true,
		},
		{
			Name: "no provider",
			Options: &commands.StreamCreateOptions{
				ResourceOptions: testing.ValidResourceOptions,
			},
			ExpectFieldError: cli.ErrMissingField(cli.ProviderFlagName),
		},
	}

	table.Run(t)
}

func TestStreamCreateCommand(t *testing.T) {
	defaultNamespace := "default"
	streamName := "my-stream"
	provider := "test-provider"

	table := testing.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "stream provider",
			Args: []string{streamName, cli.ProviderFlagName, provider},
			ExpectCreates: []runtime.Object{
				&streamv1alpha1.Stream{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      streamName,
					},
					Spec: streamv1alpha1.StreamSpec{
						Provider: provider,
					},
				},
			},
			ExpectOutput: `
Created stream "my-stream"
`,
		},
		{
			Name: "error existing stream",
			Args: []string{streamName, cli.ProviderFlagName, provider},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Stream{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      streamName,
					},
				},
			},
			ExpectCreates: []runtime.Object{
				&streamv1alpha1.Stream{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      streamName,
					},
					Spec: streamv1alpha1.StreamSpec{
						Provider: provider,
					},
				},
			},
			ShouldError: true,
		},
		{
			Name: "error during create",
			Args: []string{streamName, cli.ProviderFlagName, provider},
			WithReactors: []testing.ReactionFunc{
				testing.InduceFailure("create", "streams"),
			},
			ExpectCreates: []runtime.Object{
				&streamv1alpha1.Stream{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      streamName,
					},
					Spec: streamv1alpha1.StreamSpec{
						Provider: provider,
					},
				},
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewStreamCreateCommand)
}