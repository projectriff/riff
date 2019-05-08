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
	"github.com/projectriff/riff/pkg/riff/commands"
	"github.com/projectriff/riff/pkg/testing"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestCredentialDeleteCommand(t *testing.T) {
	table := testing.CommandTable{{
		Name: "delete secret",
		Args: []string{"my-credential"},
		GivenObjects: []runtime.Object{
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-credential",
					Namespace: "default",
				},
				StringData: map[string]string{},
			},
		},
		ExpectDeletes: []testing.DeleteActionImpl{{
			Name: "my-credential",
			ActionImpl: testing.ActionImpl{
				Namespace: "default",
				Verb:      "delete",
				Resource: schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "secrets",
				},
			}},
		},
	}, {
		Name: "secret ds't exist",
		Args: []string{"my-credential"},
		ExpectDeletes: []testing.DeleteActionImpl{{
			Name: "my-credential",
			ActionImpl: testing.ActionImpl{
				Namespace: "default",
				Verb:      "delete",
				Resource: schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "secrets",
				},
			}},
		},
		ExpectError: true,
	}, {
		Name: "delete error",
		Args: []string{"my-credential"},
		GivenObjects: []runtime.Object{
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-credential",
					Namespace: "default",
				},
				StringData: map[string]string{},
			},
		},
		WithReactors: []testing.ReactionFunc{
			testing.InduceFailure("delete", "secrets"),
		},
		ExpectDeletes: []testing.DeleteActionImpl{{
			Name: "my-credential",
			ActionImpl: testing.ActionImpl{
				Namespace: "default",
				Verb:      "delete",
				Resource: schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "secrets",
				},
			}},
		},
		ExpectError: true,
	}}

	table.Run(t, commands.NewCredentialDeleteCommand)
}
