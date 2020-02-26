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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	streamingv1alpha1 "github.com/projectriff/riff/system/pkg/apis/streaming/v1alpha1"
	"github.com/projectriff/riff/system/pkg/controllers"
	"github.com/projectriff/riff/system/pkg/tracker"
)

const streamAddressStashKey controllers.StashKey = "stream-address"

// +kubebuilder:rbac:groups=streaming.projectriff.io,resources=streams,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=streaming.projectriff.io,resources=streams/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete

func StreamReconciler(c controllers.Config, provisioner StreamProvisionerClient) *controllers.ParentReconciler {
	c.Log = c.Log.WithName("Stream")

	return &controllers.ParentReconciler{
		Type: &streamingv1alpha1.Stream{},
		SubReconcilers: []controllers.SubReconciler{
			StreamProvisionReconciler(c, provisioner),
			StreamChildBindingMetadataReconciler(c),
			StreamChildBindingSecretReconciler(c),
			StreamSyncBindingCondition(c),
		},

		Config: c,
	}
}

func StreamProvisionReconciler(c controllers.Config, provisioner StreamProvisionerClient) controllers.SubReconciler {
	c.Log = c.Log.WithName("Provision")

	return &controllers.SyncReconciler{
		Sync: func(ctx context.Context, stream *streamingv1alpha1.Stream) error {
			// delegate to the provisioner via its REST API
			var provisionerURL string

			var gateway streamingv1alpha1.Gateway
			gatewayKey := types.NamespacedName{Namespace: stream.Namespace, Name: stream.Spec.Gateway.Name}
			c.Tracker.Track(
				tracker.NewKey(gateway.GetGroupVersionKind(), gatewayKey),
				types.NamespacedName{Namespace: stream.Namespace, Name: stream.Name},
			)
			if err := c.Get(ctx, gatewayKey, &gateway); err != nil {
				if apierrs.IsNotFound(err) {
					stream.Status.MarkStreamProvisionFailed(fmt.Sprintf("Gateway %q not found", gatewayKey.Name))
					return nil
				}
				stream.Status.MarkStreamProvisionFailed(err.Error())
				return err
			}
			if gateway.Status.Address == nil || !gateway.Status.IsReady() {
				stream.Status.MarkStreamProvisionFailed(fmt.Sprintf("Gateway %q not ready", gatewayKey.Name))
				return nil
			}
			url, err := gateway.Status.Address.Parse()
			if err != nil {
				return err
			}
			provisionerURL = fmt.Sprintf("http://%s/%s/%s", url.Hostname(), stream.Namespace, stream.Name)

			address, err := provisioner.ProvisionStream(stream, provisionerURL)
			if err != nil {
				stream.Status.MarkStreamProvisionFailed(err.Error())
				return err
			}
			// stash for later child reconcilers
			controllers.StashValue(ctx, streamAddressStashKey, *address)
			stream.Status.MarkStreamProvisioned()
			return nil
		},

		Config: c,
	}
}

func StreamChildBindingMetadataReconciler(c controllers.Config) controllers.SubReconciler {
	c.Log = c.Log.WithName("ChildBindingMetadata")

	return &controllers.ChildReconciler{
		ParentType:    &streamingv1alpha1.Stream{},
		ChildType:     &corev1.ConfigMap{},
		ChildListType: &corev1.ConfigMapList{},

		DesiredChild: func(ctx context.Context, parent *streamingv1alpha1.Stream) (*corev1.ConfigMap, error) {
			_, ok := controllers.RetrieveValue(ctx, streamAddressStashKey).(StreamAddress)
			if !ok {
				return nil, nil
			}

			child := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Labels: controllers.MergeMaps(parent.Labels, map[string]string{
						streamingv1alpha1.StreamLabelKey: parent.Name,
					}),
					Annotations: make(map[string]string),
					Name:        fmt.Sprintf("%s-stream-binding-metadata", parent.Name),
					Namespace:   parent.Namespace,
				},
				Data: map[string]string{
					// spec required values
					"kind":     (&streamingv1alpha1.Stream{}).GetGroupVersionKind().GroupKind().String(),
					"provider": "riff Streaming",
					"tags":     "",
					// non-spec values
					"stream":      parent.Name,
					"contentType": parent.Spec.ContentType,
				},
			}

			return child, nil
		},
		ReflectChildStatusOnParent: func(parent *streamingv1alpha1.Stream, child *corev1.ConfigMap, err error) {
			if err != nil {
				if apierrs.IsAlreadyExists(err) {
					parent.Status.Binding.MetadataRef = corev1.LocalObjectReference{}
					name := err.(apierrs.APIStatus).Status().Details.Name
					parent.Status.MarkBindingNotReady("binding metadata %q already exists", name)
				}
				return
			}
			if parent.Status.GetCondition(streamingv1alpha1.StreamConditionResourceAvailable).IsFalse() {
				return
			}
			parent.Status.Binding.MetadataRef = corev1.LocalObjectReference{
				Name: child.Name,
			}
		},
		MergeBeforeUpdate: func(current, desired *corev1.ConfigMap) {
			current.Labels = desired.Labels
			current.Data = desired.Data
		},
		SemanticEquals: func(a1, a2 *corev1.ConfigMap) bool {
			return equality.Semantic.DeepEqual(a1.Data, a2.Data) &&
				equality.Semantic.DeepEqual(a1.Labels, a2.Labels)
		},

		Config:     c,
		IndexField: ".metadata.bindingMetadataController",
		Sanitize: func(child *corev1.ConfigMap) interface{} {
			return child.Data
		},
	}
}

