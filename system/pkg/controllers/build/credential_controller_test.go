/*
Copyright 2019 the original author or authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package build_test

import (
	"testing"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	buildv1alpha1 "github.com/projectriff/riff/system/pkg/apis/build/v1alpha1"
	"github.com/projectriff/riff/system/pkg/controllers/build"
	rtesting "github.com/projectriff/riff/system/pkg/controllers/testing"
	"github.com/projectriff/riff/system/pkg/controllers/testing/factories"
	"github.com/projectriff/riff/system/pkg/tracker"
)

func TestCredentialsReconciler(t *testing.T) {
	testNamespace := "test-namespace"
	testName := "riff-build"
	testKey := types.NamespacedName{Namespace: testNamespace, Name: testName}

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = buildv1alpha1.AddToScheme(scheme)

	testServiceAccount := factories.ServiceAccount().
		NamespaceName(testNamespace, testName)
	testCredential := factories.Secret().
		ObjectMeta(func(om factories.ObjectMeta) {
			om.AddLabel(buildv1alpha1.CredentialLabelKey, "docker-hub")
		}).
		NamespaceName(testNamespace, "my-credential")

	testApplication := factories.Application().
		NamespaceName(testNamespace, "my-application")
	testFunction := factories.Function().
		NamespaceName(testNamespace, "my-function")
	testContainer := factories.Container().
		NamespaceName(testNamespace, "my-container")

	table := rtesting.Table{{
		Name: "service account does not exist",
		Key:  testKey,
	}, {
		Name: "ignore deleted service account",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testServiceAccount.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.Deleted(1)
				}),
		},
	}, {
		Name: "ignore non-build service accounts",
		Key:  types.NamespacedName{Namespace: testNamespace, Name: "not-riff-build"},
		GivenObjects: []rtesting.Factory{
			testServiceAccount.
				NamespaceName(testNamespace, "not-riff-build"),
		},
	}, {
		Name: "error fetching service account",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("get", "ServiceAccount"),
		},
		GivenObjects: []rtesting.Factory{
			testServiceAccount,
		},
		ShouldErr: true,
	}, {
		Name: "create service account for application",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testApplication,
		},
		ExpectCreates: []rtesting.Factory{
			testServiceAccount.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddAnnotation("build.projectriff.io/credentials", "")
				}),
		},
	}, {
		Name: "create service account for application, listing fail",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("list", "ApplicationList"),
		},
		GivenObjects: []rtesting.Factory{
			testApplication,
		},
		ShouldErr: true,
	}, {
		Name: "create service account for function",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testFunction,
		},
		ExpectCreates: []rtesting.Factory{
			testServiceAccount.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddAnnotation("build.projectriff.io/credentials", "")
				}),
		},
	}, {
		Name: "create service account for function, listing fail",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("list", "FunctionList"),
		},
		GivenObjects: []rtesting.Factory{
			testFunction,
		},
		ShouldErr: true,
	}, {
		Name: "create service account for container",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testContainer,
		},
		ExpectCreates: []rtesting.Factory{
			testServiceAccount.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddAnnotation("build.projectriff.io/credentials", "")
				}),
		},
	}, {
		Name: "create service account for container, listing fail",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("list", "ContainerList"),
		},
		GivenObjects: []rtesting.Factory{
			testContainer,
		},
		ShouldErr: true,
	}, {
		Name: "create service account for credential",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testCredential,
		},
		ExpectCreates: []rtesting.Factory{
			testServiceAccount.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddAnnotation("build.projectriff.io/credentials", testCredential.Create().GetName())
				}).
				Secrets(testCredential.Create().GetName()),
		},
	}, {
		Name: "create service account for credential, listing fail",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("list", "SecretList"),
		},
		GivenObjects: []rtesting.Factory{
			testCredential,
		},
		ShouldErr: true,
	}, {
		Name: "create service account, error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("create", "ServiceAccount"),
		},
		GivenObjects: []rtesting.Factory{
			testCredential,
		},
		ShouldErr: true,
		ExpectCreates: []rtesting.Factory{
			testServiceAccount.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddAnnotation("build.projectriff.io/credentials", testCredential.Create().GetName())
				}).
				Secrets(testCredential.Create().GetName()),
		},
	}, {
		Name: "add credential to service account",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testServiceAccount,
			testCredential,
		},
		ExpectUpdates: []rtesting.Factory{
			testServiceAccount.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddAnnotation("build.projectriff.io/credentials", testCredential.Create().GetName())
				}).
				Secrets(testCredential.Create().GetName()),
		},
	}, {
		Name: "add credentials to service account",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testServiceAccount,
			testCredential.
				NamespaceName(testNamespace, "cred-1"),
			testCredential.
				NamespaceName(testNamespace, "cred-2"),
		},
		ExpectUpdates: []rtesting.Factory{
			testServiceAccount.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddAnnotation("build.projectriff.io/credentials", "cred-1,cred-2")
				}).
				Secrets("cred-1", "cred-2"),
		},
	}, {
		Name: "add credential to service account, list credentials error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("list", "SecretList"),
		},
		GivenObjects: []rtesting.Factory{
			testServiceAccount,
			testCredential,
		},
		ShouldErr: true,
	}, {
		Name: "add credential to service account, update error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("update", "ServiceAccount"),
		},
		GivenObjects: []rtesting.Factory{
			testServiceAccount,
			testCredential,
		},
		ShouldErr: true,
		ExpectUpdates: []rtesting.Factory{
			testServiceAccount.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddAnnotation("build.projectriff.io/credentials", testCredential.Create().GetName())
				}).
				Secrets(testCredential.Create().GetName()),
		},
	}, {
		Name: "add credentials to service account, preserving non-credential bound secrets",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testServiceAccount.
				Secrets("keep-me"),
			testCredential.
				NamespaceName(testNamespace, "cred-1"),
			testCredential.
				NamespaceName(testNamespace, "cred-2"),
		},
		ExpectUpdates: []rtesting.Factory{
			testServiceAccount.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddAnnotation("build.projectriff.io/credentials", "cred-1,cred-2")
				}).
				Secrets("keep-me", "cred-1", "cred-2"),
		},
	}, {
		Name: "ignore non-credential secrets for service account",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testServiceAccount.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddAnnotation("build.projectriff.io/credentials", "")
				}).
				Secrets(),
			factories.Secret().
				NamespaceName(testNamespace, "not-a-credential"),
		},
	}, {
		Name: "remove credential from service account",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testServiceAccount.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddAnnotation("build.projectriff.io/credentials", "cred-1,cred-2")
				}).
				Secrets("cred-1", "cred-2"),
		},
		ExpectUpdates: []rtesting.Factory{
			testServiceAccount.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddAnnotation("build.projectriff.io/credentials", "")
				}).
				Secrets(),
		},
	}, {
		Name: "remove credential from service account, preserving non-credential bound secrets",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testServiceAccount.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddAnnotation("build.projectriff.io/credentials", "cred-1,cred-2")
				}).
				Secrets("keep-me", "cred-1", "cred-2"),
		},
		ExpectUpdates: []rtesting.Factory{
			testServiceAccount.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddAnnotation("build.projectriff.io/credentials", "")
				}).
				Secrets("keep-me"),
		},
	}}

	table.Test(t, scheme, func(t *testing.T, row *rtesting.Testcase, client client.Client, apiReader client.Reader, tracker tracker.Tracker, recorder record.EventRecorder, log logr.Logger) reconcile.Reconciler {
		return &build.CredentialReconciler{
			Client:   client,
			Recorder: recorder,
			Log:      log,
		}
	})
}
