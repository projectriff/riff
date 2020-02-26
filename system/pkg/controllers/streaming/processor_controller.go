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

package streaming

import (
	"context"
	"fmt"
	"sort"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/projectriff/riff/system/pkg/apis"
	buildv1alpha1 "github.com/projectriff/riff/system/pkg/apis/build/v1alpha1"
	streamingv1alpha1 "github.com/projectriff/riff/system/pkg/apis/streaming/v1alpha1"
	kedav1alpha1 "github.com/projectriff/riff/system/pkg/apis/thirdparty/keda/v1alpha1"
	"github.com/projectriff/riff/system/pkg/controllers"
	"github.com/projectriff/riff/system/pkg/refs"
	"github.com/projectriff/riff/system/pkg/tracker"
)

const (
	bindingsRootPath = "/var/riff/bindings"
)

const (
	ProcessorImagesStashKey controllers.StashKey = "processor-images"
	InputStreamsStashKey    controllers.StashKey = "input-streams"
	OutputStreamsStashKey   controllers.StashKey = "output-streams"
)

// +kubebuilder:rbac:groups=streaming.projectriff.io,resources=processors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=streaming.projectriff.io,resources=processors/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=keda.k8s.io,resources=scaledobjects,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=streaming.projectriff.io,resources=streams,verbs=get;watch
// +kubebuilder:rbac:groups=build.projectriff.io,resources=containers,verbs=get;watch
// +kubebuilder:rbac:groups=build.projectriff.io,resources=functions,verbs=get;watch
// +kubebuilder:rbac:groups=core,resources=configmaps;secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete

func ProcessorReconciler(c controllers.Config, namespace string) *controllers.ParentReconciler {
	c.Log = c.Log.WithName("Processor")

	return &controllers.ParentReconciler{
		Type: &streamingv1alpha1.Processor{},
		SubReconcilers: []controllers.SubReconciler{
			ProcessorSyncProcessorImages(c, namespace),
			ProcessorBuildRefReconciler(c),
			ProcessorResolveStreamsReconciler(c),
			ProcessorChildDeploymentReconciler(c),
			ProcessorChildScaledObjectReconciler(c),
		},

		Config: c,
	}
}

func ProcessorSyncProcessorImages(c controllers.Config, namespace string) controllers.SubReconciler {
	c.Log = c.Log.WithName("ProcessorImages")

	return &controllers.SyncReconciler{
		Sync: func(ctx context.Context, processor *streamingv1alpha1.Processor) error {
			config := corev1.ConfigMap{}
			key := types.NamespacedName{Namespace: namespace, Name: processorImages}
			// track config for new images
			c.Tracker.Track(
				tracker.NewKey(schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}, key),
				types.NamespacedName{Namespace: processor.Namespace, Name: processor.Name},
			)
			if err := c.Get(ctx, key, &config); err != nil {
				return err
			}
			controllers.StashValue(ctx, ProcessorImagesStashKey, config.Data)
			return nil
		},

		Config: c,
		Setup: func(mgr controllers.Manager, bldr *controllers.Builder) error {
			bldr.Watches(&source.Kind{Type: &corev1.ConfigMap{}}, controllers.EnqueueTracked(&corev1.ConfigMap{}, c.Tracker, c.Scheme))
			return nil
		},
	}
}

