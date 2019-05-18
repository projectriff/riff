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

func TestStreamDeleteOptions(t *testing.T) {
	table := testing.OptionsTable{
		{
			Name: "invalid delete",
			Options: &commands.StreamDeleteOptions{
				DeleteOptions: testing.InvalidDeleteOptions,
			},
			ExpectFieldError: testing.InvalidDeleteOptionsFieldError,
		},
		{
			Name: "valid delete",
			Options: &commands.StreamDeleteOptions{
				DeleteOptions: testing.ValidDeleteOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestStreamDeleteCommand(t *testing.T) {
	streamName := "test-stream"
	streamOtherName := "test-other-stream"
	defaultNamespace := "default"

	table := testing.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "delete all streams",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Stream{
					ObjectMeta: metav1.ObjectMeta{
						Name:      streamName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeleteCollections: []testing.DeleteCollectionRef{{
				Group:     "stream.projectriff.io",
				Resource:  "streams",
				Namespace: defaultNamespace,
			}},
		},
		{
			Name: "delete stream",
			Args: []string{streamName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Stream{
					ObjectMeta: metav1.ObjectMeta{
						Name:      streamName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []testing.DeleteRef{{
				Group:     "stream.projectriff.io",
				Resource:  "streams",
				Namespace: defaultNamespace,
				Name:      streamName,
			}},
		},
		{
			Name: "delete streams",
			Args: []string{streamName, streamOtherName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Stream{
					ObjectMeta: metav1.ObjectMeta{
						Name:      streamName,
						Namespace: defaultNamespace,
					},
				},
				&streamv1alpha1.Stream{
					ObjectMeta: metav1.ObjectMeta{
						Name:      streamOtherName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []testing.DeleteRef{{
				Group:     "stream.projectriff.io",
				Resource:  "streams",
				Namespace: defaultNamespace,
				Name:      streamName,
			}, {
				Group:     "stream.projectriff.io",
				Resource:  "streams",
				Namespace: defaultNamespace,
				Name:      streamOtherName,
			}},
		},
		{
			Name: "stream does not exist",
			Args: []string{streamName},
			ExpectDeletes: []testing.DeleteRef{{
				Group:     "stream.projectriff.io",
				Resource:  "streams",
				Namespace: defaultNamespace,
				Name:      streamName,
			}},
			ShouldError: true,
		},
		{
			Name: "delete error",
			Args: []string{streamName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Stream{
					ObjectMeta: metav1.ObjectMeta{
						Name:      streamName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []testing.ReactionFunc{
				testing.InduceFailure("delete", "streams"),
			},
			ExpectDeletes: []testing.DeleteRef{{
				Group:     "stream.projectriff.io",
				Resource:  "streams",
				Namespace: defaultNamespace,
				Name:      streamName,
			}},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewStreamDeleteCommand)
}
