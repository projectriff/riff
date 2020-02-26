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
	"fmt"
	"testing"

	"github.com/projectriff/riff/cli/pkg/cli"
	"github.com/projectriff/riff/cli/pkg/riff/commands"
	rifftesting "github.com/projectriff/riff/cli/pkg/testing"
	authorizationv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"
)

func TestDoctorOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "valid",
			Options: &commands.DoctorOptions{
				Namespace: "default",
			},
			ShouldValidate: true,
		},
		{
			Name:              "invalid",
			Options:           &commands.DoctorOptions{},
			ExpectFieldErrors: cli.ErrMissingField(cli.NamespaceFlagName),
		},
	}

	table.Run(t)
}

func TestDoctorCommand(t *testing.T) {
	verbs := []string{"get", "list", "create", "update", "delete", "patch", "watch"}
	readVerbs := []string{"get", "list", "watch"}
	table := rifftesting.CommandTable{
		{
			Name:     "not installed",
			Args:     []string{},
			Runtimes: &[]string{},
			ExpectCreates: merge(
				selfSubjectAccessReviewRequests("riff-system", "builders", "core", "configmaps", "", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "configmaps", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "secrets", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "pods", "", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "pods", "log", readVerbs...),
			),
			WithReactors: []rifftesting.ReactionFunc{
				passAccessReview(),
			},
			ExpectOutput: `
NAMESPACE     STATUS
default       missing
riff-system   missing

RESOURCE                            NAMESPACE     NAME       READ      WRITE
configmaps                          riff-system   builders   allowed   n/a
configmaps                          default       *          allowed   allowed
secrets                             default       *          allowed   allowed
pods                                default       *          allowed   n/a
pods/log                            default       *          allowed   n/a
applications.build.projectriff.io   default       *          missing   missing
containers.build.projectriff.io     default       *          missing   missing
functions.build.projectriff.io      default       *          missing   missing
`,
		},
		{
			Name:     "installed",
			Args:     []string{},
			Runtimes: &[]string{},
			GivenObjects: []runtime.Object{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "riff-system"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "applications.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "containers.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "functions.build.projectriff.io"}},
			},
			ExpectCreates: merge(
				selfSubjectAccessReviewRequests("riff-system", "builders", "core", "configmaps", "", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "configmaps", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "secrets", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "pods", "", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "pods", "log", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "applications", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "containers", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "functions", "", verbs...),
			),
			WithReactors: []rifftesting.ReactionFunc{
				passAccessReview(),
			},
			ExpectOutput: `
NAMESPACE     STATUS
default       ok
riff-system   ok

RESOURCE                            NAMESPACE     NAME       READ      WRITE
configmaps                          riff-system   builders   allowed   n/a
configmaps                          default       *          allowed   allowed
secrets                             default       *          allowed   allowed
pods                                default       *          allowed   n/a
pods/log                            default       *          allowed   n/a
applications.build.projectriff.io   default       *          allowed   allowed
containers.build.projectriff.io     default       *          allowed   allowed
functions.build.projectriff.io      default       *          allowed   allowed
`,
		},
		{
			Name:     "custom namespace",
			Args:     []string{cli.NamespaceFlagName, "my-namespace"},
			Runtimes: &[]string{},
			GivenObjects: []runtime.Object{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "my-namespace"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "riff-system"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "applications.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "containers.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "functions.build.projectriff.io"}},
			},
			ExpectCreates: merge(
				selfSubjectAccessReviewRequests("riff-system", "builders", "core", "configmaps", "", readVerbs...),
				selfSubjectAccessReviewRequests("my-namespace", "", "core", "configmaps", "", verbs...),
				selfSubjectAccessReviewRequests("my-namespace", "", "core", "secrets", "", verbs...),
				selfSubjectAccessReviewRequests("my-namespace", "", "core", "pods", "", readVerbs...),
				selfSubjectAccessReviewRequests("my-namespace", "", "core", "pods", "log", readVerbs...),
				selfSubjectAccessReviewRequests("my-namespace", "", "build.projectriff.io", "applications", "", verbs...),
				selfSubjectAccessReviewRequests("my-namespace", "", "build.projectriff.io", "containers", "", verbs...),
				selfSubjectAccessReviewRequests("my-namespace", "", "build.projectriff.io", "functions", "", verbs...),
			),
			WithReactors: []rifftesting.ReactionFunc{
				passAccessReview(),
			},
			ExpectOutput: `
NAMESPACE      STATUS
my-namespace   ok
riff-system    ok

RESOURCE                            NAMESPACE      NAME       READ      WRITE
configmaps                          riff-system    builders   allowed   n/a
configmaps                          my-namespace   *          allowed   allowed
secrets                             my-namespace   *          allowed   allowed
pods                                my-namespace   *          allowed   n/a
pods/log                            my-namespace   *          allowed   n/a
applications.build.projectriff.io   my-namespace   *          allowed   allowed
containers.build.projectriff.io     my-namespace   *          allowed   allowed
functions.build.projectriff.io      my-namespace   *          allowed   allowed
`,
		},
		{
			Name:     "all runtimes",
			Args:     []string{},
			Runtimes: &cli.AllRuntimes,
			GivenObjects: []runtime.Object{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "riff-system"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "applications.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "containers.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "functions.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "deployers.core.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "processors.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "streams.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "inmemorygateways.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "kafkagateways.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "pulsargateways.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "adapters.knative.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "deployers.knative.projectriff.io"}},
			},
			ExpectCreates: merge(
				selfSubjectAccessReviewRequests("riff-system", "builders", "core", "configmaps", "", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "configmaps", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "secrets", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "pods", "", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "pods", "log", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "applications", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "containers", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "functions", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core.projectriff.io", "deployers", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "streaming.projectriff.io", "processors", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "streaming.projectriff.io", "streams", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "streaming.projectriff.io", "inmemorygateways", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "streaming.projectriff.io", "kafkagateways", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "streaming.projectriff.io", "pulsargateways", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "knative.projectriff.io", "adapters", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "knative.projectriff.io", "deployers", "", verbs...),
			),
			WithReactors: []rifftesting.ReactionFunc{
				passAccessReview(),
			},
			ExpectOutput: `
NAMESPACE     STATUS
default       ok
riff-system   ok

RESOURCE                                    NAMESPACE     NAME       READ      WRITE
configmaps                                  riff-system   builders   allowed   n/a
configmaps                                  default       *          allowed   allowed
secrets                                     default       *          allowed   allowed
pods                                        default       *          allowed   n/a
pods/log                                    default       *          allowed   n/a
applications.build.projectriff.io           default       *          allowed   allowed
containers.build.projectriff.io             default       *          allowed   allowed
functions.build.projectriff.io              default       *          allowed   allowed
deployers.core.projectriff.io               default       *          allowed   allowed
processors.streaming.projectriff.io         default       *          allowed   allowed
streams.streaming.projectriff.io            default       *          allowed   allowed
inmemorygateways.streaming.projectriff.io   default       *          allowed   allowed
kafkagateways.streaming.projectriff.io      default       *          allowed   allowed
pulsargateways.streaming.projectriff.io     default       *          allowed   allowed
adapters.knative.projectriff.io             default       *          allowed   allowed
deployers.knative.projectriff.io            default       *          allowed   allowed
`,
		},
		{
			Name:     "core runtime",
			Args:     []string{},
			Runtimes: &[]string{cli.CoreRuntime},
			GivenObjects: []runtime.Object{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "riff-system"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "applications.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "containers.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "functions.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "deployers.core.projectriff.io"}},
			},
			ExpectCreates: merge(
				selfSubjectAccessReviewRequests("riff-system", "builders", "core", "configmaps", "", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "configmaps", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "secrets", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "pods", "", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "pods", "log", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "applications", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "containers", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "functions", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core.projectriff.io", "deployers", "", verbs...),
			),
			WithReactors: []rifftesting.ReactionFunc{
				passAccessReview(),
			},
			ExpectOutput: `
NAMESPACE     STATUS
default       ok
riff-system   ok

RESOURCE                            NAMESPACE     NAME       READ      WRITE
configmaps                          riff-system   builders   allowed   n/a
configmaps                          default       *          allowed   allowed
secrets                             default       *          allowed   allowed
pods                                default       *          allowed   n/a
pods/log                            default       *          allowed   n/a
applications.build.projectriff.io   default       *          allowed   allowed
containers.build.projectriff.io     default       *          allowed   allowed
functions.build.projectriff.io      default       *          allowed   allowed
deployers.core.projectriff.io       default       *          allowed   allowed
`,
		},
		{
			Name:     "streaming runtime",
			Args:     []string{},
			Runtimes: &[]string{cli.StreamingRuntime},
			GivenObjects: []runtime.Object{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "riff-system"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "applications.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "containers.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "functions.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "processors.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "streams.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "inmemorygateways.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "kafkagateways.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "pulsargateways.streaming.projectriff.io"}},
			},
			ExpectCreates: merge(
				selfSubjectAccessReviewRequests("riff-system", "builders", "core", "configmaps", "", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "configmaps", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "secrets", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "pods", "", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "pods", "log", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "applications", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "containers", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "functions", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "streaming.projectriff.io", "processors", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "streaming.projectriff.io", "streams", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "streaming.projectriff.io", "inmemorygateways", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "streaming.projectriff.io", "kafkagateways", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "streaming.projectriff.io", "pulsargateways", "", verbs...),
			),
			WithReactors: []rifftesting.ReactionFunc{
				passAccessReview(),
			},
			ExpectOutput: `
NAMESPACE     STATUS
default       ok
riff-system   ok

RESOURCE                                    NAMESPACE     NAME       READ      WRITE
configmaps                                  riff-system   builders   allowed   n/a
configmaps                                  default       *          allowed   allowed
secrets                                     default       *          allowed   allowed
pods                                        default       *          allowed   n/a
pods/log                                    default       *          allowed   n/a
applications.build.projectriff.io           default       *          allowed   allowed
containers.build.projectriff.io             default       *          allowed   allowed
functions.build.projectriff.io              default       *          allowed   allowed
processors.streaming.projectriff.io         default       *          allowed   allowed
streams.streaming.projectriff.io            default       *          allowed   allowed
inmemorygateways.streaming.projectriff.io   default       *          allowed   allowed
kafkagateways.streaming.projectriff.io      default       *          allowed   allowed
pulsargateways.streaming.projectriff.io     default       *          allowed   allowed
`,
		},
		{
			Name:     "knative runtime",
			Args:     []string{},
			Runtimes: &[]string{cli.KnativeRuntime},
			GivenObjects: []runtime.Object{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "riff-system"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "applications.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "containers.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "functions.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "adapters.knative.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "deployers.knative.projectriff.io"}},
			},
			ExpectCreates: merge(
				selfSubjectAccessReviewRequests("riff-system", "builders", "core", "configmaps", "", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "configmaps", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "secrets", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "pods", "", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "pods", "log", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "applications", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "containers", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "functions", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "knative.projectriff.io", "adapters", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "knative.projectriff.io", "deployers", "", verbs...),
			),
			WithReactors: []rifftesting.ReactionFunc{
				passAccessReview(),
			},
			ExpectOutput: `
NAMESPACE     STATUS
default       ok
riff-system   ok

RESOURCE                            NAMESPACE     NAME       READ      WRITE
configmaps                          riff-system   builders   allowed   n/a
configmaps                          default       *          allowed   allowed
secrets                             default       *          allowed   allowed
pods                                default       *          allowed   n/a
pods/log                            default       *          allowed   n/a
applications.build.projectriff.io   default       *          allowed   allowed
containers.build.projectriff.io     default       *          allowed   allowed
functions.build.projectriff.io      default       *          allowed   allowed
adapters.knative.projectriff.io     default       *          allowed   allowed
deployers.knative.projectriff.io    default       *          allowed   allowed
`,
		},
		{
			Name:     "read-only access",
			Args:     []string{},
			Runtimes: &[]string{},
			GivenObjects: []runtime.Object{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "riff-system"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "applications.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "containers.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "functions.build.projectriff.io"}},
			},
			ExpectCreates: merge(
				selfSubjectAccessReviewRequests("riff-system", "builders", "core", "configmaps", "", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "configmaps", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "secrets", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "pods", "", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "pods", "log", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "applications", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "containers", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "functions", "", verbs...),
			),
			WithReactors: []rifftesting.ReactionFunc{
				denyAccessReviewOn("*", "create"),
				denyAccessReviewOn("*", "update"),
				denyAccessReviewOn("*", "delete"),
				denyAccessReviewOn("*", "patch"),
				passAccessReview(),
			},
			ExpectOutput: `
NAMESPACE     STATUS
default       ok
riff-system   ok

RESOURCE                            NAMESPACE     NAME       READ      WRITE
configmaps                          riff-system   builders   allowed   n/a
configmaps                          default       *          allowed   denied
secrets                             default       *          allowed   denied
pods                                default       *          allowed   n/a
pods/log                            default       *          allowed   n/a
applications.build.projectriff.io   default       *          allowed   denied
containers.build.projectriff.io     default       *          allowed   denied
functions.build.projectriff.io      default       *          allowed   denied
`,
		},
		{
			Name:     "no-watch access",
			Args:     []string{},
			Runtimes: &[]string{},
			GivenObjects: []runtime.Object{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "riff-system"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "applications.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "containers.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "functions.build.projectriff.io"}},
			},
			ExpectCreates: merge(
				selfSubjectAccessReviewRequests("riff-system", "builders", "core", "configmaps", "", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "configmaps", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "secrets", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "pods", "", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "pods", "log", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "applications", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "containers", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "functions", "", verbs...),
			),
			WithReactors: []rifftesting.ReactionFunc{
				denyAccessReviewOn("*", "watch"),
				passAccessReview(),
			},
			ExpectOutput: `
NAMESPACE     STATUS
default       ok
riff-system   ok

RESOURCE                            NAMESPACE     NAME       READ    WRITE
configmaps                          riff-system   builders   mixed   n/a
configmaps                          default       *          mixed   allowed
secrets                             default       *          mixed   allowed
pods                                default       *          mixed   n/a
pods/log                            default       *          mixed   n/a
applications.build.projectriff.io   default       *          mixed   allowed
containers.build.projectriff.io     default       *          mixed   allowed
functions.build.projectriff.io      default       *          mixed   allowed
`,
		},
		{
			Name:     "error getting namespace",
			Args:     []string{},
			Runtimes: &[]string{},
			GivenObjects: []runtime.Object{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "riff-system"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "applications.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "containers.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "functions.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "deployers.core.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "processors.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "streams.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "inmemorygateways.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "kafkagateways.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "pulsargateways.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "adapters.knative.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "deployers.knative.projectriff.io"}},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("get", "namespaces"),
			},
			ShouldError: true,
		},
		{
			Name:     "error getting crd",
			Args:     []string{},
			Runtimes: &[]string{},
			GivenObjects: []runtime.Object{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "riff-system"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "applications.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "containers.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "functions.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "deployers.core.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "processors.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "streams.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "inmemorygateways.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "kafkagateways.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "pulsargateways.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "adapters.knative.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "deployers.knative.projectriff.io"}},
			},
			ExpectCreates: merge(
				selfSubjectAccessReviewRequests("riff-system", "builders", "core", "configmaps", "", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "configmaps", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "secrets", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "pods", "", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "pods", "log", readVerbs...),
			),
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("get", "customresourcedefinitions"),
				passAccessReview(),
			},
			ShouldError: true,
		},
		{
			Name:     "error creating selfsubjectaccessreview",
			Args:     []string{},
			Runtimes: &[]string{},
			GivenObjects: []runtime.Object{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "riff-system"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "applications.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "containers.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "functions.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "deployers.core.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "processors.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "streams.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "inmemorygateways.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "kafkagateways.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "pulsargateways.streaming.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "adapters.knative.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "deployers.knative.projectriff.io"}},
			},
			ExpectCreates: selfSubjectAccessReviewRequests("riff-system", "builders", "core", "configmaps", "", "get"),
			WithReactors: []rifftesting.ReactionFunc{
				failAccessReview(),
			},
			ShouldError: true,
		},
		{
			Name:     "error evaluating selfsubjectaccessreview",
			Args:     []string{},
			Runtimes: &[]string{},
			GivenObjects: []runtime.Object{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "riff-system"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "applications.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "containers.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "functions.build.projectriff.io"}},
			},
			ExpectCreates: selfSubjectAccessReviewRequests("riff-system", "builders", "core", "configmaps", "", "get"),
			WithReactors: []rifftesting.ReactionFunc{
				failAccessReviewEvaluationOn("*", "*"),
			},
			ShouldError: true,
		},
		{
			Name:     "error evaluating selfsubjectaccessreview",
			Args:     []string{},
			Runtimes: &[]string{},
			GivenObjects: []runtime.Object{
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "riff-system"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "applications.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "containers.build.projectriff.io"}},
				&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: "functions.build.projectriff.io"}},
			},
			ExpectCreates: merge(
				selfSubjectAccessReviewRequests("riff-system", "builders", "core", "configmaps", "", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "configmaps", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "secrets", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "pods", "", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "core", "pods", "log", readVerbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "applications", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "containers", "", verbs...),
				selfSubjectAccessReviewRequests("default", "", "build.projectriff.io", "functions", "", verbs...),
			),
			WithReactors: []rifftesting.ReactionFunc{
				unknownAccessReviewOn("*", "*"),
			},
			ExpectOutput: `
NAMESPACE     STATUS
default       ok
riff-system   ok

RESOURCE                            NAMESPACE     NAME       READ      WRITE
configmaps                          riff-system   builders   unknown   n/a
configmaps                          default       *          unknown   unknown
secrets                             default       *          unknown   unknown
pods                                default       *          unknown   n/a
pods/log                            default       *          unknown   n/a
applications.build.projectriff.io   default       *          unknown   unknown
containers.build.projectriff.io     default       *          unknown   unknown
functions.build.projectriff.io      default       *          unknown   unknown
`,
		},
	}

	table.Run(t, commands.NewDoctorCommand)
}