func StreamChildBindingSecretReconciler(c controllers.Config) controllers.SubReconciler {
	c.Log = c.Log.WithName("ChildBindingSecret")

	return &controllers.ChildReconciler{
		ParentType:    &streamingv1alpha1.Stream{},
		ChildType:     &corev1.Secret{},
		ChildListType: &corev1.SecretList{},

		DesiredChild: func(ctx context.Context, parent *streamingv1alpha1.Stream) (*corev1.Secret, error) {
			address, ok := controllers.RetrieveValue(ctx, streamAddressStashKey).(StreamAddress)
			if !ok || address.Gateway == "" || address.Topic == "" {
				return nil, nil
			}

			child := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels: controllers.MergeMaps(parent.Labels, map[string]string{
						streamingv1alpha1.StreamLabelKey: parent.Name,
					}),
					Annotations: make(map[string]string),
					Name:        fmt.Sprintf("%s-stream-binding-secret", parent.Name),
					Namespace:   parent.Namespace,
				},
				Data: map[string][]byte{
					"gateway": []byte(address.Gateway),
					"topic":   []byte(address.Topic),
				},
			}

			return child, nil
		},
		ReflectChildStatusOnParent: func(parent *streamingv1alpha1.Stream, child *corev1.Secret, err error) {
			if err != nil {
				if apierrs.IsAlreadyExists(err) {
					parent.Status.Binding.SecretRef = corev1.LocalObjectReference{}
					name := err.(apierrs.APIStatus).Status().Details.Name
					parent.Status.MarkBindingNotReady("binding secret %q already exists", name)
				}
				return
			}
			if parent.Status.GetCondition(streamingv1alpha1.StreamConditionResourceAvailable).IsFalse() {
				return
			}
			parent.Status.Binding.SecretRef = corev1.LocalObjectReference{
				Name: child.Name,
			}
		},
		MergeBeforeUpdate: func(current, desired *corev1.Secret) {
			current.Labels = desired.Labels
			current.Data = desired.Data
		},
		SemanticEquals: func(a1, a2 *corev1.Secret) bool {
			return equality.Semantic.DeepEqual(a1.Data, a2.Data) &&
				equality.Semantic.DeepEqual(a1.Labels, a2.Labels)
		},

		Config:     c,
		IndexField: ".metadata.bindingSecretController",
		Sanitize: func(child *corev1.Secret) interface{} {
			return child.Name
		},
	}
}

func StreamSyncBindingCondition(c controllers.Config) controllers.SubReconciler {
	c.Log = c.Log.WithName("BindingCondition")

	return &controllers.SyncReconciler{
		Sync: func(ctx context.Context, stream *streamingv1alpha1.Stream) error {
			// aggregate binding metadata and secret into a single condition
			if stream.Status.Binding.MetadataRef.Name == "" || stream.Status.Binding.SecretRef.Name == "" {
				return nil
			}
			if stream.Status.GetCondition(streamingv1alpha1.StreamConditionBindingReady).IsUnknown() {
				stream.Status.MarkBindingReady()
			}
			return nil
		},

		Config: c,
	}
}
