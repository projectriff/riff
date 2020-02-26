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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	streamingv1alpha1 "github.com/projectriff/system/pkg/apis/streaming/v1alpha1"
	kedav1alpha1 "github.com/projectriff/system/pkg/apis/thirdparty/keda/v1alpha1"
	"github.com/projectriff/system/pkg/controllers"
	"github.com/projectriff/system/pkg/controllers/streaming"
	rtesting "github.com/projectriff/system/pkg/controllers/testing"
	"github.com/projectriff/system/pkg/controllers/testing/factories"
	"github.com/projectriff/system/pkg/tracker"
)

func TestProcessorReconciler(t *testing.T) {
	testNamespace := "test-namespace"
	testSystemNamespace := "system-namespace"
	testName := "test-processor"
	testKey := types.NamespacedName{Namespace: testNamespace, Name: testName}
	testImagePrefix := "example.com/repo"
	testSha256 := "cf8b4c69d5460f88530e1c80b8856a70801f31c50b191c8413043ba9b160a43e"
	testImage := fmt.Sprintf("%s@sha256:%s", testImagePrefix, testSha256)
	testProcessorImage := fmt.Sprintf("%s/processor", testImagePrefix)

	processorConditionDeploymentReady := factories.Condition().Type(streamingv1alpha1.ProcessorConditionDeploymentReady)
	processorConditionReady := factories.Condition().Type(streamingv1alpha1.ProcessorConditionReady)
	processorConditionScaledObjectReady := factories.Condition().Type(streamingv1alpha1.ProcessorConditionScaledObjectReady)
	processorConditionStreamsReady := factories.Condition().Type(streamingv1alpha1.ProcessorConditionStreamsReady)
	deploymentConditionAvailable := factories.Condition().Type("Available")
	deploymentConditionProgressing := factories.Condition().Type("Progressing")

	processorImages := "riff-streaming-processor"
	processorImageKey := "processorImage"

	processorImagesConfigMap := factories.ConfigMap().
		NamespaceName(testSystemNamespace, processorImages).
		AddData(processorImageKey, testProcessorImage)

	testFunction := factories.Function().
		NamespaceName(testNamespace, "my-function").
		StatusLatestImage(testImage)
	testContainer := factories.Container().
		NamespaceName(testNamespace, "my-container").
		StatusLatestImage(testImage)

	testStream1 := factories.Stream().
		NamespaceName(testNamespace, "stream-1").
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Created(1)
			om.UID("00000000-0000-0000-0000-000000000001")
		}).
		ContentType("text/plain").
		StatusBinding("stream-1-binding-metadata", "stream-1-binding-secret").
		StatusReady()
	testStream2 := factories.Stream().
		NamespaceName(testNamespace, "stream-2").
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Created(1)
			om.UID("00000000-0000-0000-0000-000000000002")
		}).
		ContentType("text/plain").
		StatusBinding("stream-2-binding-metadata", "stream-2-binding-secret").
		StatusReady()
	testStream3 := factories.Stream().
		NamespaceName(testNamespace, "stream-3").
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Created(1)
			om.UID("00000000-0000-0000-0000-000000000003")
		}).
		ContentType("text/plain").
		StatusBinding("stream-3-binding-metadata", "stream-3-binding-secret").
		StatusReady()
	testStream4 := factories.Stream().
		NamespaceName(testNamespace, "stream-4").
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Created(1)
			om.UID("00000000-0000-0000-0000-000000000004")
		}).
		ContentType("text/plain").
		StatusBinding("stream-4-binding-metadata", "stream-4-binding-secret").
		StatusReady()

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = streamingv1alpha1.AddToScheme(scheme)
	_ = kedav1alpha1.AddToScheme(scheme)
	_ = buildv1alpha1.AddToScheme(scheme)

	processorMinimal := factories.Processor().
		NamespaceName(testNamespace, testName).
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Created(1)
			om.Generation(1)
		})
	processor := processorMinimal.
		Default().
		StatusLatestImage(testImage).
		StatusDeploymentRef("%s-processor-000", testName).
		StatusScaledObjectRef("%s-processor-000", testName)

	deploymentCreate := factories.Deployment().
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Namespace(testNamespace)
			om.GenerateName("%s-processor-", testName)
			om.AddLabel(streamingv1alpha1.ProcessorLabelKey, testName)
			om.ControlledBy(processor, scheme)
		}).
		AddSelectorLabel(streamingv1alpha1.ProcessorLabelKey, testName).
		PodTemplateSpec(func(pts factories.PodTemplateSpec) {
			pts.ContainerNamed("function", func(c *corev1.Container) {
				c.Image = testImage
				c.Ports = []corev1.ContainerPort{
					{ContainerPort: 8081},
				}
			})
			pts.ContainerNamed("processor", func(c *corev1.Container) {
				c.Image = testProcessorImage
				c.Env = []corev1.EnvVar{
					{Name: "CNB_BINDINGS", Value: "/var/riff/bindings"},
					{Name: "INPUT_START_OFFSETS", Value: ""},
					{Name: "INPUT_NAMES", Value: ""},
					{Name: "OUTPUT_NAMES", Value: ""},
					{Name: "GROUP", Value: testName},
					{Name: "FUNCTION", Value: "localhost:8081"},
				}
			})
		}).
		Replicas(1)
	deploymentGiven := deploymentCreate.
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Name("%s-processor-000", testName)
			om.Created(1)
		})

	scaledObjectCreate := factories.KedaScaledObject().
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Namespace(testNamespace)
			om.GenerateName("%s-processor-", testName)
			om.AddLabel(streamingv1alpha1.ProcessorLabelKey, testName)
			om.ControlledBy(processor, scheme)
		}).
		ScaleTargetRefDeployment("%s-processor-000", testName).
		PollingInterval(1).
		CooldownPeriod(30).
		MinReplicaCount(1).
		MaxReplicaCount(30)
	scaledObjectGiven := scaledObjectCreate.
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Name("%s-processor-000", testName)
			om.Created(1)
		})

	t.Run("ProcessorReconciler", func(t *testing.T) {
		table := rtesting.Table{{
			Name: "processor does not exist",
			Key:  testKey,
		}, {
			Name: "ignore deleted processor",
			Key:  testKey,
			GivenObjects: []rtesting.Factory{
				processor.
					ObjectMeta(func(om factories.ObjectMeta) {
						om.Deleted(1)
					}),
			},
		}, {
			Name: "error fetching processor",
			Key:  testKey,
			WithReactors: []rtesting.ReactionFunc{
				rtesting.InduceFailure("get", "Processor"),
			},
			GivenObjects: []rtesting.Factory{
				processor,
			},
			ShouldErr: true,
		}, {
			Name: "create resources",
			Key:  testKey,
			GivenObjects: []rtesting.Factory{
				processor.
					Image(testImage),
				processorImagesConfigMap,
			},
			ExpectTracks: []rtesting.TrackRequest{
				rtesting.NewTrackRequest(processorImagesConfigMap, processor, scheme),
			},
			ExpectEvents: []rtesting.Event{
				rtesting.NewEvent(processor, scheme, corev1.EventTypeNormal, "Created",
					`Created Deployment "%s-processor-001"`, testName),
				rtesting.NewEvent(processor, scheme, corev1.EventTypeNormal, "Created",
					`Created ScaledObject "%s-processor-002"`, testName),
				rtesting.NewEvent(processor, scheme, corev1.EventTypeNormal, "StatusUpdated",
					`Updated status`),
			},
			ExpectCreates: []rtesting.Factory{
				deploymentCreate,
				scaledObjectCreate.
					ScaleTargetRefDeployment("%s-processor-001", testName),
			},
			ExpectStatusUpdates: []rtesting.Factory{
				processorMinimal.
					StatusObservedGeneration(1).
					StatusConditions(
						processorConditionDeploymentReady.Unknown(),
						processorConditionReady.Unknown(),
						processorConditionScaledObjectReady.True(),
						processorConditionStreamsReady.True(),
					).
					StatusLatestImage(testImage).
					StatusDeploymentRef("%s-processor-001", testName).
					StatusScaledObjectRef("%s-processor-002", testName),
			},
		}, {
			Name: "ready",
			Key:  testKey,
			GivenObjects: []rtesting.Factory{
				processor.
					Image(testImage),
				processorImagesConfigMap,
				deploymentGiven.
					StatusConditions(
						deploymentConditionAvailable.True(),
						deploymentConditionProgressing.True(),
					),
				scaledObjectGiven.
					ScaleTargetRefDeployment("%s-processor-000", testName),
			},
			ExpectTracks: []rtesting.TrackRequest{
				rtesting.NewTrackRequest(processorImagesConfigMap, processor, scheme),
			},
			ExpectEvents: []rtesting.Event{
				rtesting.NewEvent(processor, scheme, corev1.EventTypeNormal, "StatusUpdated",
					`Updated status`),
			},
			ExpectStatusUpdates: []rtesting.Factory{
				processorMinimal.
					StatusObservedGeneration(1).
					StatusConditions(
						processorConditionDeploymentReady.True(),
						processorConditionReady.True(),
						processorConditionScaledObjectReady.True(),
						processorConditionStreamsReady.True(),
					).
					StatusLatestImage(testImage).
					StatusDeploymentRef("%s-processor-000", testName).
					StatusScaledObjectRef("%s-processor-000", testName),
			},
		}}

		table.Test(t, scheme, func(t *testing.T, row *rtesting.Testcase, client client.Client, apiReader client.Reader, tracker tracker.Tracker, recorder record.EventRecorder, log logr.Logger) reconcile.Reconciler {
			return streaming.ProcessorReconciler(
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
	})

	t.Run("ProcessorSyncProcessorImages", func(t *testing.T) {
		table := rtesting.SubTable{
			{
				Name:   "missing images configmap",
				Parent: processor,
				ExpectTracks: []rtesting.TrackRequest{
					rtesting.NewTrackRequest(processorImagesConfigMap, processor, scheme),
				},
				ShouldErr: true,
			},
			{
				Name:   "stash processor image",
				Parent: processor,
				GivenObjects: []rtesting.Factory{
					processorImagesConfigMap,
				},
				ExpectTracks: []rtesting.TrackRequest{
					rtesting.NewTrackRequest(processorImagesConfigMap, processor, scheme),
				},
				ExpectStashedValues: map[controllers.StashKey]interface{}{
					streaming.ProcessorImagesStashKey: processorImagesConfigMap.Create().Data,
				},
			},
		}

		table.Test(t, scheme, func(t *testing.T, row *rtesting.SubTestcase, client client.Client, tracker tracker.Tracker, recorder record.EventRecorder, log logr.Logger) controllers.SubReconciler {
			return streaming.ProcessorSyncProcessorImages(
				controllers.Config{
					Client:    client,
					APIReader: client,
					Recorder:  recorder,
					Log:       log,
					Scheme:    scheme,
					Tracker:   tracker,
				},
				testSystemNamespace,
			)
		})
	})

	t.Run("ProcessorBuildRefReconciler", func(t *testing.T) {
		table := rtesting.SubTable{
			{
				Name: "use specified image",
				Parent: processor.
					Image(testImage),
				ExpectParent: processor.
					Image(testImage).
					StatusLatestImage(testImage),
			},
			{
				Name: "container build",
				Parent: processor.
					BuildContainerRef(testContainer),
				GivenObjects: []rtesting.Factory{
					testContainer,
				},
				ExpectParent: processor.
					BuildContainerRef(testContainer).
					StatusLatestImage(testImage),
				ExpectTracks: []rtesting.TrackRequest{
					rtesting.NewTrackRequest(testContainer, processor, scheme),
				},
			},
			{
				Name: "container build, not found",
				Parent: processor.
					BuildContainerRef(testContainer),
				ExpectTracks: []rtesting.TrackRequest{
					rtesting.NewTrackRequest(testContainer, processor, scheme),
				},
			},
			{
				Name: "container build, lookup error",
				Parent: processor.
					BuildContainerRef(testContainer),
				WithReactors: []rtesting.ReactionFunc{
					rtesting.InduceFailure("get", "Container"),
				},
				ShouldErr: true,
				ExpectTracks: []rtesting.TrackRequest{
					rtesting.NewTrackRequest(testContainer, processor, scheme),
				},
			},
			{
				Name: "function build",
				Parent: processor.
					BuildFunctionRef(testFunction),
				GivenObjects: []rtesting.Factory{
					testFunction,
				},
				ExpectParent: processor.
					BuildFunctionRef(testFunction).
					StatusLatestImage(testImage),
				ExpectTracks: []rtesting.TrackRequest{
					rtesting.NewTrackRequest(testFunction, processor, scheme),
				},
			},
			{
				Name: "function build, not found",
				Parent: processor.
					BuildFunctionRef(testFunction),
				ExpectTracks: []rtesting.TrackRequest{
					rtesting.NewTrackRequest(testFunction, processor, scheme),
				},
			},
			{
				Name: "function build, lookup error",
				Parent: processor.
					BuildFunctionRef(testFunction),
				WithReactors: []rtesting.ReactionFunc{
					rtesting.InduceFailure("get", "Function"),
				},
				ShouldErr: true,
				ExpectTracks: []rtesting.TrackRequest{
					rtesting.NewTrackRequest(testFunction, processor, scheme),
				},
			},
		}

		table.Test(t, scheme, func(t *testing.T, row *rtesting.SubTestcase, client client.Client, tracker tracker.Tracker, recorder record.EventRecorder, log logr.Logger) controllers.SubReconciler {
			return streaming.ProcessorBuildRefReconciler(
				controllers.Config{
					Client:    client,
					APIReader: client,
					Recorder:  recorder,
					Log:       log,
					Scheme:    scheme,
					Tracker:   tracker,
				},
			)
		})
	})

	t.Run("ProcessorResolveStreamsReconciler", func(t *testing.T) {
		table := rtesting.SubTable{
			{
				Name: "skip",
				Parent: processor.
					StatusLatestImage(""),
				ExpectStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey:  nil,
					streaming.OutputStreamsStashKey: nil,
				},
			},
			{
				Name:   "no streams",
				Parent: processor,
				ExpectParent: processor.
					StatusConditions(
						processorConditionStreamsReady.True(),
					),
				ExpectStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey:  []streamingv1alpha1.Stream{},
					streaming.OutputStreamsStashKey: []streamingv1alpha1.Stream{},
				},
			},
			{
				Name: "resolve streams",
				Parent: processor.
					Inputs(
						testStream1.CreateInputStreamBinding("alias-1", streamingv1alpha1.Earliest),
						testStream2.CreateInputStreamBinding("alias-2", streamingv1alpha1.Earliest),
					).
					Outputs(
						testStream3.CreateOutputStreamBinding("alias-3"),
						testStream4.CreateOutputStreamBinding("alias-4"),
					),
				GivenObjects: []rtesting.Factory{
					testStream1,
					testStream2,
					testStream3,
					testStream4,
				},
				ExpectParent: processor.
					Inputs(
						testStream1.CreateInputStreamBinding("alias-1", streamingv1alpha1.Earliest),
						testStream2.CreateInputStreamBinding("alias-2", streamingv1alpha1.Earliest),
					).
					Outputs(
						testStream3.CreateOutputStreamBinding("alias-3"),
						testStream4.CreateOutputStreamBinding("alias-4"),
					).
					StatusConditions(
						processorConditionStreamsReady.True(),
					),
				ExpectTracks: []rtesting.TrackRequest{
					rtesting.NewTrackRequest(testStream1, processor, scheme),
					rtesting.NewTrackRequest(testStream2, processor, scheme),
					rtesting.NewTrackRequest(testStream3, processor, scheme),
					rtesting.NewTrackRequest(testStream4, processor, scheme),
				},
				ExpectStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey: []streamingv1alpha1.Stream{
						*testStream1.Create(),
						*testStream2.Create(),
					},
					streaming.OutputStreamsStashKey: []streamingv1alpha1.Stream{
						*testStream3.Create(),
						*testStream4.Create(),
					},
				},
			},
			{
				Name: "input stream not found",
				Parent: processor.
					Inputs(
						testStream1.CreateInputStreamBinding("alias-1", streamingv1alpha1.Earliest),
					),
				ShouldErr: true,
				ExpectParent: processor.
					Inputs(
						testStream1.CreateInputStreamBinding("alias-1", streamingv1alpha1.Earliest),
					),
				ExpectTracks: []rtesting.TrackRequest{
					rtesting.NewTrackRequest(testStream1, processor, scheme),
				},
				ExpectStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey:  nil,
					streaming.OutputStreamsStashKey: nil,
				},
			},
			{
				Name: "output stream not found",
				Parent: processor.
					Outputs(
						testStream1.CreateOutputStreamBinding("alias-1"),
					),
				ShouldErr: true,
				ExpectParent: processor.
					Outputs(
						testStream1.CreateOutputStreamBinding("alias-1"),
					),
				ExpectTracks: []rtesting.TrackRequest{
					rtesting.NewTrackRequest(testStream1, processor, scheme),
				},
				ExpectStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey:  []streamingv1alpha1.Stream{},
					streaming.OutputStreamsStashKey: nil,
				},
			},
			{
				Name: "stream not ready",
				Parent: processor.
					Inputs(
						testStream1.CreateInputStreamBinding("alias-1", streamingv1alpha1.Earliest),
					),
				GivenObjects: []rtesting.Factory{
					testStream1.
						StatusConditions(),
				},
				ExpectParent: processor.
					Inputs(
						testStream1.CreateInputStreamBinding("alias-1", streamingv1alpha1.Earliest),
					).
					StatusConditions(
						processorConditionReady.False().Reason("StreamNotReady", "stream stream-1 is not ready: stream has no ready condition"),
						processorConditionStreamsReady.False().Reason("StreamNotReady", "stream stream-1 is not ready: stream has no ready condition"),
					),
				ExpectTracks: []rtesting.TrackRequest{
					rtesting.NewTrackRequest(testStream1, processor, scheme),
				},
				ExpectStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey: []streamingv1alpha1.Stream{
						*testStream1.StatusConditions().Create(),
					},
					streaming.OutputStreamsStashKey: []streamingv1alpha1.Stream{},
				},
			},
		}

		table.Test(t, scheme, func(t *testing.T, row *rtesting.SubTestcase, client client.Client, tracker tracker.Tracker, recorder record.EventRecorder, log logr.Logger) controllers.SubReconciler {
			return streaming.ProcessorResolveStreamsReconciler(
				controllers.Config{
					Client:    client,
					APIReader: client,
					Recorder:  recorder,
					Log:       log,
					Scheme:    scheme,
					Tracker:   tracker,
				},
			)
		})
	})

	t.Run("ProcessorChildDeploymentReconciler", func(t *testing.T) {
		table := rtesting.SubTable{
			{
				Name:   "skip, missing latest image",
				Parent: processorMinimal,
				GivenStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey:    []streamingv1alpha1.Stream{},
					streaming.OutputStreamsStashKey:   []streamingv1alpha1.Stream{},
					streaming.ProcessorImagesStashKey: processorImagesConfigMap.Create().Data,
				},
			},
			{
				Name: "skip, missing input streams",
				Parent: processorMinimal.
					StatusLatestImage(testImage),
				GivenStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey:    nil,
					streaming.OutputStreamsStashKey:   []streamingv1alpha1.Stream{},
					streaming.ProcessorImagesStashKey: processorImagesConfigMap.Create().Data,
				},
			},
			{
				Name: "skip, missing output streams",
				Parent: processorMinimal.
					StatusLatestImage(testImage),
				GivenStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey:    []streamingv1alpha1.Stream{},
					streaming.OutputStreamsStashKey:   nil,
					streaming.ProcessorImagesStashKey: processorImagesConfigMap.Create().Data,
				},
			},
			{
				Name: "skip, missing processor images map",
				Parent: processorMinimal.
					StatusLatestImage(testImage),
				GivenStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey:    []streamingv1alpha1.Stream{},
					streaming.OutputStreamsStashKey:   []streamingv1alpha1.Stream{},
					streaming.ProcessorImagesStashKey: nil,
				},
			},
			{
				Name: "skip, missing processor images map, missing processor image",
				Parent: processorMinimal.
					StatusLatestImage(testImage),
				GivenStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey:  []streamingv1alpha1.Stream{},
					streaming.OutputStreamsStashKey: []streamingv1alpha1.Stream{},
					streaming.ProcessorImagesStashKey: map[string]string{
						processorImageKey: "",
					},
				},
			},
			{
				Name: "create deployment",
				Parent: processor.
					Inputs(
						testStream1.CreateInputStreamBinding("alias-in-1", streamingv1alpha1.Earliest),
						testStream2.CreateInputStreamBinding("alias-in-2", streamingv1alpha1.Latest),
					).
					Outputs(
						testStream3.CreateOutputStreamBinding("alias-out-2"),
						testStream4.CreateOutputStreamBinding("alias-out-4"),
					),
				GivenStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey: []streamingv1alpha1.Stream{
						*testStream1.Create(),
						*testStream2.Create(),
					},
					streaming.OutputStreamsStashKey: []streamingv1alpha1.Stream{
						*testStream3.Create(),
						*testStream4.Create(),
					},
					streaming.ProcessorImagesStashKey: processorImagesConfigMap.Create().Data,
				},
				ExpectParent: processor.
					Inputs(
						testStream1.CreateInputStreamBinding("alias-in-1", streamingv1alpha1.Earliest),
						testStream2.CreateInputStreamBinding("alias-in-2", streamingv1alpha1.Latest),
					).
					Outputs(
						testStream3.CreateOutputStreamBinding("alias-out-2"),
						testStream4.CreateOutputStreamBinding("alias-out-4"),
					).
					StatusDeploymentRef("%s-processor-001", testName),
				ExpectEvents: []rtesting.Event{
					rtesting.NewEvent(processor, scheme, corev1.EventTypeNormal, "Created",
						`Created Deployment "%s-processor-001"`, testName),
				},
				ExpectCreates: []rtesting.Factory{
					deploymentCreate.
						PodTemplateSpec(func(pts factories.PodTemplateSpec) {
							pts.ContainerNamed("processor", func(c *corev1.Container) {
								c.Env = []corev1.EnvVar{
									{Name: "CNB_BINDINGS", Value: "/var/riff/bindings"},
									{Name: "INPUT_START_OFFSETS", Value: "earliest,latest"},
									{Name: "INPUT_NAMES", Value: "alias-in-1,alias-in-2"},
									{Name: "OUTPUT_NAMES", Value: "alias-out-2,alias-out-4"},
									{Name: "GROUP", Value: "test-processor"},
									{Name: "FUNCTION", Value: "localhost:8081"},
								}
								c.VolumeMounts = []corev1.VolumeMount{
									{
										Name:      "stream-00000000-0000-0000-0000-000000000001-metadata",
										ReadOnly:  true,
										MountPath: "/var/riff/bindings/input_000/metadata",
									},
									{
										Name:      "stream-00000000-0000-0000-0000-000000000001-secret",
										ReadOnly:  true,
										MountPath: "/var/riff/bindings/input_000/secret",
									},
									{
										Name:      "stream-00000000-0000-0000-0000-000000000002-metadata",
										ReadOnly:  true,
										MountPath: "/var/riff/bindings/input_001/metadata",
									},
									{
										Name:      "stream-00000000-0000-0000-0000-000000000002-secret",
										ReadOnly:  true,
										MountPath: "/var/riff/bindings/input_001/secret",
									},
									{
										Name:      "stream-00000000-0000-0000-0000-000000000003-metadata",
										ReadOnly:  true,
										MountPath: "/var/riff/bindings/output_000/metadata",
									},
									{
										Name:      "stream-00000000-0000-0000-0000-000000000003-secret",
										ReadOnly:  true,
										MountPath: "/var/riff/bindings/output_000/secret",
									},
									{
										Name:      "stream-00000000-0000-0000-0000-000000000004-metadata",
										ReadOnly:  true,
										MountPath: "/var/riff/bindings/output_001/metadata",
									},
									{
										Name:      "stream-00000000-0000-0000-0000-000000000004-secret",
										ReadOnly:  true,
										MountPath: "/var/riff/bindings/output_001/secret",
									},
								}
							})
							pts.Volumes(
								corev1.Volume{
									Name: "stream-00000000-0000-0000-0000-000000000001-metadata",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "stream-1-binding-metadata",
											},
										},
									},
								},
								corev1.Volume{
									Name: "stream-00000000-0000-0000-0000-000000000001-secret",
									VolumeSource: corev1.VolumeSource{
										Secret: &corev1.SecretVolumeSource{
											SecretName: "stream-1-binding-secret",
										},
									},
								},
								corev1.Volume{
									Name: "stream-00000000-0000-0000-0000-000000000002-metadata",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "stream-2-binding-metadata",
											},
										},
									},
								},
								corev1.Volume{
									Name: "stream-00000000-0000-0000-0000-000000000002-secret",
									VolumeSource: corev1.VolumeSource{
										Secret: &corev1.SecretVolumeSource{
											SecretName: "stream-2-binding-secret",
										},
									},
								},
								corev1.Volume{
									Name: "stream-00000000-0000-0000-0000-000000000003-metadata",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "stream-3-binding-metadata",
											},
										},
									},
								},
								corev1.Volume{
									Name: "stream-00000000-0000-0000-0000-000000000003-secret",
									VolumeSource: corev1.VolumeSource{
										Secret: &corev1.SecretVolumeSource{
											SecretName: "stream-3-binding-secret",
										},
									},
								},
								corev1.Volume{
									Name: "stream-00000000-0000-0000-0000-000000000004-metadata",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "stream-4-binding-metadata",
											},
										},
									},
								},
								corev1.Volume{
									Name: "stream-00000000-0000-0000-0000-000000000004-secret",
									VolumeSource: corev1.VolumeSource{
										Secret: &corev1.SecretVolumeSource{
											SecretName: "stream-4-binding-secret",
										},
									},
								},
							)
						}),
				},
			}, {
				Name: "update deployment",
				Parent: processor.
					Inputs(
						testStream1.CreateInputStreamBinding("alias-in-1", streamingv1alpha1.Earliest),
					),
				GivenStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey: []streamingv1alpha1.Stream{
						*testStream1.Create(),
					},
					streaming.OutputStreamsStashKey:   []streamingv1alpha1.Stream{},
					streaming.ProcessorImagesStashKey: processorImagesConfigMap.Create().Data,
				},
				GivenObjects: []rtesting.Factory{
					deploymentGiven,
				},
				ExpectParent: processor.
					Inputs(
						testStream1.CreateInputStreamBinding("alias-in-1", streamingv1alpha1.Earliest),
					).
					Outputs().
					StatusDeploymentRef("%s-processor-000", testName),
				ExpectEvents: []rtesting.Event{
					rtesting.NewEvent(processor, scheme, corev1.EventTypeNormal, "Updated",
						`Updated Deployment "%s-processor-000"`, testName),
				},
				ExpectUpdates: []rtesting.Factory{
					deploymentGiven.
						PodTemplateSpec(func(pts factories.PodTemplateSpec) {
							pts.ContainerNamed("processor", func(c *corev1.Container) {
								c.Env = []corev1.EnvVar{
									{Name: "CNB_BINDINGS", Value: "/var/riff/bindings"},
									{Name: "INPUT_START_OFFSETS", Value: "earliest"},
									{Name: "INPUT_NAMES", Value: "alias-in-1"},
									{Name: "OUTPUT_NAMES", Value: ""},
									{Name: "GROUP", Value: "test-processor"},
									{Name: "FUNCTION", Value: "localhost:8081"},
								}
								c.VolumeMounts = []corev1.VolumeMount{
									{
										Name:      "stream-00000000-0000-0000-0000-000000000001-metadata",
										ReadOnly:  true,
										MountPath: "/var/riff/bindings/input_000/metadata",
									},
									{
										Name:      "stream-00000000-0000-0000-0000-000000000001-secret",
										ReadOnly:  true,
										MountPath: "/var/riff/bindings/input_000/secret",
									},
								}
							})
							pts.Volumes(
								corev1.Volume{
									Name: "stream-00000000-0000-0000-0000-000000000001-metadata",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "stream-1-binding-metadata",
											},
										},
									},
								},
								corev1.Volume{
									Name: "stream-00000000-0000-0000-0000-000000000001-secret",
									VolumeSource: corev1.VolumeSource{
										Secret: &corev1.SecretVolumeSource{
											SecretName: "stream-1-binding-secret",
										},
									},
								},
							)
						}),
				},
			},
		}

		table.Test(t, scheme, func(t *testing.T, row *rtesting.SubTestcase, client client.Client, tracker tracker.Tracker, recorder record.EventRecorder, log logr.Logger) controllers.SubReconciler {
			return streaming.ProcessorChildDeploymentReconciler(
				controllers.Config{
					Client:    client,
					APIReader: client,
					Recorder:  recorder,
					Log:       log,
					Scheme:    scheme,
					Tracker:   tracker,
				},
			)
		})
	})

	t.Run("ProcessorChildScaledObjectReconciler", func(t *testing.T) {
		table := rtesting.SubTable{
			{
				Name: "skip, missing deployment",
				Parent: processorMinimal.
					Inputs(
						testStream1.CreateInputStreamBinding("alias-in-1", streamingv1alpha1.Earliest),
						testStream1.CreateInputStreamBinding("alias-in-2", streamingv1alpha1.Latest),
					),
				GivenStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey: []streamingv1alpha1.Stream{
						*testStream1.Create(),
						*testStream2.Create(),
					},
				},
			},
			{
				Name: "skip, missing inputs",
				Parent: processorMinimal.
					Inputs(
						testStream1.CreateInputStreamBinding("alias-in-1", streamingv1alpha1.Earliest),
						testStream1.CreateInputStreamBinding("alias-in-2", streamingv1alpha1.Latest),
					).
					StatusDeploymentRef("%s-processor-000", testName),
				GivenStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey: nil,
				},
			},
			{
				Name: "create scaled object",
				Parent: processor.
					Inputs(
						testStream1.CreateInputStreamBinding("alias-in-1", streamingv1alpha1.Earliest),
						testStream1.CreateInputStreamBinding("alias-in-2", streamingv1alpha1.Latest),
					),
				GivenStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey: []streamingv1alpha1.Stream{
						*testStream1.Create(),
						*testStream2.Create(),
					},
				},
				GivenObjects: []rtesting.Factory{
					factories.Secret().
						NamespaceName(testNamespace, "stream-1-binding-secret").
						AddData("gateway", "stream-1-gateway.local:6565").
						AddData("topic", fmt.Sprintf("%s/%s", testNamespace, "stream-1")),
					factories.Secret().
						NamespaceName(testNamespace, "stream-2-binding-secret").
						AddData("gateway", "stream-2-gateway.local:6565").
						AddData("topic", fmt.Sprintf("%s/%s", testNamespace, "stream-2")),
				},
				ExpectParent: processor.
					Inputs(
						testStream1.CreateInputStreamBinding("alias-in-1", streamingv1alpha1.Earliest),
						testStream1.CreateInputStreamBinding("alias-in-2", streamingv1alpha1.Latest),
					).
					StatusScaledObjectRef("%s-processor-001", testName).
					StatusConditions(
						processorConditionScaledObjectReady.True(),
					),
				ExpectEvents: []rtesting.Event{
					rtesting.NewEvent(processor, scheme, corev1.EventTypeNormal, "Created",
						`Created ScaledObject "%s-processor-001"`, testName),
				},
				ExpectCreates: []rtesting.Factory{
					scaledObjectCreate.
						Triggers(
							kedav1alpha1.ScaleTriggers{
								Type: "liiklus",
								Metadata: map[string]string{
									"address": "stream-1-gateway.local:6565",
									"group":   testName,
									"topic":   fmt.Sprintf("%s/stream-1", testNamespace),
								},
							},
							kedav1alpha1.ScaleTriggers{
								Type: "liiklus",
								Metadata: map[string]string{
									"address": "stream-2-gateway.local:6565",
									"group":   testName,
									"topic":   fmt.Sprintf("%s/stream-2", testNamespace),
								},
							},
						),
				},
			},
			{
				Name: "update scaled object",
				Parent: processor.
					Inputs(
						testStream1.CreateInputStreamBinding("alias-in-1", streamingv1alpha1.Earliest),
						testStream1.CreateInputStreamBinding("alias-in-2", streamingv1alpha1.Latest),
					),
				GivenStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey: []streamingv1alpha1.Stream{
						*testStream1.Create(),
						*testStream2.Create(),
					},
				},
				GivenObjects: []rtesting.Factory{
					scaledObjectGiven,
					factories.Secret().
						NamespaceName(testNamespace, "stream-1-binding-secret").
						AddData("gateway", "stream-1-gateway.local:6565").
						AddData("topic", fmt.Sprintf("%s/%s", testNamespace, "stream-1")),
					factories.Secret().
						NamespaceName(testNamespace, "stream-2-binding-secret").
						AddData("gateway", "stream-2-gateway.local:6565").
						AddData("topic", fmt.Sprintf("%s/%s", testNamespace, "stream-2")),
				},
				ExpectParent: processor.
					Inputs(
						testStream1.CreateInputStreamBinding("alias-in-1", streamingv1alpha1.Earliest),
						testStream1.CreateInputStreamBinding("alias-in-2", streamingv1alpha1.Latest),
					).
					StatusConditions(
						processorConditionScaledObjectReady.True(),
					),
				ExpectEvents: []rtesting.Event{
					rtesting.NewEvent(processor, scheme, corev1.EventTypeNormal, "Updated",
						`Updated ScaledObject "%s-processor-000"`, testName),
				},
				ExpectUpdates: []rtesting.Factory{
					scaledObjectGiven.
						Triggers(
							kedav1alpha1.ScaleTriggers{
								Type: "liiklus",
								Metadata: map[string]string{
									"address": "stream-1-gateway.local:6565",
									"group":   testName,
									"topic":   fmt.Sprintf("%s/stream-1", testNamespace),
								},
							},
							kedav1alpha1.ScaleTriggers{
								Type: "liiklus",
								Metadata: map[string]string{
									"address": "stream-2-gateway.local:6565",
									"group":   testName,
									"topic":   fmt.Sprintf("%s/stream-2", testNamespace),
								},
							},
						),
				},
			},
			{
				Name: "scale to zero for not ready streams",
				Parent: processor.
					Inputs(
						testStream1.CreateInputStreamBinding("alias-in-1", streamingv1alpha1.Earliest),
					).
					StatusConditions(
						processorConditionStreamsReady.False(),
					),
				GivenStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey: []streamingv1alpha1.Stream{
						*testStream1.Create(),
					},
				},
				GivenObjects: []rtesting.Factory{
					scaledObjectGiven,
					factories.Secret().
						NamespaceName(testNamespace, "stream-1-binding-secret").
						AddData("gateway", "stream-1-gateway.local:6565").
						AddData("topic", fmt.Sprintf("%s/%s", testNamespace, "stream-1")),
				},
				ExpectParent: processor.
					Inputs(
						testStream1.CreateInputStreamBinding("alias-in-1", streamingv1alpha1.Earliest),
					).
					StatusConditions(
						processorConditionScaledObjectReady.True(),
						processorConditionStreamsReady.False(),
					),
				ExpectEvents: []rtesting.Event{
					rtesting.NewEvent(processor, scheme, corev1.EventTypeNormal, "Updated",
						`Updated ScaledObject "%s-processor-000"`, testName),
				},
				ExpectUpdates: []rtesting.Factory{
					scaledObjectGiven.
						MaxReplicaCount(0).
						Triggers(
							kedav1alpha1.ScaleTriggers{
								Type: "liiklus",
								Metadata: map[string]string{
									"address": "stream-1-gateway.local:6565",
									"group":   testName,
									"topic":   fmt.Sprintf("%s/stream-1", testNamespace),
								},
							},
						),
				},
			},
			{
				Name: "binding secret not found",
				Parent: processorMinimal.
					StatusLatestImage(testImage).
					StatusDeploymentRef("%s-processor-000", testName).
					Inputs(
						testStream1.CreateInputStreamBinding("alias-in-1", streamingv1alpha1.Earliest),
					),
				GivenStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey: []streamingv1alpha1.Stream{
						*testStream1.Create(),
					},
				},
			},
			{
				Name: "binding secret error",
				Parent: processorMinimal.
					StatusLatestImage(testImage).
					StatusDeploymentRef("%s-processor-000", testName).
					Inputs(
						testStream1.CreateInputStreamBinding("alias-in-1", streamingv1alpha1.Earliest),
					),
				WithReactors: []rtesting.ReactionFunc{
					rtesting.InduceFailure("get", "Secret"),
				},
				GivenStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey: []streamingv1alpha1.Stream{
						*testStream1.Create(),
					},
				},
				ShouldErr: true,
			},
			{
				Name: "binding secret missing gateway",
				Parent: processorMinimal.
					StatusLatestImage(testImage).
					StatusDeploymentRef("%s-processor-000", testName).
					Inputs(
						testStream1.CreateInputStreamBinding("alias-in-1", streamingv1alpha1.Earliest),
					),
				GivenStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey: []streamingv1alpha1.Stream{
						*testStream1.Create(),
					},
				},
				GivenObjects: []rtesting.Factory{
					factories.Secret().
						NamespaceName(testNamespace, "stream-1-binding-secret").
						AddData("topic", fmt.Sprintf("%s/%s", testNamespace, "stream-1")),
				},
			},
			{
				Name: "binding secret missing topic",
				Parent: processorMinimal.
					StatusLatestImage(testImage).
					StatusDeploymentRef("%s-processor-000", testName).
					Inputs(
						testStream1.CreateInputStreamBinding("alias-in-1", streamingv1alpha1.Earliest),
					),
				GivenStashedValues: map[controllers.StashKey]interface{}{
					streaming.InputStreamsStashKey: []streamingv1alpha1.Stream{
						*testStream1.Create(),
					},
				},
				GivenObjects: []rtesting.Factory{
					factories.Secret().
						NamespaceName(testNamespace, "stream-1-binding-secret").
						AddData("gateway", "stream-2-gateway.local:6565"),
				},
			},
		}

		table.Test(t, scheme, func(t *testing.T, row *rtesting.SubTestcase, client client.Client, tracker tracker.Tracker, recorder record.EventRecorder, log logr.Logger) controllers.SubReconciler {
			return streaming.ProcessorChildScaledObjectReconciler(
				controllers.Config{
					Client:    client,
					APIReader: client,
					Recorder:  recorder,
					Log:       log,
					Scheme:    scheme,
					Tracker:   tracker,
				},
			)
		})
	})
}
