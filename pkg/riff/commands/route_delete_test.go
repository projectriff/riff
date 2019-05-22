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
	"testing"

	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/riff/commands"
	rifftesting "github.com/projectriff/riff/pkg/testing"
	requestv1alpha1 "github.com/projectriff/system/pkg/apis/request/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestRouteDeleteOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid delete",
			Options: &commands.RouteDeleteOptions{
				DeleteOptions: rifftesting.InvalidDeleteOptions,
			},
			ExpectFieldError: rifftesting.InvalidDeleteOptionsFieldError,
		},
		{
			Name: "valid delete",
			Options: &commands.RouteDeleteOptions{
				DeleteOptions: rifftesting.ValidDeleteOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestRouteDeleteCommand(t *testing.T) {
	routeName := "test-route"
	routeOtherName := "test-other-route"
	defaultNamespace := "default"

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "delete all routes",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Name:      routeName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeleteCollections: []rifftesting.DeleteCollectionRef{{
				Group:     "request.projectriff.io",
				Resource:  "routes",
				Namespace: defaultNamespace,
			}},
			ExpectOutput: `
Deleted routes in namespace "default"
`,
		},
		{
			Name: "delete all routes error",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Name:      routeName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("delete-collection", "routes"),
			},
			ExpectDeleteCollections: []rifftesting.DeleteCollectionRef{{
				Group:     "request.projectriff.io",
				Resource:  "routes",
				Namespace: defaultNamespace,
			}},
			ShouldError: true,
		},
		{
			Name: "delete route",
			Args: []string{routeName},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Name:      routeName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "request.projectriff.io",
				Resource:  "routes",
				Namespace: defaultNamespace,
				Name:      routeName,
			}},
			ExpectOutput: `
Deleted route "test-route"
`,
		},
		{
			Name: "delete routes",
			Args: []string{routeName, routeOtherName},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Name:      routeName,
						Namespace: defaultNamespace,
					},
				},
				&requestv1alpha1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Name:      routeOtherName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "request.projectriff.io",
				Resource:  "routes",
				Namespace: defaultNamespace,
				Name:      routeName,
			}, {
				Group:     "request.projectriff.io",
				Resource:  "routes",
				Namespace: defaultNamespace,
				Name:      routeOtherName,
			}},
			ExpectOutput: `
Deleted route "test-route"
Deleted route "test-other-route"
`,
		},
		{
			Name: "route does not exist",
			Args: []string{routeName},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "request.projectriff.io",
				Resource:  "routes",
				Namespace: defaultNamespace,
				Name:      routeName,
			}},
			ShouldError: true,
		},
		{
			Name: "delete error",
			Args: []string{routeName},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Name:      routeName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("delete", "routes"),
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "request.projectriff.io",
				Resource:  "routes",
				Namespace: defaultNamespace,
				Name:      routeName,
			}},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewRouteDeleteCommand)
}
