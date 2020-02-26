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
	"context"
	"testing"

	"github.com/projectriff/riff/cli/pkg/build/commands"
	"github.com/projectriff/riff/cli/pkg/cli"
	rifftesting "github.com/projectriff/riff/cli/pkg/testing"
	"github.com/projectriff/riff/system/pkg/apis"
	buildv1alpha1 "github.com/projectriff/riff/system/pkg/apis/build/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestApplicationListOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid list",
			Options: &commands.ApplicationListOptions{
				ListOptions: rifftesting.InvalidListOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidListOptionsFieldError,
		},
		{
			Name: "valid list",
			Options: &commands.ApplicationListOptions{
				ListOptions: rifftesting.ValidListOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestApplicationListCommand(t *testing.T) {
	applicationName := "test-application"
	applicationOtherName := "test-other-application"
	defaultNamespace := "default"
	otherNamespace := "other-namespace"

	table := rifftesting.CommandTable{
		{
			Name: "invalid args",
			Args: []string{},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				// disable default namespace
				c.Client.(*rifftesting.FakeClient).Namespace = ""
				return ctx, nil
			},
			ShouldError: true,
		},
		{
			Name: "empty",
			Args: []string{},
			ExpectOutput: `
No applications found.
`,
		},
		{
			Name: "lists an item",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Name:      applicationName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectOutput: `
NAME               LATEST IMAGE   STATUS      AGE
test-application   <empty>        <unknown>   <unknown>
`,
		},
		{
			Name: "filters by namespace",
			Args: []string{cli.NamespaceFlagName, otherNamespace},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Name:      applicationName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectOutput: `
No applications found.
`,
		},
		{
			Name: "all namespace",
			Args: []string{cli.AllNamespacesFlagName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Name:      applicationName,
						Namespace: defaultNamespace,
					},
				},
				&buildv1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Name:      applicationOtherName,
						Namespace: otherNamespace,
					},
				},
			},
			ExpectOutput: `
NAMESPACE         NAME                     LATEST IMAGE   STATUS      AGE
default           test-application         <empty>        <unknown>   <unknown>
other-namespace   test-other-application   <empty>        <unknown>   <unknown>
`,
		},
		{
			Name: "table populates all columns",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "petclinic",
						Namespace: defaultNamespace,
					},
					Spec: buildv1alpha1.ApplicationSpec{
						Image: "projectriff/petclinic",
					},
					Status: buildv1alpha1.ApplicationStatus{
						Status: apis.Status{
							Conditions: apis.Conditions{
								{Type: buildv1alpha1.ApplicationConditionReady, Status: "True"},
							},
						},
						BuildStatus: buildv1alpha1.BuildStatus{
							LatestImage: "projectriff/petclinic@sah256:abcdef1234",
						},
					},
				},
			},
			ExpectOutput: `
NAME        LATEST IMAGE                              STATUS   AGE
petclinic   projectriff/petclinic@sah256:abcdef1234   Ready    <unknown>
`,
		},
		{
			Name: "list error",
			Args: []string{},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("list", "applications"),
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewApplicationListCommand)
}