func ProcessorBuildRefReconciler(c controllers.Config) controllers.SubReconciler {
	c.Log = c.Log.WithName("BuildRef")

	return &controllers.SyncReconciler{
		Sync: func(ctx context.Context, parent *streamingv1alpha1.Processor) error {
			build := parent.Spec.Build
			if build == nil {
				parent.Status.LatestImage = parent.Spec.Template.Spec.Containers[0].Image
				return nil
			}

			switch {

			case build.ContainerRef != "":
				var container buildv1alpha1.Container
				key := types.NamespacedName{Namespace: parent.Namespace, Name: build.ContainerRef}
				// track container for new images
				c.Tracker.Track(
					tracker.NewKey(container.GetGroupVersionKind(), key),
					types.NamespacedName{Namespace: parent.Namespace, Name: parent.Name},
				)
				if err := c.Get(ctx, key, &container); err != nil {
					if apierrs.IsNotFound(err) {
						return nil
					}
					return err
				}
				if container.Status.LatestImage != "" {
					parent.Status.LatestImage = container.Status.LatestImage
				}
				return nil

			case build.FunctionRef != "":
				var function buildv1alpha1.Function
				key := types.NamespacedName{Namespace: parent.Namespace, Name: build.FunctionRef}
				// track function for new images
				c.Tracker.Track(
					tracker.NewKey(function.GetGroupVersionKind(), key),
					types.NamespacedName{Namespace: parent.Namespace, Name: parent.Name},
				)
				if err := c.Get(ctx, key, &function); err != nil {
					if apierrs.IsNotFound(err) {
						return nil
					}
					return err
				}
				if function.Status.LatestImage != "" {
					parent.Status.LatestImage = function.Status.LatestImage
				}
				return nil

			}

			panic(fmt.Errorf("invalid processor build"))
		},

		Config: c,
		Setup: func(mgr controllers.Manager, bldr *controllers.Builder) error {
			bldr.Watches(&source.Kind{Type: &buildv1alpha1.Container{}}, controllers.EnqueueTracked(&buildv1alpha1.Container{}, c.Tracker, c.Scheme))
			bldr.Watches(&source.Kind{Type: &buildv1alpha1.Function{}}, controllers.EnqueueTracked(&buildv1alpha1.Function{}, c.Tracker, c.Scheme))
			return nil
		},
	}
}

func ProcessorResolveStreamsReconciler(c controllers.Config) controllers.SubReconciler {
	c.Log = c.Log.WithName("ResolveStreams")

	resolveStream := func(ctx context.Context, streamKey, processorKey types.NamespacedName) (*streamingv1alpha1.Stream, error) {
		var stream streamingv1alpha1.Stream
		// track stream for new coordinates
		c.Tracker.Track(
			tracker.NewKey(stream.GetGroupVersionKind(), streamKey),
			processorKey,
		)
		if err := c.Client.Get(ctx, streamKey, &stream); err != nil {
			return nil, err
		}
		return &stream, nil
	}

	return &controllers.SyncReconciler{
		Sync: func(ctx context.Context, processor *streamingv1alpha1.Processor) error {
			if processor.Status.LatestImage == "" {
				return nil
			}

			processorKey := types.NamespacedName{Namespace: processor.Namespace, Name: processor.Name}

			inputStreams := make([]streamingv1alpha1.Stream, len(processor.Spec.Inputs))
			for i, binding := range processor.Spec.Inputs {
				key := types.NamespacedName{Namespace: processor.Namespace, Name: binding.Stream}
				stream, err := resolveStream(ctx, key, processorKey)
				if err != nil {
					return err
				}
				inputStreams[i] = *stream
			}
			controllers.StashValue(ctx, InputStreamsStashKey, inputStreams)

			outputStreams := make([]streamingv1alpha1.Stream, len(processor.Spec.Outputs))
			for i, binding := range processor.Spec.Outputs {
				key := types.NamespacedName{Namespace: processor.Namespace, Name: binding.Stream}
				stream, err := resolveStream(ctx, key, processorKey)
				if err != nil {
					return err
				}
				outputStreams[i] = *stream
			}
			controllers.StashValue(ctx, OutputStreamsStashKey, outputStreams)

			streams := []streamingv1alpha1.Stream{}
			streams = append(streams, inputStreams...)
			streams = append(streams, outputStreams...)
			processor.Status.MarkStreamsReady()
			for _, stream := range streams {
				ready := stream.Status.GetCondition(stream.Status.GetReadyConditionType())
				if ready == nil {
					ready = &apis.Condition{Message: "stream has no ready condition"}
				}
				if !ready.IsTrue() {
					processor.Status.MarkStreamsNotReady(fmt.Sprintf("stream %s is not ready: %s", stream.Name, ready.Message))
					break
				}
			}

			return nil
		},

		Config: c,
		Setup: func(mgr controllers.Manager, bldr *controllers.Builder) error {
			bldr.Watches(&source.Kind{Type: &streamingv1alpha1.Stream{}}, controllers.EnqueueTracked(&streamingv1alpha1.Stream{}, c.Tracker, c.Scheme))
			return nil
		},
	}
}

