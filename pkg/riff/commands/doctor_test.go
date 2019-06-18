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

	"github.com/projectriff/riff/pkg/riff/commands"
	rifftesting "github.com/projectriff/riff/pkg/testing"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestDoctorOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name:           "valid list",
			Options:        &commands.DoctorOptions{},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestDoctorCommand(t *testing.T) {
	table := rifftesting.CommandTable{
		{
			Name: "installation is ok",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "istio-system"},
				},
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "knative-build"},
				},
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "knative-serving"},
				},
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "riff-system"},
				},
			},
			ExpectOutput: `
Namespace "istio-system"      OK
Namespace "knative-build"     OK
Namespace "knative-serving"   OK
Namespace "riff-system"       OK

Installation is OK
`,
		},
		{
			Name: "istio-system is missing",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "knative-build"},
				},
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "knative-serving"},
				},
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "riff-system"},
				},
			},
			ExpectOutput: `
Namespace "istio-system"      Missing
Namespace "knative-build"     OK
Namespace "knative-serving"   OK
Namespace "riff-system"       OK

Installation is not healthy
`,
		},
		{
			Name: "multiple namespaces are missing",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "knative-build"},
				},
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "knative-serving"},
				},
			},
			ExpectOutput: `
Namespace "istio-system"      Missing
Namespace "knative-build"     OK
Namespace "knative-serving"   OK
Namespace "riff-system"       Missing

Installation is not healthy
`,
		},
		{
			Name: "error",
			Args: []string{},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("list", "namespaces"),
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewDoctorCommand)
}
