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
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	buildv1alpha1 "github.com/projectriff/riff/system/pkg/apis/build/v1alpha1"
	knativev1alpha1 "github.com/projectriff/riff/system/pkg/apis/knative/v1alpha1"
	knativeservingv1 "github.com/projectriff/riff/system/pkg/apis/thirdparty/knative/serving/v1"
	"github.com/projectriff/riff/system/pkg/controllers"
	"github.com/projectriff/riff/system/pkg/controllers/knative"
	rtesting "github.com/projectriff/riff/system/pkg/controllers/testing"
	"github.com/projectriff/riff/system/pkg/controllers/testing/factories"
	"github.com/projectriff/riff/system/pkg/tracker"
)

func TestDeployerReconciler(t *testing.T) {
	testNamespace := "test-namespace"
	testName := "test-deployer"
	testKey := types.NamespacedName{Namespace: testNamespace, Name: testName}
	testImagePrefix := "example.com/repo"
	testSha256 := "cf8b4c69d5460f88530e1c80b8856a70801f31c50b191c8413043ba9b160a43e"
	testImage := fmt.Sprintf("%s/%s@sha256:%s", testImagePrefix, testName, testSha256)
	testAddressURL := "http://internal.local"
	testURL := "http://example.com"

	deployerConditionConfigurationReady := factories.Condition().Type(knativev1alpha1.DeployerConditionConfigurationReady)
	deployerConditionReady := factories.Condition().Type(knativev1alpha1.DeployerConditionReady)
	deployerConditionRouteReady := factories.Condition().Type(knativev1alpha1.DeployerConditionRouteReady)

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = buildv1alpha1.AddToScheme(scheme)
	_ = knativev1alpha1.AddToScheme(scheme)
	_ = knativeservingv1.AddToScheme(scheme)

	testDeployer := factories.DeployerKnative().
		NamespaceName(testNamespace, testName)

	testApplication := factories.Application().
		NamespaceName(testNamespace, "my-application").
		StatusLatestImage(testImage)
	testFunction := factories.Function().
		NamespaceName(testNamespace, "my-function").
		StatusLatestImage(testImage)
	testContainer := factories.Container().
		NamespaceName(testNamespace, "my-container").
		StatusLatestImage(testImage)

	testConfigurationCreate := factories.KnativeConfiguration().
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Namespace(testNamespace)
			om.GenerateName("%s-deployer-", testName)
			om.ControlledBy(testDeployer, scheme)
			om.AddLabel(knativev1alpha1.DeployerLabelKey, testName)
		}).
		PodTemplateSpec(func(pts factories.PodTemplateSpec) {
			pts.AddLabel(knativev1alpha1.DeployerLabelKey, testName)
		}).
		UserContainer(func(container *corev1.Container) {
			container.Image = testImage
		})
	testConfigurationGiven := testConfigurationCreate.
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Name("%s001", om.Create().GenerateName)
			om.Created(1)
			om.Generation(1)
		}).
		StatusObservedGeneration(1)

	testRouteCreate := factories.KnativeRoute().
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Namespace(testNamespace)
			om.Name(testName)
			om.ControlledBy(testDeployer, scheme)
			om.AddLabel(knativev1alpha1.DeployerLabelKey, testName)
			om.AddLabel("serving.knative.dev/visibility", "cluster-local")
		}).
		Traffic(
			knativeservingv1.TrafficTarget{
				ConfigurationName: fmt.Sprintf("%s-deployer-%s", testName, "001"),
				Percent:           rtesting.Int64Ptr(100),
			},
		)
	testRouteGiven := testRouteCreate.
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Created(1)
			om.Generation(1)
		}).
		StatusObservedGeneration(1)

	table := rtesting.Table{{
		Name: "deployer does not exist",
		Key:  testKey,
	}, {
		Name: "ignore deleted deployer",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testDeployer.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.Deleted(1)
				}),
		},
	}, {
		Name: "get deployer failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("get", "Deployer"),
		},
		GivenObjects: []rtesting.Factory{
			testDeployer.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.Deleted(1)
				}),
		},
		ShouldErr: true,
	}, {
		Name: "create knative resources, from application",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testDeployer.
				ApplicationRef(testApplication.Create().GetName()),
			testApplication,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testApplication, testDeployer, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Created",
				`Created Configuration "%s-deployer-001"`, testName),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Created",
				`Created Route "%s"`, testName),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			testConfigurationCreate,
			testRouteCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				).
				StatusLatestImage(testImage).
				StatusConfigurationRef(testConfigurationGiven.Create().GetName()).
				StatusRouteRef(testRouteGiven.Create().GetName()),
		},
	}, {
		Name: "create knative resources, from application, application not found",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testDeployer.
				ApplicationRef(testApplication.Create().GetName()),
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testApplication, testDeployer, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				),
		},
	}, {
		Name: "create knative resources, from application, get application failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("get", "Application"),
		},
		GivenObjects: []rtesting.Factory{
			testDeployer.
				ApplicationRef(testApplication.Create().GetName()),
		},
		ShouldErr: true,
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testApplication, testDeployer, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				),
		},
	}, {
		Name: "create knative resources, from application, no latest",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testDeployer.
				ApplicationRef(testApplication.Create().GetName()),
			testApplication.
				StatusLatestImage(""),
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testApplication, testDeployer, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				),
		},
	}, {
		Name: "create knative resources, from function",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testDeployer.
				FunctionRef(testFunction.Create().GetName()),
			testFunction,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testFunction, testDeployer, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Created",
				`Created Configuration "%s-deployer-001"`, testName),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Created",
				`Created Route "%s"`, testName),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			testConfigurationCreate,
			testRouteCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				).
				StatusLatestImage(testImage).
				StatusConfigurationRef(testConfigurationGiven.Create().GetName()).
				StatusRouteRef(testRouteGiven.Create().GetName()),
		},
	}, {
		Name: "create knative resources, from function, function not found",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testDeployer.
				FunctionRef(testFunction.Create().GetName()),
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testFunction, testDeployer, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				),
		},
	}, {
		Name: "create knative resources, from function, get function failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("get", "Function"),
		},
		GivenObjects: []rtesting.Factory{
			testDeployer.
				FunctionRef(testFunction.Create().GetName()),
		},
		ShouldErr: true,
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testFunction, testDeployer, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				),
		},
	}, {
		Name: "create knative resources, from function, no latest",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testDeployer.
				FunctionRef(testFunction.Create().GetName()),
			testFunction.
				StatusLatestImage(""),
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testFunction, testDeployer, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				),
		},
	}, {
		Name: "create knative resources, from container",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testDeployer.
				ContainerRef(testContainer.Create().GetName()),
			testContainer,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testContainer, testDeployer, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Created",
				`Created Configuration "%s-deployer-001"`, testName),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Created",
				`Created Route "%s"`, testName),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			testConfigurationCreate,
			testRouteCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				).
				StatusLatestImage(testImage).
				StatusConfigurationRef(testConfigurationGiven.Create().GetName()).
				StatusRouteRef(testRouteGiven.Create().GetName()),
		},
	}, {
		Name: "create knative resources, from container, container not found",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testDeployer.
				ContainerRef(testContainer.Create().GetName()),
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testContainer, testDeployer, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				),
		},
	}, {
		Name: "create knative resources, from container, get container failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("get", "Container"),
		},
		GivenObjects: []rtesting.Factory{
			testDeployer.
				ContainerRef(testContainer.Create().GetName()),
		},
		ShouldErr: true,
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testContainer, testDeployer, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				),
		},
	}, {
		Name: "create knative resources, from container, no latest",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testDeployer.
				ContainerRef(testContainer.Create().GetName()),
			testContainer.
				StatusLatestImage(""),
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testContainer, testDeployer, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				),
		},
	}, {
		Name: "create knative resources, from image",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testDeployer.
				Image(testImage),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Created",
				`Created Configuration "%s-deployer-001"`, testName),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Created",
				`Created Route "%s"`, testName),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			testConfigurationCreate,
			testRouteCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				).
				StatusLatestImage(testImage).
				StatusConfigurationRef(testConfigurationGiven.Create().GetName()).
				StatusRouteRef(testRouteGiven.Create().GetName()),
		},
	}, {
		Name: "create knative resources, create configuration failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("create", "Configuration"),
		},
		GivenObjects: []rtesting.Factory{
			testDeployer.
				Image(testImage),
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeWarning, "CreationFailed",
				`Failed to create Configuration "": inducing failure for create Configuration`),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			testConfigurationCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				).
				StatusLatestImage(testImage),
		},
	}, {
		Name: "create knative resources, create route failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("create", "Route"),
		},
		GivenObjects: []rtesting.Factory{
			testDeployer.
				Image(testImage),
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Created",
				`Created Configuration "%s-deployer-001"`, testName),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeWarning, "CreationFailed",
				`Failed to create Route "%s": inducing failure for create Route`, testName),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			testConfigurationCreate,
			testRouteCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				).
				StatusLatestImage(testImage).
				StatusConfigurationRef(testConfigurationGiven.Create().GetName()),
		},
	}, {
		Name: "create knative resources, route exists",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("create", "Route", rtesting.InduceFailureOpts{
				Error: apierrs.NewAlreadyExists(schema.GroupResource{}, testName),
			}),
		},
		GivenObjects: []rtesting.Factory{
			testDeployer.
				Image(testImage),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Created",
				`Created Configuration "%s-deployer-001"`, testName),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeWarning, "CreationFailed",
				`Failed to create Route "%s":  "%s" already exists`, testName, testName),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			testConfigurationCreate,
			testRouteCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.False().Reason("NotOwned", `There is an existing Route "test-deployer" that the Deployer does not own.`),
					deployerConditionRouteReady.False().Reason("NotOwned", `There is an existing Route "test-deployer" that the Deployer does not own.`),
				).
				StatusLatestImage(testImage).
				StatusConfigurationRef(testConfigurationGiven.Create().GetName()),
		},
	}, {
		Name: "create knative resources, delete extra configurations",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testDeployer.
				Image(testImage),
			testConfigurationGiven.
				NamespaceName(testNamespace, "extra-configuration-1"),
			testConfigurationGiven.
				NamespaceName(testNamespace, "extra-configuration-2"),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Deleted",
				`Deleted Configuration "%s"`, "extra-configuration-1"),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Deleted",
				`Deleted Configuration "%s"`, "extra-configuration-2"),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Created",
				`Created Configuration "%s-deployer-001"`, testName),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Created",
				`Created Route "%s"`, testName),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			testConfigurationCreate,
			testRouteCreate,
		},
		ExpectDeletes: []rtesting.DeleteRef{
			{Group: "serving.knative.dev", Kind: "Configuration", Namespace: testNamespace, Name: "extra-configuration-1"},
			{Group: "serving.knative.dev", Kind: "Configuration", Namespace: testNamespace, Name: "extra-configuration-2"},
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				).
				StatusLatestImage(testImage).
				StatusConfigurationRef(testConfigurationGiven.Create().GetName()).
				StatusRouteRef(testRouteGiven.Create().GetName()),
		},
	}, {
		Name: "create knative resources, delete extra configurations, delete failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("delete", "Configuration"),
		},
		GivenObjects: []rtesting.Factory{
			testDeployer.
				Image(testImage),
			testConfigurationGiven.
				NamespaceName(testNamespace, "extra-configuration-1"),
			testConfigurationGiven.
				NamespaceName(testNamespace, "extra-configuration-2"),
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeWarning, "DeleteFailed",
				`Failed to delete Configuration "%s": inducing failure for delete Configuration`, "extra-configuration-1"),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectDeletes: []rtesting.DeleteRef{
			{Group: "serving.knative.dev", Kind: "Configuration", Namespace: testNamespace, Name: "extra-configuration-1"},
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				).
				StatusLatestImage(testImage),
		},
	}, {
		Name: "create knative resources, delete extra routes",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testDeployer.
				Image(testImage),
			testRouteGiven.
				NamespaceName(testNamespace, "extra-route-1"),
			testRouteGiven.
				NamespaceName(testNamespace, "extra-route-2"),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Created",
				`Created Configuration "%s-deployer-001"`, testName),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Deleted",
				`Deleted Route "%s"`, "extra-route-1"),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Deleted",
				`Deleted Route "%s"`, "extra-route-2"),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Created",
				`Created Route "%s"`, testName),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			testConfigurationCreate,
			testRouteCreate,
		},
		ExpectDeletes: []rtesting.DeleteRef{
			{Group: "serving.knative.dev", Kind: "Route", Namespace: testNamespace, Name: "extra-route-1"},
			{Group: "serving.knative.dev", Kind: "Route", Namespace: testNamespace, Name: "extra-route-2"},
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				).
				StatusLatestImage(testImage).
				StatusConfigurationRef(testConfigurationGiven.Create().GetName()).
				StatusRouteRef(testRouteGiven.Create().GetName()),
		},
	}, {
		Name: "create knative resources, delete extra routes, delete failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("delete", "Route"),
		},
		GivenObjects: []rtesting.Factory{
			testDeployer.
				Image(testImage),
			testRouteGiven.
				NamespaceName(testNamespace, "extra-route-1"),
			testRouteGiven.
				NamespaceName(testNamespace, "extra-route-2"),
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Created",
				`Created Configuration "%s-deployer-001"`, testName),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeWarning, "DeleteFailed",
				`Failed to delete Route "%s": inducing failure for delete Route`, "extra-route-1"),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			testConfigurationCreate,
		},
		ExpectDeletes: []rtesting.DeleteRef{
			{Group: "serving.knative.dev", Kind: "Route", Namespace: testNamespace, Name: "extra-route-1"},
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				).
				StatusLatestImage(testImage).
				StatusConfigurationRef(testConfigurationGiven.Create().GetName()),
		},
	}, {
		Name: "update configuration",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testDeployer.
				Image(testImage),
			testConfigurationGiven.
				UserContainer(func(container *corev1.Container) {
					container.Image = "bogus"
				}),
			testRouteGiven,
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Updated",
				`Updated Configuration "%s"`, testConfigurationGiven.Create().GetName()),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectUpdates: []rtesting.Factory{
			testConfigurationGiven,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				).
				StatusLatestImage(testImage).
				StatusConfigurationRef(testConfigurationGiven.Create().GetName()).
				StatusRouteRef(testRouteGiven.Create().GetName()),
		},
	}, {
		Name: "update configuration, listing failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("list", "ConfigurationList"),
		},
		GivenObjects: []rtesting.Factory{
			testDeployer.
				Image(testImage),
			testConfigurationGiven.
				UserContainer(func(container *corev1.Container) {
					container.Image = "bogus"
				}),
			testRouteGiven,
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				).
				StatusLatestImage(testImage),
		},
	}, {
		Name: "update configuration, update failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("update", "Configuration"),
		},
		GivenObjects: []rtesting.Factory{
			testDeployer.
				Image(testImage),
			testConfigurationGiven.
				UserContainer(func(container *corev1.Container) {
					container.Image = "bogus"
				}),
			testRouteGiven,
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeWarning, "UpdateFailed",
				`Failed to update Configuration "%s": inducing failure for update Configuration`, testConfigurationGiven.Create().GetName()),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectUpdates: []rtesting.Factory{
			testConfigurationGiven,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				).
				StatusLatestImage(testImage),
		},
	}, {
		Name: "update route",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testDeployer.
				Image(testImage),
			testConfigurationGiven,
			testRouteGiven.
				Traffic(),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Updated",
				`Updated Route "%s"`, testRouteGiven.Create().GetName()),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectUpdates: []rtesting.Factory{
			testRouteGiven,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				).
				StatusLatestImage(testImage).
				StatusConfigurationRef(testConfigurationGiven.Create().GetName()).
				StatusRouteRef(testRouteGiven.Create().GetName()),
		},
	}, {
		Name: "update route, listing failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("list", "RouteList"),
		},
		GivenObjects: []rtesting.Factory{
			testDeployer.
				Image(testImage),
			testConfigurationGiven,
			testRouteGiven.
				Traffic(),
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				).
				StatusLatestImage(testImage).
				StatusConfigurationRef(testConfigurationGiven.Create().GetName()),
		},
	}, {
		Name: "update route, update failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("update", "Route"),
		},
		GivenObjects: []rtesting.Factory{
			testDeployer.
				Image(testImage),
			testConfigurationGiven,
			testRouteGiven.
				Traffic(),
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeWarning, "UpdateFailed",
				`Failed to update Route "%s": inducing failure for update Route`, testRouteGiven.Create().GetName()),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectUpdates: []rtesting.Factory{
			testRouteGiven,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				).
				StatusLatestImage(testImage).
				StatusConfigurationRef(testConfigurationGiven.Create().GetName()),
		},
	}, {
		Name: "update status failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("update", "Deployer"),
		},
		GivenObjects: []rtesting.Factory{
			testDeployer.
				Image(testImage),
			testConfigurationGiven,
			testRouteGiven,
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeWarning, "StatusUpdateFailed",
				`Failed to update status: inducing failure for update Deployer`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				).
				StatusLatestImage(testImage).
				StatusConfigurationRef(testConfigurationGiven.Create().GetName()).
				StatusRouteRef(testRouteGiven.Create().GetName()),
		},
	}, {
		Name: "update knative resources, copy annotations and labels",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testDeployer.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddLabel("test-label", "test-label-value")
				}).
				PodTemplateSpec(func(pts factories.PodTemplateSpec) {
					pts.AddLabel("test-label-pts", "test-label-value")
				}).
				Image(testImage),
			testConfigurationGiven,
			testRouteGiven,
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Updated",
				`Updated Configuration "%s"`, testConfigurationGiven.Create().GetName()),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Updated",
				`Updated Route "%s"`, testRouteGiven.Create().GetName()),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectUpdates: []rtesting.Factory{
			testConfigurationGiven.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddLabel("test-label", "test-label-value")
				}).
				PodTemplateSpec(func(pts factories.PodTemplateSpec) {
					pts.AddLabel("test-label", "test-label-value")
				}),
			testRouteGiven.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddLabel("test-label", "test-label-value")
				}),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				).
				StatusLatestImage(testImage).
				StatusConfigurationRef(testConfigurationGiven.Create().GetName()).
				StatusRouteRef(testRouteGiven.Create().GetName()),
		},
	}, {
		Name: "update knative resources, with container concurrency",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testDeployer.
				Image(testImage).
				ContainerConcurrency(1),
			testConfigurationGiven,
			testRouteGiven,
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Updated",
				`Updated Configuration "%s"`, testConfigurationGiven.Create().GetName()),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectUpdates: []rtesting.Factory{
			testConfigurationGiven.
				ContainerConcurrency(1),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				).
				StatusLatestImage(testImage).
				StatusConfigurationRef(testConfigurationGiven.Create().GetName()).
				StatusRouteRef(testRouteGiven.Create().GetName()),
		},
	}, {
		Name: "update knative resources, with scale",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testDeployer.
				Image(testImage).
				MinScale(1).
				MaxScale(2),
			testConfigurationGiven,
			testRouteGiven,
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "Updated",
				`Updated Configuration "%s"`, testConfigurationGiven.Create().GetName()),
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectUpdates: []rtesting.Factory{
			testConfigurationGiven.
				// TODO figure out which annotation is actually impactful
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddAnnotation("autoscaling.knative.dev/minScale", "1")
					om.AddAnnotation("autoscaling.knative.dev/maxScale", "2")
				}).
				PodTemplateSpec(func(pts factories.PodTemplateSpec) {
					pts.AddAnnotation("autoscaling.knative.dev/minScale", "1")
					pts.AddAnnotation("autoscaling.knative.dev/maxScale", "2")
				}),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionRouteReady.Unknown(),
				).
				StatusLatestImage(testImage).
				StatusConfigurationRef(testConfigurationGiven.Create().GetName()).
				StatusRouteRef(testRouteGiven.Create().GetName()),
		},
	}, {
		Name: "ready",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testDeployer.
				Image(testImage),
			testConfigurationGiven.
				StatusConditions(
					factories.Condition().Type(knativeservingv1.ConfigurationConditionReady).True(),
				),
			testRouteGiven.
				StatusConditions(
					factories.Condition().Type(knativeservingv1.RouteConditionReady).True(),
				).
				StatusAddressURL(testAddressURL).
				StatusURL(testURL),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.True(),
					deployerConditionReady.True(),
					deployerConditionRouteReady.True(),
				).
				StatusLatestImage(testImage).
				StatusConfigurationRef(testConfigurationGiven.Create().GetName()).
				StatusRouteRef(testRouteGiven.Create().GetName()).
				StatusAddressURL(testAddressURL).
				StatusURL(testURL),
		},
	}, {
		Name: "not ready, configuration",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testDeployer.
				Image(testImage),
			testConfigurationGiven.
				StatusConditions(
					factories.Condition().Type(knativeservingv1.ConfigurationConditionReady).False().Reason("TestReason", "a human readable message"),
				),
			testRouteGiven.
				StatusReady().
				StatusAddressURL(testAddressURL).
				StatusURL(testURL),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.False().Reason("TestReason", "a human readable message"),
					deployerConditionReady.False().Reason("TestReason", "a human readable message"),
					deployerConditionRouteReady.True(),
				).
				StatusLatestImage(testImage).
				StatusConfigurationRef(testConfigurationGiven.Create().GetName()).
				StatusRouteRef(testRouteGiven.Create().GetName()).
				StatusAddressURL(testAddressURL).
				StatusURL(testURL),
		},
	}, {
		Name: "not ready, route",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			testDeployer.
				Image(testImage),
			testConfigurationGiven.
				StatusReady(),
			testRouteGiven.
				StatusConditions(
					factories.Condition().Type(knativeservingv1.RouteConditionReady).False().Reason("TestReason", "a human readable message"),
				).
				StatusAddressURL(testAddressURL).
				StatusURL(testURL),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(testDeployer, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			testDeployer.
				StatusConditions(
					deployerConditionConfigurationReady.True(),
					deployerConditionReady.False().Reason("TestReason", "a human readable message"),
					deployerConditionRouteReady.False().Reason("TestReason", "a human readable message"),
				).
				StatusLatestImage(testImage).
				StatusConfigurationRef(testConfigurationGiven.Create().GetName()).
				StatusRouteRef(testRouteGiven.Create().GetName()).
				StatusAddressURL(testAddressURL).
				StatusURL(testURL),
		},
	}}

	table.Test(t, scheme, func(t *testing.T, row *rtesting.Testcase, client client.Client, apiReader client.Reader, tracker tracker.Tracker, recorder record.EventRecorder, log logr.Logger) reconcile.Reconciler {
		return knative.DeployerReconciler(
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