func ProcessorChildDeploymentReconciler(c controllers.Config) controllers.SubReconciler {
	c.Log = c.Log.WithName("ChildDeployment")

	one := int32(1)

	constructVolumes := func(processor *streamingv1alpha1.Processor, inputStreams, outputStreams []streamingv1alpha1.Stream) ([]corev1.Volume, []corev1.VolumeMount) {
		volumes := []corev1.Volume{}
		volumeMounts := []corev1.VolumeMount{}

		// De-dupe streams and create one volume for each
		streams := make(map[string]streamingv1alpha1.Stream)
		for _, s := range inputStreams {
			streams[s.Name] = s
		}
		for _, s := range outputStreams {
			streams[s.Name] = s
		}
		for _, stream := range streams {
			if stream.Status.Binding.MetadataRef.Name != "" {
				volumes = append(volumes,
					corev1.Volume{
						Name: fmt.Sprintf("stream-%s-metadata", stream.UID),
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: stream.Status.Binding.MetadataRef.Name,
								},
							},
						},
					},
				)
			}
			if stream.Status.Binding.SecretRef.Name != "" {
				volumes = append(volumes,
					corev1.Volume{
						Name: fmt.Sprintf("stream-%s-secret", stream.UID),
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: stream.Status.Binding.SecretRef.Name,
							},
						},
					},
				)
			}
		}

		// Create one volume mount for each *binding*, split into inputs/outputs.
		// The consumer of those will know to count from 0..Nbindings-1 thanks to the INPUT/OUTPUT_NAMES var
		for i, binding := range processor.Spec.Inputs {
			stream := streams[binding.Stream]
			if stream.Status.Binding.MetadataRef.Name != "" {
				volumeMounts = append(volumeMounts,
					corev1.VolumeMount{
						Name:      fmt.Sprintf("stream-%s-metadata", stream.UID),
						MountPath: fmt.Sprintf("%s/input_%03d/metadata", bindingsRootPath, i),
						ReadOnly:  true,
					},
				)
			}
			if stream.Status.Binding.SecretRef.Name != "" {
				volumeMounts = append(volumeMounts,
					corev1.VolumeMount{
						Name:      fmt.Sprintf("stream-%s-secret", stream.UID),
						MountPath: fmt.Sprintf("%s/input_%03d/secret", bindingsRootPath, i),
						ReadOnly:  true,
					},
				)
			}
		}
		for i, binding := range processor.Spec.Outputs {
			stream := streams[binding.Stream]
			if stream.Status.Binding.MetadataRef.Name != "" {
				volumeMounts = append(volumeMounts,
					corev1.VolumeMount{
						Name:      fmt.Sprintf("stream-%s-metadata", stream.UID),
						MountPath: fmt.Sprintf("%s/output_%03d/metadata", bindingsRootPath, i),
						ReadOnly:  true,
					},
				)
			}
			if stream.Status.Binding.SecretRef.Name != "" {
				volumeMounts = append(volumeMounts,
					corev1.VolumeMount{
						Name:      fmt.Sprintf("stream-%s-secret", stream.UID),
						MountPath: fmt.Sprintf("%s/output_%03d/secret", bindingsRootPath, i),
						ReadOnly:  true,
					},
				)
			}
		}

		// sort volumes to avoid update diffs caused by iteration order
		sort.SliceStable(volumes, func(i, j int) bool {
			return volumes[i].Name < volumes[j].Name
		})

		return volumes, volumeMounts
	}

	constructEnv := func(processor *streamingv1alpha1.Processor) []v1.EnvVar {
		inputStartOffsets := make([]string, len(processor.Spec.Inputs))
		for i, binding := range processor.Spec.Inputs {
			inputStartOffsets[i] = binding.StartOffset
		}
		inputAliases := make([]string, len(processor.Spec.Inputs))
		for i, binding := range processor.Spec.Inputs {
			inputAliases[i] = binding.Alias
		}
		outputAliases := make([]string, len(processor.Spec.Outputs))
		for i, binding := range processor.Spec.Outputs {
			outputAliases[i] = binding.Alias
		}

		return []v1.EnvVar{
			{
				Name:  "CNB_BINDINGS",
				Value: bindingsRootPath,
			},
			{
				Name:  "INPUT_START_OFFSETS",
				Value: strings.Join(inputStartOffsets, ","),
			},
			{
				Name:  "INPUT_NAMES",
				Value: strings.Join(inputAliases, ","),
			},
			{
				Name:  "OUTPUT_NAMES",
				Value: strings.Join(outputAliases, ","),
			},
			{
				Name:  "GROUP",
				Value: processor.Name,
			},
			{
				Name:  "FUNCTION",
				Value: "localhost:8081",
			},
		}
	}

	return &controllers.ChildReconciler{
		ParentType:    &streamingv1alpha1.Processor{},
		ChildType:     &appsv1.Deployment{},
		ChildListType: &appsv1.DeploymentList{},

		DesiredChild: func(ctx context.Context, parent *streamingv1alpha1.Processor) (*appsv1.Deployment, error) {
			if parent.Status.LatestImage == "" {
				// no image, skip
				return nil, nil
			}
			inputStreams, ok := controllers.RetrieveValue(ctx, InputStreamsStashKey).([]streamingv1alpha1.Stream)
			if !ok {
				return nil, nil
			}
			outputStreams, ok := controllers.RetrieveValue(ctx, OutputStreamsStashKey).([]streamingv1alpha1.Stream)
			if !ok {
				return nil, nil
			}
			processorImages, ok := controllers.RetrieveValue(ctx, ProcessorImagesStashKey).(map[string]string)
			if !ok {
				return nil, nil
			}
			processorImage := processorImages[processorImageKey]
			if processorImage == "" {
				return nil, nil
			}

			labels := controllers.MergeMaps(parent.Labels, map[string]string{
				streamingv1alpha1.ProcessorLabelKey: parent.Name,
			})
			volumes, volumeMounts := constructVolumes(parent, inputStreams, outputStreams)
			env := constructEnv(parent)

			// merge provided template with controlled values
			template := parent.Spec.Template.DeepCopy()
			template.Labels = controllers.MergeMaps(template.Labels, labels)
			template.Spec.Containers[0].Image = parent.Status.LatestImage
			template.Spec.Containers[0].Ports = []v1.ContainerPort{
				{
					ContainerPort: 8081,
				},
			}
			template.Spec.Containers = append(template.Spec.Containers, v1.Container{
				Name:         "processor",
				Image:        processorImage,
				Env:          env,
				VolumeMounts: volumeMounts,
			})
			template.Spec.Volumes = append(template.Spec.Volumes, volumes...)

			child := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: fmt.Sprintf("%s-processor-", parent.Name),
					Namespace:    parent.Namespace,
					Labels:       labels,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &one,
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							streamingv1alpha1.ProcessorLabelKey: parent.Name,
						},
					},
					Template: *template,
				},
			}

			return child, nil
		},
		ReflectChildStatusOnParent: func(parent *streamingv1alpha1.Processor, child *appsv1.Deployment, err error) {
			if child == nil {
				parent.Status.DeploymentRef = nil
			} else {
				parent.Status.DeploymentRef = refs.NewTypedLocalObjectReferenceForObject(child, c.Scheme)
				parent.Status.PropagateDeploymentStatus(&child.Status)
			}
		},
		HarmonizeImmutableFields: func(current, desired *appsv1.Deployment) {
			desired.Spec.Replicas = current.Spec.Replicas
		},
		MergeBeforeUpdate: func(current, desired *appsv1.Deployment) {
			current.Labels = desired.Labels
			current.Spec = desired.Spec
		},
		SemanticEquals: func(a1, a2 *appsv1.Deployment) bool {
			return equality.Semantic.DeepEqual(a1.Spec, a2.Spec) &&
				equality.Semantic.DeepEqual(a1.Labels, a2.Labels)
		},

		Config:     c,
		IndexField: ".metadata.processorDeploymentController",
		Sanitize: func(child *appsv1.Deployment) interface{} {
			return child.Spec
		},
	}
}

