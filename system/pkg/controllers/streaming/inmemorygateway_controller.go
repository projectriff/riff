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

package streaming

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/source"

	streamingv1alpha1 "github.com/projectriff/riff/system/pkg/apis/streaming/v1alpha1"
	"github.com/projectriff/riff/system/pkg/controllers"
	"github.com/projectriff/riff/system/pkg/refs"
	"github.com/projectriff/riff/system/pkg/tracker"
)

// +kubebuilder:rbac:groups=streaming.projectriff.io,resources=inmemorygateways,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=streaming.projectriff.io,resources=inmemorygateways/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=streaming.projectriff.io,resources=gateways,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete

func InMemoryGatewayReconciler(c controllers.Config, namespace string) *controllers.ParentReconciler {
	c.Log = c.Log.WithName("InMemoryGateway")

	return &controllers.ParentReconciler{
		Type: &streamingv1alpha1.InMemoryGateway{},
		SubReconcilers: []controllers.SubReconciler{
			InMemoryGatewaySyncConfigReconciler(c, namespace),
			InMemoryGatewayChildGatewayReconciler(c),
		},

		Config: c,
	}
}

func InMemoryGatewaySyncConfigReconciler(c controllers.Config, namespace string) controllers.SubReconciler {
	c.Log = c.Log.WithName("SyncConfig")

	return &controllers.SyncReconciler{
		Sync: func(ctx context.Context, parent *streamingv1alpha1.InMemoryGateway) error {
			var config corev1.ConfigMap
			key := types.NamespacedName{Namespace: namespace, Name: inmemoryGatewayImages}
			// track config for new images
			c.Tracker.Track(
				tracker.NewKey(schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}, key),
				types.NamespacedName{Namespace: parent.Namespace, Name: parent.Name},
			)
			if err := c.Get(ctx, key, &config); err != nil {
				return err
			}
			parent.Status.GatewayImage = config.Data[gatewayImageKey]
			parent.Status.ProvisionerImage = config.Data[provisionerImageKey]
			return nil
		},

		Config: c,
		Setup: func(mgr controllers.Manager, bldr *controllers.Builder) error {
			bldr.Watches(&source.Kind{Type: &corev1.ConfigMap{}}, controllers.EnqueueTracked(&corev1.ConfigMap{}, c.Tracker, c.Scheme))
			return nil
		},
	}
}

func InMemoryGatewayChildGatewayReconciler(c controllers.Config) controllers.SubReconciler {
	c.Log = c.Log.WithName("ChildGateway")

	return &controllers.ChildReconciler{
		ParentType:    &streamingv1alpha1.InMemoryGateway{},
		ChildType:     &streamingv1alpha1.Gateway{},
		ChildListType: &streamingv1alpha1.GatewayList{},

		DesiredChild: func(parent *streamingv1alpha1.InMemoryGateway) (*streamingv1alpha1.Gateway, error) {
			labels := controllers.MergeMaps(parent.Labels, map[string]string{
				streamingv1alpha1.InMemoryGatewayLabelKey: parent.Name,
				streamingv1alpha1.GatewayTypeLabelKey:     streamingv1alpha1.InMemoryGatewayType,
			})

			var template *corev1.PodTemplateSpec
			if parent.Status.Address != nil {
				gatewayAddress, err := parent.Status.Address.Parse()
				if err != nil {
					return nil, err
				}

				template = &corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: labels,
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "gateway",
								Image: parent.Status.GatewayImage,
								Env: []corev1.EnvVar{
									{Name: "storage_positions_type", Value: "MEMORY"},
									{Name: "storage_records_type", Value: "MEMORY"},
									{Name: "server_port", Value: "8000"},
								},
							},
							{
								Name:  "provisioner",
								Image: parent.Status.ProvisionerImage,
								Env: []corev1.EnvVar{
									{Name: "GATEWAY", Value: fmt.Sprintf("%s:6565", gatewayAddress.Hostname())},
								},
							},
						},
					},
				}
			}

			child := &streamingv1alpha1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: make(map[string]string),
					Name:        parent.Name,
					Namespace:   parent.Namespace,
				},
				Spec: streamingv1alpha1.GatewaySpec{
					Template: template,
					Ports: []corev1.ServicePort{
						{Name: "gateway", Port: 6565},
						{Name: "provisioner", Port: 80, TargetPort: intstr.FromInt(8080)},
					},
				},
			}

			return child, nil
		},
		ReflectChildStatusOnParent: func(parent *streamingv1alpha1.InMemoryGateway, child *streamingv1alpha1.Gateway, err error) {
			if err != nil {
				if apierrs.IsAlreadyExists(err) {
					name := err.(apierrs.APIStatus).Status().Details.Name
					parent.Status.MarkGatewayNotOwned(name)
				}
				return
			}
			parent.Status.GatewayRef = refs.NewTypedLocalObjectReferenceForObject(child, c.Scheme)
			parent.Status.Address = child.Status.Address
			parent.Status.PropagateGatewayStatus(&child.Status)
		},
		MergeBeforeUpdate: func(current, desired *streamingv1alpha1.Gateway) {
			current.Labels = desired.Labels
			current.Spec = desired.Spec
		},
		SemanticEquals: func(a1, a2 *streamingv1alpha1.Gateway) bool {
			return equality.Semantic.DeepEqual(a1.Spec, a2.Spec) &&
				equality.Semantic.DeepEqual(a1.Labels, a2.Labels)
		},

		Config:     c,
		IndexField: ".metadata.inmemoryGatewayController",
		Sanitize: func(child *streamingv1alpha1.Gateway) interface{} {
			return child.Spec
		},
	}
}
