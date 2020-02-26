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
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	streamingv1alpha1 "github.com/projectriff/riff/system/pkg/apis/streaming/v1alpha1"
	"github.com/projectriff/riff/system/pkg/controllers"
	"github.com/projectriff/riff/system/pkg/controllers/streaming"
	rtesting "github.com/projectriff/riff/system/pkg/controllers/testing"
	"github.com/projectriff/riff/system/pkg/controllers/testing/factories"
	"github.com/projectriff/riff/system/pkg/tracker"
)

func TestPulsarGatewayReconciler(t *testing.T) {
	testNamespace := "test-namespace"
	testSystemNamespace := "system-namespace"
	testName := "test-gateway"
	testKey := types.NamespacedName{Namespace: testNamespace, Name: testName}
	testImagePrefix := "example.com/repo"
	testGatewayImage := fmt.Sprintf("%s/%s", testImagePrefix, "gateway")
	testProvisionerImage := fmt.Sprintf("%s/%s", testImagePrefix, "provisioner")
	testProvisionerHostname := fmt.Sprintf("%s.%s.svc.cluster.local", testName, testNamespace)
	testProvisionerURL := fmt.Sprintf("http://%s", testProvisionerHostname)
	testServiceURL := "pulsar://pulsar.local:6650"

	pulsarGatewayImages := "riff-streaming-pulsar-gateway" // contains image names for the pulsar gateway
	gatewayImageKey := "gatewayImage"
	provisionerImageKey := "provisionerImage"

	pulsarGatewayConditionGatewayReady := factories.Condition().Type(streamingv1alpha1.PulsarGatewayConditionGatewayReady)
	pulsarGatewayConditionReady := factories.Condition().Type(streamingv1alpha1.PulsarGatewayConditionReady)
	gatewayConditionReady := factories.Condition().Type(streamingv1alpha1.GatewayConditionReady)

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = streamingv1alpha1.AddToScheme(scheme)

	pulsarGatewayImagesConfigMap := factories.ConfigMap().
		NamespaceName(testSystemNamespace, pulsarGatewayImages).
		AddData(gatewayImageKey, testGatewayImage).
		AddData(provisionerImageKey, testProvisionerImage)

	pulsarGatewayMinimal := factories.PulsarGateway().
		NamespaceName(testNamespace, testName).
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Created(1)
			om.Generation(1)
		}).
		ServiceURL(testServiceURL)
	pulsarGateway := pulsarGatewayMinimal.
		StatusGatewayRef(testName).
		StatusGatewayImage(testGatewayImage).
		StatusProvisionerImage(testProvisionerImage)
	pulsarGatewayReady := pulsarGateway.
		StatusObservedGeneration(1).
		StatusConditions(
			pulsarGatewayConditionGatewayReady.True(),
			pulsarGatewayConditionReady.True(),
		)

	gatewayCreate := factories.Gateway().
		NamespaceName(testNamespace, testName).
		ObjectMeta(func(om factories.ObjectMeta) {
			om.AddLabel(streamingv1alpha1.PulsarGatewayLabelKey, testName)
			om.AddLabel(streamingv1alpha1.GatewayTypeLabelKey, streamingv1alpha1.PulsarGatewayType)
			om.ControlledBy(pulsarGateway, scheme)
		}).
		Ports(
			corev1.ServicePort{Name: "gateway", Port: 6565},
			corev1.ServicePort{Name: "provisioner", Port: 80, TargetPort: intstr.FromInt(8080)},
		)
	gatewayGiven := gatewayCreate.
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Created(1)
		}).
		StatusAddress(testProvisionerURL)
	gatewayComplete := gatewayGiven.
		PodTemplateSpec(func(pts factories.PodTemplateSpec) {
			pts.AddLabel(streamingv1alpha1.PulsarGatewayLabelKey, testName)
			pts.AddLabel(streamingv1alpha1.GatewayTypeLabelKey, streamingv1alpha1.PulsarGatewayType)
			pts.ContainerNamed("gateway", func(c *corev1.Container) {
				c.Image = testGatewayImage
				c.Env = []corev1.EnvVar{
					{Name: "pulsar_serviceUrl", Value: testServiceURL},
					{Name: "storage_positions_type", Value: "MEMORY"},
					{Name: "storage_records_type", Value: "PULSAR"},
					{Name: "server_port", Value: "8000"},
				}
			})
			pts.ContainerNamed("provisioner", func(c *corev1.Container) {
				c.Image = testProvisionerImage
				c.Env = []corev1.EnvVar{
					{Name: "GATEWAY", Value: fmt.Sprintf("%s:6565", testProvisionerHostname)},
					{Name: "BROKER", Value: testServiceURL},
					{Name: "TENANT", Value: "public"},
					{Name: "NAMESPACE", Value: "default"},
				}
			})
		})

	table := rtesting.Table{{
		Name: "pulsargateway does not exist",
		Key:  testKey,
	}, {
		Name: "ignore deleted pulsargateway",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			pulsarGateway.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.Deleted(1)
				}),
		},
	}, {
		Name: "error fetching pulsargateway",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("get", "PulsarGateway"),
		},
		GivenObjects: []rtesting.Factory{
			pulsarGateway,
		},
		ShouldErr: true,
	}, {
		Name: "creates gateway",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			pulsarGatewayMinimal,
			pulsarGatewayImagesConfigMap,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(pulsarGatewayImagesConfigMap, pulsarGateway, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(pulsarGateway, scheme, corev1.EventTypeNormal, "Created",
				`Created Gateway "%s"`, testName),
			rtesting.NewEvent(pulsarGateway, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			gatewayCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			pulsarGateway.
				StatusObservedGeneration(1).
				StatusConditions(
					pulsarGatewayConditionGatewayReady.Unknown(),
					pulsarGatewayConditionReady.Unknown(),
				),
		},
	}, {
		Name: "propagate address",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			pulsarGateway,
			pulsarGatewayImagesConfigMap,
			gatewayGiven,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(pulsarGatewayImagesConfigMap, pulsarGateway, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(pulsarGateway, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			pulsarGateway.
				StatusAddress(testProvisionerURL).
				StatusObservedGeneration(1).
				StatusConditions(
					pulsarGatewayConditionGatewayReady.Unknown(),
					pulsarGatewayConditionReady.Unknown(),
				),
		},
	}, {
		Name: "updates gateway",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			pulsarGateway.
				StatusAddress(testProvisionerURL),
			pulsarGatewayImagesConfigMap,
			gatewayGiven,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(pulsarGatewayImagesConfigMap, pulsarGateway, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(pulsarGateway, scheme, corev1.EventTypeNormal, "Updated",
				`Updated Gateway "%s"`, testName),
			rtesting.NewEvent(pulsarGateway, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectUpdates: []rtesting.Factory{
			gatewayComplete,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			pulsarGateway.
				StatusAddress(testProvisionerURL).
				StatusObservedGeneration(1).
				StatusConditions(
					pulsarGatewayConditionGatewayReady.Unknown(),
					pulsarGatewayConditionReady.Unknown(),
				),
		},
	}, {
		Name: "ready",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			pulsarGateway.
				StatusAddress(testProvisionerURL),
			pulsarGatewayImagesConfigMap,
			gatewayComplete.
				StatusReady(),
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(pulsarGatewayImagesConfigMap, pulsarGateway, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(pulsarGateway, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			pulsarGateway.
				StatusAddress(testProvisionerURL).
				StatusObservedGeneration(1).
				StatusConditions(
					pulsarGatewayConditionGatewayReady.True(),
					pulsarGatewayConditionReady.True(),
				),
		},
	}, {
		Name: "not ready",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			pulsarGateway.
				StatusAddress(testProvisionerURL),
			pulsarGatewayImagesConfigMap,
			gatewayComplete.
				StatusConditions(
					gatewayConditionReady.False().Reason("TestReason", "a human readable message"),
				),
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(pulsarGatewayImagesConfigMap, pulsarGateway, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(pulsarGateway, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			pulsarGateway.
				StatusAddress(testProvisionerURL).
				StatusObservedGeneration(1).
				StatusConditions(
					pulsarGatewayConditionGatewayReady.False().Reason("TestReason", "a human readable message"),
					pulsarGatewayConditionReady.False().Reason("TestReason", "a human readable message"),
				),
		},
	}, {
		Name: "missing gateway images configmap",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			pulsarGatewayReady,
		},
		ShouldErr: true,
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(pulsarGatewayImagesConfigMap, pulsarGateway, scheme),
		},
	}, {
		Name: "invalid address",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			pulsarGatewayReady.
				StatusAddress("\000"),
			pulsarGatewayImagesConfigMap,
			gatewayGiven.
				StatusAddress("\000"),
		},
		ShouldErr: true,
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(pulsarGatewayImagesConfigMap, pulsarGateway, scheme),
		},
	}, {
		Name: "conflicting gateway",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("create", "Gateway", rtesting.InduceFailureOpts{
				Error: apierrs.NewAlreadyExists(schema.GroupResource{}, testName),
			}),
		},
		GivenObjects: []rtesting.Factory{
			pulsarGatewayMinimal,
			pulsarGatewayImagesConfigMap,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(pulsarGatewayImagesConfigMap, pulsarGateway, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(pulsarGateway, scheme, corev1.EventTypeWarning, "CreationFailed",
				`Failed to create Gateway "%s":  "%s" already exists`, testName, testName),
			rtesting.NewEvent(pulsarGateway, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			gatewayCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			pulsarGatewayMinimal.
				StatusObservedGeneration(1).
				StatusConditions(
					pulsarGatewayConditionGatewayReady.False().Reason("NotOwned", `There is an existing Gateway "test-gateway" that the PulsarGateway does not own.`),
					pulsarGatewayConditionReady.False().Reason("NotOwned", `There is an existing Gateway "test-gateway" that the PulsarGateway does not own.`),
				).
				StatusGatewayImage(testGatewayImage).
				StatusProvisionerImage(testProvisionerImage),
		},
	}, {
		Name: "conflicting gateway, owned",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("create", "Gateway", rtesting.InduceFailureOpts{
				Error: apierrs.NewAlreadyExists(schema.GroupResource{}, testName),
			}),
		},
		GivenObjects: []rtesting.Factory{
			pulsarGatewayMinimal,
			pulsarGatewayImagesConfigMap,
		},
		APIGivenObjects: []rtesting.Factory{
			gatewayGiven,
		},
		ShouldErr: true,
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(pulsarGatewayImagesConfigMap, pulsarGateway, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(pulsarGateway, scheme, corev1.EventTypeWarning, "CreationFailed",
				`Failed to create Gateway "%s":  "%s" already exists`, testName, testName),
			rtesting.NewEvent(pulsarGateway, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			gatewayCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			pulsarGatewayMinimal.
				StatusConditions(
					pulsarGatewayConditionGatewayReady.Unknown(),
					pulsarGatewayConditionReady.Unknown(),
				).
				StatusGatewayImage(testGatewayImage).
				StatusProvisionerImage(testProvisionerImage),
		},
	}}

	table.Test(t, scheme, func(t *testing.T, row *rtesting.Testcase, client client.Client, apiReader client.Reader, tracker tracker.Tracker, recorder record.EventRecorder, log logr.Logger) reconcile.Reconciler {
		return streaming.PulsarGatewayReconciler(
			controllers.Config{
				Client:    client,
				APIReader: apiReader,
				Recorder:  recorder,
				Log:       log,
				Scheme:    scheme,
				Tracker:   tracker,
			},
			testSystemNamespace,
		)
	})
}