func merge(objectSets ...[]runtime.Object) []runtime.Object {
	var result []runtime.Object
	for _, objects := range objectSets {
		result = append(result, objects...)
	}
	return result
}

func selfSubjectAccessReviewRequests(namespace, name, group, resource string, subresource string, verbs ...string) []runtime.Object {
	result := make([]runtime.Object, len(verbs))
	for i, verb := range verbs {
		result[i] = &authorizationv1.SelfSubjectAccessReview{
			Spec: authorizationv1.SelfSubjectAccessReviewSpec{
				ResourceAttributes: &authorizationv1.ResourceAttributes{
					Namespace:   namespace,
					Name:        name,
					Group:       group,
					Verb:        verb,
					Resource:    resource,
					Subresource: subresource,
				},
			},
		}
	}
	return result
}

func failAccessReview() func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
	// fork of riffesting.InduceFailure that returns the review to avoid a panic in the fake client
	return func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
		if !action.Matches("create", "selfsubjectaccessreviews") {
			return false, nil, nil
		}
		creationAction, _ := action.(clientgotesting.CreateAction)
		review, _ := creationAction.GetObject().(*authorizationv1.SelfSubjectAccessReview)
		return true, review, fmt.Errorf("inducing failure for %s %s", action.GetVerb(), action.GetResource().Resource)
	}
}