func ProcessorChildScaledObjectReconciler(c controllers.Config) controllers.SubReconciler {
	c.Log = c.Log.WithName("ChildScaledObject")

	zero := int32(0)
	one := int32(1)
	thirty := int32(30)

	collectStreamAddresses := func(ctx context.Context, streams []streamingv1alpha1.Stream) ([][]string, error) {
		addresses := make([][]string, len(streams))
		for i, stream := range streams {
			var secret corev1.Secret
			if err := c.Get(ctx, types.NamespacedName{Namespace: stream.Namespace, Name: stream.Status.Binding.SecretRef.Name}, &secret); err != nil {
				if apierrs.IsNotFound(err) {
					return nil, nil
				}
				c.Log.Error(err, "failed to get binding secret", "stream", stream.Name, "secret", stream.Status.Binding.SecretRef.Name)
				return nil, err
			}
			gateway, ok := secret.Data["gateway"]
			if !ok {
				c.Log.Info("invalid binding: binding missing data 'gateway'", "secret", secret.Name)
				return nil, nil
			}
			topic, ok := secret.Data["topic"]
			if !ok {
				c.Log.Info("invalid binding: binding missing data 'topic'", "secret", secret.Name)
				return nil, nil
			}
			addresses[i] = []string{string(gateway), string(topic)}
		}
		return addresses, nil
	}

	return &controllers.ChildReconciler{
		ParentType:    &streamingv1alpha1.Processor{},
		ChildType:     &kedav1alpha1.ScaledObject{},
		ChildListType: &kedav1alpha1.ScaledObjectList{},

		DesiredChild: func(ctx context.Context, parent *streamingv1alpha1.Processor) (*kedav1alpha1.ScaledObject, error) {
			if parent.Status.DeploymentRef == nil {
				// no deployment, skip
				return nil, nil
			}
			inputStreams, ok := controllers.RetrieveValue(ctx, InputStreamsStashKey).([]streamingv1alpha1.Stream)
			if !ok {
				return nil, nil
			}

			labels := controllers.MergeMaps(parent.Labels, map[string]string{
				streamingv1alpha1.ProcessorLabelKey: parent.Name,
			})

			maxReplicas := thirty
			if parent.Status.GetCondition(streamingv1alpha1.ProcessorConditionStreamsReady).IsFalse() {
				// scale to zero while dependencies are not ready
				maxReplicas = zero
			}

			inputAddresses, err := collectStreamAddresses(ctx, inputStreams)
			if err != nil || inputAddresses == nil {
				return nil, err
			}
			triggers := make([]kedav1alpha1.ScaleTriggers, len(inputAddresses))
			for i, input := range inputAddresses {
				triggers[i].Type = "liiklus"
				triggers[i].Metadata = map[string]string{
					"address": input[0],
					"group":   parent.Name,
					"topic":   input[1],
				}
			}

			child := &kedav1alpha1.ScaledObject{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: fmt.Sprintf("%s-processor-", parent.Name),
					Namespace:    parent.Namespace,
					Labels:       labels,
				},
				Spec: kedav1alpha1.ScaledObjectSpec{
					ScaleTargetRef: &kedav1alpha1.ObjectReference{
						DeploymentName: parent.Status.DeploymentRef.Name,
					},
					PollingInterval: &one,
					CooldownPeriod:  &thirty,
					Triggers:        triggers,
					MinReplicaCount: &one,
					MaxReplicaCount: &maxReplicas,
				},
			}

			return child, nil
		},
		ReflectChildStatusOnParent: func(parent *streamingv1alpha1.Processor, child *kedav1alpha1.ScaledObject, err error) {
			if child == nil {
				parent.Status.ScaledObjectRef = nil
			} else {
				parent.Status.ScaledObjectRef = refs.NewTypedLocalObjectReferenceForObject(child, c.Scheme)
				parent.Status.PropagateScaledObjectStatus(&child.Status)
			}
		},
		MergeBeforeUpdate: func(current, desired *kedav1alpha1.ScaledObject) {
			current.Labels = desired.Labels
			current.Spec = desired.Spec
		},
		SemanticEquals: func(a1, a2 *kedav1alpha1.ScaledObject) bool {
			return equality.Semantic.DeepEqual(a1.Spec, a2.Spec) &&
				equality.Semantic.DeepEqual(a1.Labels, a2.Labels)
		},

		Config:     c,
		IndexField: ".metadata.processorScaledObjectController",
		Sanitize: func(child *kedav1alpha1.ScaledObject) interface{} {
			return child.Spec
		},
		Setup: func(mgr controllers.Manager, bldr *controllers.Builder) error {
			bldr.Watches(&source.Kind{Type: &corev1.Secret{}}, controllers.EnqueueTracked(&corev1.Secret{}, c.Tracker, c.Scheme))
			return nil
		},
	}
}
