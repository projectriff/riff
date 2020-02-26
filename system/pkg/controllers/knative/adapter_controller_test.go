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

package knative_test

import (
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	knativev1alpha1 "github.com/projectriff/system/pkg/apis/knative/v1alpha1"
	knativeservingv1 "github.com/projectriff/system/pkg/apis/thirdparty/knative/serving/v1"
	"github.com/projectriff/system/pkg/controllers"
	"github.com/projectriff/system/pkg/controllers/knative"
	rtesting "github.com/projectriff/system/pkg/controllers/testing"
	"github.com/projectriff/system/pkg/controllers/testing/factories"
	"github.com/projectriff/system/pkg/tracker"
)

func TestAdapterReconciler(t *testing.T) {
	testNamespace := "test-namespace"
	testName := "test-adapter"
	testKey := types.NamespacedName{Namespace: testNamespace, Name: testName}
	testImagePrefix := "example.com/repo"
	testSha256 := "cf8b4c69d5460f88530e1c80b8856a70801f31c50b191c8413043ba9b160a43e"
	testImage := fmt.Sprintf("%s/%s@sha256:%s", testImagePrefix, testName, testSha256)

	adapterConditionBuildReady := factories.Condition().Type(knativev1alpha1.AdapterConditionBuildReady)
	adapterConditionReady := factories.Condition().Type(knativev1alpha1.AdapterConditionReady)
	adapterConditionTargetFound := factories.Condition().Type(knativev1alpha1.AdapterConditionTargetFound)

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = buildv1alpha1.AddToScheme(scheme)
	_ = knativev1alpha1.AddToScheme(scheme)
	_ = knativeservingv1.AddToScheme(scheme)

	testAdapter := factories.AdapterKnative().
		NamespaceName(testNamespace, testName)

	testApplication := factories.Application().
		NamespaceName(testNamespace, "my-application")
	testFunction := factories.Function().
		NamespaceName(testNamespace, "my-function")
	testContainer := factories.Container().
		NamespaceName(testNamespace, "my-container")

	testConfiguration := factories.KnativeConfiguration().
		NamespaceName(testNamespace, "my-configuration").
		UserContainer(nil)
	testService := factories.KnativeService().
		NamespaceName(testNamespace, "my-service").
		UserContainer(nil)

	table := rtesting.Table{{
		Name: "adapter does not exist",
		Key:  testKey,
	}, {
		Name: "ignore deleted adapter",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testAdapter.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.Deleted(1)
				}),
		},
	}, {
		Name: "error fetching adapter",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("get", "Adapter"),
		},
		GivenObjects: []rtesting.Factory{
			testAdapter,
		},
		ShouldErr: true,
	}, {
		Name: "error updating adapter status",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("update", "Adapter"),
		},
		GivenObjects: []rtesting.Factory{
			testAdapter.
				ApplicationRef(testApplication.Create().GetName()).
				ServiceRef(testService.Create().GetName()),
			testApplication.
				StatusLatestImage(testImage).
				StatusReady(),
			testService,
		},
		ShouldErr: true,
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testApplication, testAdapter, scheme),
			rtesting.NewTrackRequest(testService, testAdapter, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testAdapter, scheme, corev1.EventTypeWarning, "StatusUpdateFailed",
				`Failed to update status: inducing failure for update Adapter`),
		},
		ExpectUpdates: []rtesting.Factory{
			testService.
				UserContainer(func(uc *corev1.Container) {
					uc.Image = testImage
				}),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testAdapter.
				StatusConditions(
					adapterConditionBuildReady.True(),
					adapterConditionReady.True(),
					adapterConditionTargetFound.True(),
				).
				StatusLatestImage(testImage),
		},
	}, {
		Name: "adapt application to service",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testAdapter.
				ApplicationRef(testApplication.Create().GetName()).
				ServiceRef(testService.Create().GetName()),
			testApplication.
				StatusLatestImage(testImage).
				StatusReady(),
			testService,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testApplication, testAdapter, scheme),
			rtesting.NewTrackRequest(testService, testAdapter, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testAdapter, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectUpdates: []rtesting.Factory{
			testService.
				UserContainer(func(uc *corev1.Container) {
					uc.Image = testImage
				}),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testAdapter.
				StatusConditions(
					adapterConditionBuildReady.True(),
					adapterConditionReady.True(),
					adapterConditionTargetFound.True(),
				).
				StatusLatestImage(testImage),
		},
	}, {
		Name: "adapt application to service, application not ready",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testAdapter.
				ApplicationRef(testApplication.Create().GetName()).
				ServiceRef(testService.Create().GetName()),
			testApplication,
			testService,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testApplication, testAdapter, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testAdapter, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testAdapter.
				StatusConditions(
					adapterConditionBuildReady.Unknown(),
					adapterConditionReady.Unknown(),
					adapterConditionTargetFound.Unknown(),
				),
		},
	}, {
		Name: "adapt application to service, application not found",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testAdapter.
				ApplicationRef(testApplication.Create().GetName()).
				ServiceRef(testService.Create().GetName()),
			testService,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testApplication, testAdapter, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testAdapter, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testAdapter.
				StatusConditions(
					adapterConditionBuildReady.False().Reason("NotFound", `The application "my-application" was not found.`),
					adapterConditionReady.False().Reason("NotFound", `The application "my-application" was not found.`),
					adapterConditionTargetFound.Unknown(),
				),
		},
	}, {
		Name: "adapt application to service, get application failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("get", "Application"),
		},
		GivenObjects: []rtesting.Factory{
			testAdapter.
				ApplicationRef(testApplication.Create().GetName()).
				ServiceRef(testService.Create().GetName()),
			testApplication.
				StatusLatestImage(testImage).
				StatusReady(),
			testService,
		},
		ShouldErr: true,
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testApplication, testAdapter, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testAdapter, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testAdapter.
				StatusConditions(
					adapterConditionBuildReady.Unknown(),
					adapterConditionReady.Unknown(),
					adapterConditionTargetFound.Unknown(),
				),
		},
	}, {
		Name: "adapt function to service",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testAdapter.
				FunctionRef(testFunction.Create().GetName()).
				ServiceRef(testService.Create().GetName()),
			testFunction.
				StatusLatestImage(testImage).
				StatusReady(),
			testService,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testFunction, testAdapter, scheme),
			rtesting.NewTrackRequest(testService, testAdapter, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testAdapter, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectUpdates: []rtesting.Factory{
			testService.
				UserContainer(func(uc *corev1.Container) {
					uc.Image = testImage
				}),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testAdapter.
				StatusConditions(
					adapterConditionBuildReady.True(),
					adapterConditionReady.True(),
					adapterConditionTargetFound.True(),
				).
				StatusLatestImage(testImage),
		},
	}, {
		Name: "adapt function to service, function not ready",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testAdapter.
				FunctionRef(testFunction.Create().GetName()).
				ServiceRef(testService.Create().GetName()),
			testFunction,
			testService,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testFunction, testAdapter, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testAdapter, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testAdapter.
				StatusConditions(
					adapterConditionBuildReady.Unknown(),
					adapterConditionReady.Unknown(),
					adapterConditionTargetFound.Unknown(),
				),
		},
	}, {
		Name: "adapt function to service, function not found",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testAdapter.
				FunctionRef(testFunction.Create().GetName()).
				ServiceRef(testService.Create().GetName()),
			testService,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testFunction, testAdapter, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testAdapter, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testAdapter.
				StatusConditions(
					adapterConditionBuildReady.False().Reason("NotFound", `The function "my-function" was not found.`),
					adapterConditionReady.False().Reason("NotFound", `The function "my-function" was not found.`),
					adapterConditionTargetFound.Unknown(),
				),
		},
	}, {
		Name: "adapt function to service, get function failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("get", "function"),
		},
		GivenObjects: []rtesting.Factory{
			testAdapter.
				FunctionRef(testFunction.Create().GetName()).
				ServiceRef(testService.Create().GetName()),
			testFunction.
				StatusLatestImage(testImage).
				StatusReady(),
			testService,
		},
		ShouldErr: true,
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testFunction, testAdapter, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testAdapter, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testAdapter.
				StatusConditions(
					adapterConditionBuildReady.Unknown(),
					adapterConditionReady.Unknown(),
					adapterConditionTargetFound.Unknown(),
				),
		},
	}, {
		Name: "adapt container to service",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testAdapter.
				ContainerRef(testContainer.Create().GetName()).
				ServiceRef(testService.Create().GetName()),
			testContainer.
				StatusLatestImage(testImage).
				StatusReady(),
			testService,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testContainer, testAdapter, scheme),
			rtesting.NewTrackRequest(testService, testAdapter, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testAdapter, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectUpdates: []rtesting.Factory{
			testService.
				UserContainer(func(uc *corev1.Container) {
					uc.Image = testImage
				}),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testAdapter.
				StatusConditions(
					adapterConditionBuildReady.True(),
					adapterConditionReady.True(),
					adapterConditionTargetFound.True(),
				).
				StatusLatestImage(testImage),
		},
	}, {
		Name: "adapt container to service, container not ready",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testAdapter.
				ContainerRef(testContainer.Create().GetName()).
				ServiceRef(testService.Create().GetName()),
			testContainer,
			testService,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testContainer, testAdapter, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testAdapter, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testAdapter.
				StatusConditions(
					adapterConditionBuildReady.Unknown(),
					adapterConditionReady.Unknown(),
					adapterConditionTargetFound.Unknown(),
				),
		},
	}, {
		Name: "adapt container to service, container not found",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testAdapter.
				ContainerRef(testContainer.Create().GetName()).
				ServiceRef(testService.Create().GetName()),
			testService,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testContainer, testAdapter, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testAdapter, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testAdapter.
				StatusConditions(
					adapterConditionBuildReady.False().Reason("NotFound", `The container "my-container" was not found.`),
					adapterConditionReady.False().Reason("NotFound", `The container "my-container" was not found.`),
					adapterConditionTargetFound.Unknown(),
				),
		},
	}, {
		Name: "adapt container to service, get container failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("get", "Container"),
		},
		GivenObjects: []rtesting.Factory{
			testAdapter.
				ContainerRef(testContainer.Create().GetName()).
				ServiceRef(testService.Create().GetName()),
			testContainer.
				StatusLatestImage(testImage).
				StatusReady(),
			testService,
		},
		ShouldErr: true,
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testContainer, testAdapter, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testAdapter, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testAdapter.
				StatusConditions(
					adapterConditionBuildReady.Unknown(),
					adapterConditionReady.Unknown(),
					adapterConditionTargetFound.Unknown(),
				),
		},
	}, {
		Name: "adapt container to service, service not found",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testAdapter.
				ContainerRef(testContainer.Create().GetName()).
				ServiceRef(testService.Create().GetName()),
			testContainer.
				StatusLatestImage(testImage).
				StatusReady(),
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testContainer, testAdapter, scheme),
			rtesting.NewTrackRequest(testService, testAdapter, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testAdapter, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testAdapter.
				StatusConditions(
					adapterConditionBuildReady.True(),
					adapterConditionReady.False().Reason("NotFound", `The service "my-service" was not found.`),
					adapterConditionTargetFound.False().Reason("NotFound", `The service "my-service" was not found.`),
				).
				StatusLatestImage(testImage),
		},
	}, {
		Name: "adapt container to service, get service failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("get", "Service"),
		},
		GivenObjects: []rtesting.Factory{
			testAdapter.
				ContainerRef(testContainer.Create().GetName()).
				ServiceRef(testService.Create().GetName()),
			testContainer.
				StatusLatestImage(testImage).
				StatusReady(),
			testService,
		},
		ShouldErr: true,
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testContainer, testAdapter, scheme),
			rtesting.NewTrackRequest(testService, testAdapter, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testAdapter, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testAdapter.
				StatusConditions(
					adapterConditionBuildReady.True(),
					adapterConditionReady.Unknown(),
					adapterConditionTargetFound.Unknown(),
				).
				StatusLatestImage(testImage),
		},
	}, {
		Name: "adapt container to service, service is up to date",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testAdapter.
				ContainerRef(testContainer.Create().GetName()).
				ServiceRef(testService.Create().GetName()).
				StatusConditions(
					adapterConditionBuildReady.True(),
					adapterConditionReady.True(),
					adapterConditionTargetFound.True(),
				).
				StatusLatestImage(testImage),
			testContainer.
				StatusLatestImage(testImage).
				StatusReady(),
			testService.
				UserContainer(func(uc *corev1.Container) {
					uc.Image = testImage
				}),
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testContainer, testAdapter, scheme),
			rtesting.NewTrackRequest(testService, testAdapter, scheme),
		},
	}, {
		Name: "adapt container to service, update service failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("update", "Service"),
		},
		GivenObjects: []rtesting.Factory{
			testAdapter.
				ContainerRef(testContainer.Create().GetName()).
				ServiceRef(testService.Create().GetName()),
			testContainer.
				StatusLatestImage(testImage).
				StatusReady(),
			testService,
		},
		ShouldErr: true,
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testContainer, testAdapter, scheme),
			rtesting.NewTrackRequest(testService, testAdapter, scheme),
		},
		ExpectUpdates: []rtesting.Factory{
			testService.
				UserContainer(func(uc *corev1.Container) {
					uc.Image = testImage
				}),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testAdapter, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testAdapter.
				StatusConditions(
					adapterConditionBuildReady.True(),
					// TODO the update failed, we should not be reporting as ready
					adapterConditionReady.True(),
					adapterConditionTargetFound.True(),
				).
				StatusLatestImage(testImage),
		},
	}, {
		Name: "adapt container to configuration",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testAdapter.
				ContainerRef(testContainer.Create().GetName()).
				ConfigurationRef(testConfiguration.Create().GetName()),
			testContainer.
				StatusLatestImage(testImage).
				StatusReady(),
			testConfiguration,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testContainer, testAdapter, scheme),
			rtesting.NewTrackRequest(testConfiguration, testAdapter, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testAdapter, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectUpdates: []rtesting.Factory{
			testConfiguration.
				UserContainer(func(uc *corev1.Container) {
					uc.Image = testImage
				}),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testAdapter.
				StatusConditions(
					adapterConditionBuildReady.True(),
					adapterConditionReady.True(),
					adapterConditionTargetFound.True(),
				).
				StatusLatestImage(testImage),
		},
	}, {
		Name: "adapt container to configuration, configuration not found",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testAdapter.
				ContainerRef(testContainer.Create().GetName()).
				ConfigurationRef(testConfiguration.Create().GetName()),
			testContainer.
				StatusLatestImage(testImage).
				StatusReady(),
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testContainer, testAdapter, scheme),
			rtesting.NewTrackRequest(testConfiguration, testAdapter, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testAdapter, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testAdapter.
				StatusConditions(
					adapterConditionBuildReady.True(),
					adapterConditionReady.False().Reason("NotFound", `The configuration "my-configuration" was not found.`),
					adapterConditionTargetFound.False().Reason("NotFound", `The configuration "my-configuration" was not found.`),
				).
				StatusLatestImage(testImage),
		},
	}, {
		Name: "adapt container to configuration, get configuration failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("get", "Configuration"),
		},
		GivenObjects: []rtesting.Factory{
			testAdapter.
				ContainerRef(testContainer.Create().GetName()).
				ConfigurationRef(testConfiguration.Create().GetName()),
			testContainer.
				StatusLatestImage(testImage).
				StatusReady(),
			testConfiguration,
		},
		ShouldErr: true,
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testContainer, testAdapter, scheme),
			rtesting.NewTrackRequest(testConfiguration, testAdapter, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testAdapter, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testAdapter.
				StatusConditions(
					adapterConditionBuildReady.True(),
					adapterConditionReady.Unknown(),
					adapterConditionTargetFound.Unknown(),
				).
				StatusLatestImage(testImage),
		},
	}, {
		Name: "adapt container to configuration, configuration is up to date",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testAdapter.
				ContainerRef(testContainer.Create().GetName()).
				ConfigurationRef(testConfiguration.Create().GetName()).
				StatusConditions(
					adapterConditionBuildReady.True(),
					adapterConditionReady.True(),
					adapterConditionTargetFound.True(),
				).
				StatusLatestImage(testImage),
			testContainer.
				StatusLatestImage(testImage).
				StatusReady(),
			testConfiguration.
				UserContainer(func(uc *corev1.Container) {
					uc.Image = testImage
				}),
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testContainer, testAdapter, scheme),
			rtesting.NewTrackRequest(testConfiguration, testAdapter, scheme),
		},
	}, {
		Name: "adapt container to configuration, update configuration failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("update", "Configuration"),
		},
		GivenObjects: []rtesting.Factory{
			testAdapter.
				ContainerRef(testContainer.Create().GetName()).
				ConfigurationRef(testConfiguration.Create().GetName()),
			testContainer.
				StatusLatestImage(testImage).
				StatusReady(),
			testConfiguration,
		},
		ShouldErr: true,
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testContainer, testAdapter, scheme),
			rtesting.NewTrackRequest(testConfiguration, testAdapter, scheme),
		},
		ExpectUpdates: []rtesting.Factory{
			testConfiguration.
				UserContainer(func(uc *corev1.Container) {
					uc.Image = testImage
				}),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testAdapter, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testAdapter.
				StatusConditions(
					adapterConditionBuildReady.True(),
					// TODO the update failed, we should not be reporting as ready
					adapterConditionReady.True(),
					adapterConditionTargetFound.True(),
				).
				StatusLatestImage(testImage),
		},
	}}

	table.Test(t, scheme, func(t *testing.T, row *rtesting.Testcase, client client.Client, apiReader client.Reader, tracker tracker.Tracker, recorder record.EventRecorder, log logr.Logger) reconcile.Reconciler {
		return knative.AdapterReconciler(
			controllers.Config{
				Client:    client,
				APIReader: apiReader,
				Recorder:  recorder,
				Log:       log,
				Scheme:    scheme,
				Tracker:   tracker,
			},
		)
	})
}