func failAccessReviewEvaluationOn(resource string, verb string) func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
	return func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
		if !action.Matches("create", "selfsubjectaccessreviews") {
			return false, nil, nil
		}
		creationAction, _ := action.(clientgotesting.CreateAction)
		review, _ := creationAction.GetObject().(*authorizationv1.SelfSubjectAccessReview)
		if (resource != "*" && review.Spec.ResourceAttributes.Resource != resource) || (verb != "*" && review.Spec.ResourceAttributes.Verb != verb) {
			return false, nil, nil
		}
		review = review.DeepCopy()
		review.Status.EvaluationError = "induced failure"
		return true, review, nil
	}
}
func unknownAccessReviewOn(resource string, verb string) func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
	return func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
		if !action.Matches("create", "selfsubjectaccessreviews") {
			return false, nil, nil
		}
		creationAction, _ := action.(clientgotesting.CreateAction)
		review, _ := creationAction.GetObject().(*authorizationv1.SelfSubjectAccessReview)
		if (resource != "*" && review.Spec.ResourceAttributes.Resource != resource) || (verb != "*" && review.Spec.ResourceAttributes.Verb != verb) {
			return false, nil, nil
		}
		review = review.DeepCopy()
		return true, review, nil
	}
}

func denyAccessReviewOn(resource string, verb string) func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
	return func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
		if !action.Matches("create", "selfsubjectaccessreviews") {
			return false, nil, nil
		}
		creationAction, _ := action.(clientgotesting.CreateAction)
		review, _ := creationAction.GetObject().(*authorizationv1.SelfSubjectAccessReview)
		if (resource != "*" && review.Spec.ResourceAttributes.Resource != resource) || (verb != "*" && review.Spec.ResourceAttributes.Verb != verb) {
			return false, nil, nil
		}
		review = review.DeepCopy()
		review.Status.Denied = true
		return true, review, nil
	}
}

func passAccessReview() func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
	return func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
		if !action.Matches("create", "selfsubjectaccessreviews") {
			return false, nil, nil
		}
		creationAction, _ := action.(clientgotesting.CreateAction)
		review, _ := creationAction.GetObject().(*authorizationv1.SelfSubjectAccessReview)
		review = review.DeepCopy()
		review.Status.Allowed = true
		return true, review, nil
	}
}
