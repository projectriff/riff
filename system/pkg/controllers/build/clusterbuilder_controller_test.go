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

	kpackbuildv1alpha1 "github.com/projectriff/riff/system/pkg/apis/thirdparty/kpack/build/v1alpha1"
	"github.com/projectriff/riff/system/pkg/controllers/build"
	rtesting "github.com/projectriff/riff/system/pkg/controllers/testing"
	"github.com/projectriff/riff/system/pkg/controllers/testing/factories"
	"github.com/projectriff/riff/system/pkg/tracker"
)

func TestClusterBuildersReconciler(t *testing.T) {
	testNamespace := "riff-system"
	testName := "builders"
	testKey := types.NamespacedName{Namespace: testNamespace, Name: testName}
	testApplicationImage := "projectriff/builder:application"
	testFunctionImage := "projectriff/builder:function"

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = kpackbuildv1alpha1.AddToScheme(scheme)

	testApplicationBuilder := factories.KpackClusterBuilder().
		NamespaceName("", "riff-application").
		Image(testApplicationImage)
	testApplicationBuilderReady := testApplicationBuilder.
		StatusReady().
		StatusLatestImage(testApplicationImage)
	testFunctionBuilder := factories.KpackClusterBuilder().
		NamespaceName("", "riff-function").
		Image(testFunctionImage)
	testFunctionBuilderReady := testFunctionBuilder.
		StatusReady().
		StatusLatestImage(testFunctionImage)

	testBuilders := factories.ConfigMap().
		NamespaceName(testNamespace, testName)

	table := rtesting.Table{{
		Name: "builders configmap does not exist",
		Key:  testKey,
		ExpectCreates: []rtesting.Factory{
			testBuilders,
		},
	}, {
		Name: "builders configmap unchanged",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testBuilders,
		},
	}, {
		Name: "ignore deleted builders configmap",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testBuilders.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.Deleted(1)
				}),
			testApplicationBuilder,
			testFunctionBuilder,
		},
	}, {
		Name: "ignore other configmaps in the correct namespace",
		Key:  types.NamespacedName{Namespace: testNamespace, Name: "not-builders"},
		GivenObjects: []rtesting.Factory{
			testBuilders.
				NamespaceName(testNamespace, "not-builders"),
			testApplicationBuilder,
			testFunctionBuilder,
		},
	}, {
		Name: "ignore other configmaps in the wrong namespace",
		Key:  types.NamespacedName{Namespace: "not-riff-system", Name: testName},
		GivenObjects: []rtesting.Factory{
			testBuilders.
				NamespaceName("not-riff-system", testName),
			testApplicationBuilder,
			testFunctionBuilder,
		},
	}, {
		Name: "create builders configmap, not ready",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testApplicationBuilder,
			testFunctionBuilder,
		},
		ExpectCreates: []rtesting.Factory{
			testBuilders.
				AddData("riff-application", "").
				AddData("riff-function", ""),
		},
	}, {
		Name: "create builders configmap, ready",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testApplicationBuilderReady,
			testFunctionBuilderReady,
		},
		ExpectCreates: []rtesting.Factory{
			testBuilders.
				AddData("riff-application", testApplicationImage).
				AddData("riff-function", testFunctionImage),
		},
	}, {
		Name: "create builders configmap, error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("create", "ConfigMap"),
		},
		GivenObjects: []rtesting.Factory{
			testApplicationBuilder,
			testFunctionBuilder,
		},
		ShouldErr: true,
		ExpectCreates: []rtesting.Factory{
			testBuilders.
				AddData("riff-application", "").
				AddData("riff-function", ""),
		},
	}, {
		Name: "update builders configmap",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testBuilders,
			testApplicationBuilderReady,
			testFunctionBuilderReady,
		},
		ExpectUpdates: []rtesting.Factory{
			testBuilders.
				AddData("riff-application", testApplicationImage).
				AddData("riff-function", testFunctionImage),
		},
	}, {
		Name: "update builders configmap, error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("update", "ConfigMap"),
		},
		GivenObjects: []rtesting.Factory{
			testBuilders,
			testApplicationBuilderReady,
			testFunctionBuilderReady,
		},
		ShouldErr: true,
		ExpectUpdates: []rtesting.Factory{
			testBuilders.
				AddData("riff-application", testApplicationImage).
				AddData("riff-function", testFunctionImage),
		},
	}, {
		Name: "get builders configmap error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("get", "ConfigMap"),
		},
		GivenObjects: []rtesting.Factory{
			testBuilders,
			testApplicationBuilder,
			testFunctionBuilder,
		},
		ShouldErr: true,
	}, {
		Name: "list builders error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("list", "ClusterBuilderList"),
		},
		GivenObjects: []rtesting.Factory{
			testBuilders,
			testApplicationBuilder,
			testFunctionBuilder,
		},
		ShouldErr: true,
	}}

	table.Test(t, scheme, func(t *testing.T, row *rtesting.Testcase, client client.Client, apiReader client.Reader, tracker tracker.Tracker, recorder record.EventRecorder, log logr.Logger) reconcile.Reconciler {
		return &build.ClusterBuilderReconciler{
			Client:    client,
			Recorder:  recorder,
			Log:       log,
			Namespace: testNamespace,
		}
	})
}
