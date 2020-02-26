/*
Copyright 2020 the original author or authors.

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

package streaming_test

import (
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/projectriff/system/pkg/apis"
	streamingv1alpha1 "github.com/projectriff/system/pkg/apis/streaming/v1alpha1"
	"github.com/projectriff/system/pkg/controllers"
	"github.com/projectriff/system/pkg/controllers/streaming"
	rtesting "github.com/projectriff/system/pkg/controllers/testing"
	"github.com/projectriff/system/pkg/controllers/testing/factories"
	"github.com/projectriff/system/pkg/tracker"
)

func TestStreamReconciler(t *testing.T) {
	var streamProvisioner *streaming.MockStreamProvisionerClient

	testNamespace := "test-namespace"
	testName := "test-stream"
	testKey := types.NamespacedName{Namespace: testNamespace, Name: testName}
	testGateway := "test-gateway"
	testBindingMetadata := fmt.Sprintf("%s-stream-binding-metadata", testName)
	testBindingSecret := fmt.Sprintf("%s-stream-binding-secret", testName)
	testProvisionerHost := fmt.Sprintf("%s.%s.svc.cluster.local", testGateway, testNamespace)
	testProvisionerURL := fmt.Sprintf("http://%s/%s/%s", testProvisionerHost, testNamespace, testName)
	testAddressGateway := fmt.Sprintf("%s:6565", testProvisionerHost)
	testAddressTopic := fmt.Sprintf("%s/%s", testNamespace, testName)
	testAddress := &streaming.StreamAddress{Gateway: testAddressGateway, Topic: testAddressTopic}

	streamConditionBindingReady := factories.Condition().Type(streamingv1alpha1.StreamConditionBindingReady)
	streamConditionReady := factories.Condition().Type(streamingv1alpha1.StreamConditionReady)
	streamConditionResourceAvailable := factories.Condition().Type(streamingv1alpha1.StreamConditionResourceAvailable)
	gatewayConditionReady := factories.Condition().Type(streamingv1alpha1.GatewayConditionReady)

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = streamingv1alpha1.AddToScheme(scheme)

	streamMinimal := factories.Stream().
		NamespaceName(testNamespace, testName).
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Created(1)
			om.Generation(1)
		})
	stream := streamMinimal.
		Gateway(testGateway).
		ContentType("text/plain")
	streamReady := stream.
		StatusObservedGeneration(1).
		StatusConditions(
			streamConditionBindingReady.True(),
			streamConditionReady.True(),
			streamConditionResourceAvailable.True(),
		).
		StatusBinding(testBindingMetadata, testBindingSecret)

	bindingMetadataCreate := factories.ConfigMap().
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Namespace(testNamespace)
			om.Name(testBindingMetadata)
			om.AddLabel(streamingv1alpha1.StreamLabelKey, testName)
			om.ControlledBy(stream, scheme)
		}).
		AddData("contentType", "text/plain").
		AddData("kind", "Stream.streaming.projectriff.io").
		AddData("provider", "riff Streaming").
		AddData("stream", testName).
		AddData("tags", "")
	bindingMetadataGiven := bindingMetadataCreate.
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Created(1)
		})

	bindingSecretCreate := factories.Secret().
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Namespace(testNamespace)
			om.Name(testBindingSecret)
			om.AddLabel(streamingv1alpha1.StreamLabelKey, testName)
			om.ControlledBy(stream, scheme)
		}).
		AddData("gateway", testAddressGateway).
		AddData("topic", testAddressTopic)
	bindingSecretGiven := bindingSecretCreate.
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Created(1)
		})

	gatewayMinimal := factories.Gateway().
		NamespaceName(testNamespace, testGateway)
	gateway := gatewayMinimal.
		StatusConditions(
			gatewayConditionReady.True(),
		).
		StatusAddress(testProvisionerURL)

	matchedByObject := func(expectedFactory rtesting.Factory) interface{} {
		return mock.MatchedBy(func(actual apis.Object) bool {
			expected := expectedFactory.CreateObject()

			return expected.GetNamespace() == actual.GetNamespace() &&
				expected.GetName() == actual.GetName()
		})
	}

	table := rtesting.Table{{
		Name: "stream does not exist",
		Key:  testKey,
	}, {
		Name: "ignore deleted stream",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			stream.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.Deleted(1)
				}),
			gateway,
		},
	}, {
		Name: "error fetching stream",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("get", "Stream"),
		},
		GivenObjects: []rtesting.Factory{
			stream,
			gateway,
		},
		ShouldErr: true,
	}, {
		Name: "provision and create binding",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			stream,
			gateway,
		},
		Prepare: func(t *testing.T) error {
			streamProvisioner.On("ProvisionStream", matchedByObject(stream), testProvisionerURL).Return(testAddress, nil)
			return nil
		},
		CleanUp: func(t *testing.T) error {
			streamProvisioner.AssertExpectations(t)
			return nil
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(gateway, stream, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(stream, scheme, corev1.EventTypeNormal, "Created",
				`Created ConfigMap "%s"`, testBindingMetadata),
			rtesting.NewEvent(stream, scheme, corev1.EventTypeNormal, "Created",
				`Created Secret "%s"`, testBindingSecret),
			rtesting.NewEvent(stream, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			bindingMetadataCreate,
			bindingSecretCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			streamReady,
		},
	}, {
		Name: "update binding",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			streamReady,
			gateway,
			bindingMetadataGiven.
				AddData("foo", "bar"),
			bindingSecretGiven.
				AddData("foo", "bar"),
		},
		Prepare: func(t *testing.T) error {
			streamProvisioner.On("ProvisionStream", matchedByObject(stream), testProvisionerURL).Return(testAddress, nil)
			return nil
		},
		CleanUp: func(t *testing.T) error {
			streamProvisioner.AssertExpectations(t)
			return nil
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(gateway, stream, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(stream, scheme, corev1.EventTypeNormal, "Updated",
				`Updated ConfigMap "%s"`, testBindingMetadata),
			rtesting.NewEvent(stream, scheme, corev1.EventTypeNormal, "Updated",
				`Updated Secret "%s"`, testBindingSecret),
		},
		ExpectUpdates: []rtesting.Factory{
			bindingMetadataGiven,
			bindingSecretGiven,
		},
	}, {
		Name: "missing gateway",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			stream,
		},
		Prepare: func(t *testing.T) error {
			return nil
		},
		CleanUp: func(t *testing.T) error {
			streamProvisioner.AssertExpectations(t)
			return nil
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(gateway, stream, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(stream, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			stream.
				StatusObservedGeneration(1).
				StatusConditions(
					streamConditionBindingReady.Unknown(),
					streamConditionReady.False().Reason("ProvisionFailed", `Gateway "test-gateway" not found`),
					streamConditionResourceAvailable.False().Reason("ProvisionFailed", `Gateway "test-gateway" not found`),
				),
		},
	}, {
		Name: "error fetching gateway",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("get", "Gateway"),
		},
		GivenObjects: []rtesting.Factory{
			stream,
			gateway,
		},
		Prepare: func(t *testing.T) error {
			return nil
		},
		CleanUp: func(t *testing.T) error {
			streamProvisioner.AssertExpectations(t)
			return nil
		},
		ShouldErr: true,
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(gateway, stream, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(stream, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			stream.
				StatusConditions(
					streamConditionBindingReady.Unknown(),
					streamConditionReady.False().Reason("ProvisionFailed", "inducing failure for get Gateway"),
					streamConditionResourceAvailable.False().Reason("ProvisionFailed", "inducing failure for get Gateway"),
				),
		},
	}, {
		Name: "gateway is not ready",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			stream,
			gateway.
				StatusConditions(
					gatewayConditionReady.Unknown(),
				),
		},
		Prepare: func(t *testing.T) error {
			return nil
		},
		CleanUp: func(t *testing.T) error {
			streamProvisioner.AssertExpectations(t)
			return nil
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(gateway, stream, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(stream, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			stream.
				StatusObservedGeneration(1).
				StatusConditions(
					streamConditionBindingReady.Unknown(),
					streamConditionReady.False().Reason("ProvisionFailed", `Gateway "test-gateway" not ready`),
					streamConditionResourceAvailable.False().Reason("ProvisionFailed", `Gateway "test-gateway" not ready`),
				),
		},
	}, {
		Name: "gateway has invalid address",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			stream,
			gateway.
				StatusAddress("\n"),
		},
		Prepare: func(t *testing.T) error {
			return nil
		},
		CleanUp: func(t *testing.T) error {
			streamProvisioner.AssertExpectations(t)
			return nil
		},
		ShouldErr: true,
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(gateway, stream, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(stream, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			stream.
				StatusConditions(
					streamConditionBindingReady.Unknown(),
					streamConditionReady.Unknown(),
					streamConditionResourceAvailable.Unknown(),
				),
		},
	}, {
		Name: "provision failed",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			stream,
			gateway,
		},
		Prepare: func(t *testing.T) error {
			streamProvisioner.On("ProvisionStream", matchedByObject(stream), testProvisionerURL).Return(nil, fmt.Errorf("remote error"))
			return nil
		},
		CleanUp: func(t *testing.T) error {
			streamProvisioner.AssertExpectations(t)
			return nil
		},
		ShouldErr: true,
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(gateway, stream, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(stream, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			stream.
				StatusConditions(
					streamConditionBindingReady.Unknown(),
					streamConditionReady.False().Reason("ProvisionFailed", "remote error"),
					streamConditionResourceAvailable.False().Reason("ProvisionFailed", "remote error"),
				),
		},
	}, {
		Name: "conflicting binding metadata",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			streamReady,
			gateway,
			bindingSecretGiven,
		},
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("create", "ConfigMap", rtesting.InduceFailureOpts{
				Error: apierrs.NewAlreadyExists(schema.GroupResource{}, testBindingMetadata),
			}),
		},
		Prepare: func(t *testing.T) error {
			streamProvisioner.On("ProvisionStream", matchedByObject(stream), testProvisionerURL).Return(testAddress, nil)
			return nil
		},
		CleanUp: func(t *testing.T) error {
			streamProvisioner.AssertExpectations(t)
			return nil
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(gateway, stream, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(stream, scheme, corev1.EventTypeWarning, "CreationFailed",
				`Failed to create ConfigMap "%s":  "%s" already exists`, testBindingMetadata, testBindingMetadata),
			rtesting.NewEvent(stream, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			bindingMetadataCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			stream.
				StatusObservedGeneration(1).
				StatusConditions(
					streamConditionBindingReady.False().Reason("BindingFailed", `binding metadata "test-stream-stream-binding-metadata" already exists`),
					streamConditionReady.False().Reason("BindingFailed", `binding metadata "test-stream-stream-binding-metadata" already exists`),
					streamConditionResourceAvailable.True(),
				).
				StatusBinding("", testBindingSecret),
		},
	}, {
		Name: "conflicting binding secret",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			streamReady,
			gateway,
			bindingMetadataGiven,
		},
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("create", "Secret", rtesting.InduceFailureOpts{
				Error: apierrs.NewAlreadyExists(schema.GroupResource{}, testBindingSecret),
			}),
		},
		Prepare: func(t *testing.T) error {
			streamProvisioner.On("ProvisionStream", matchedByObject(stream), testProvisionerURL).Return(testAddress, nil)
			return nil
		},
		CleanUp: func(t *testing.T) error {
			streamProvisioner.AssertExpectations(t)
			return nil
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(gateway, stream, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(stream, scheme, corev1.EventTypeWarning, "CreationFailed",
				`Failed to create Secret "%s":  "%s" already exists`, testBindingSecret, testBindingSecret),
			rtesting.NewEvent(stream, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			bindingSecretCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			stream.
				StatusObservedGeneration(1).
				StatusConditions(
					streamConditionBindingReady.False().Reason("BindingFailed", `binding secret "test-stream-stream-binding-secret" already exists`),
					streamConditionReady.False().Reason("BindingFailed", `binding secret "test-stream-stream-binding-secret" already exists`),
					streamConditionResourceAvailable.True(),
				).
				StatusBinding(testBindingMetadata, ""),
		},
	}}

	table.Test(t, scheme, func(t *testing.T, row *rtesting.Testcase, client client.Client, apiReader client.Reader, tracker tracker.Tracker, recorder record.EventRecorder, log logr.Logger) reconcile.Reconciler {
		streamProvisioner = &streaming.MockStreamProvisionerClient{}
		return streaming.StreamReconciler(
			controllers.Config{
				Client:    client,
				APIReader: apiReader,
				Recorder:  recorder,
				Log:       log,
				Scheme:    scheme,
				Tracker:   tracker,
			},
			streamProvisioner,
		)
	})
}
