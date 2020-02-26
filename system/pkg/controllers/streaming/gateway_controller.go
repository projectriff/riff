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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/projectriff/system/pkg/apis"
	streamingv1alpha1 "github.com/projectriff/system/pkg/apis/streaming/v1alpha1"
	"github.com/projectriff/system/pkg/controllers"
	"github.com/projectriff/system/pkg/refs"
)

// +kubebuilder:rbac:groups=streaming.projectriff.io,resources=gateways,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=streaming.projectriff.io,resources=gateways/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete

func GatewayReconciler(c controllers.Config) *controllers.ParentReconciler {
	c.Log = c.Log.WithName("Gateway")

	return &controllers.ParentReconciler{
		Type: &streamingv1alpha1.Gateway{},
		SubReconcilers: []controllers.SubReconciler{
			GatewayChildServiceReconciler(c),
			GatewayChildDeploymentReconciler(c),
		},

		Config: c,
	}
}

func GatewayChildServiceReconciler(c controllers.Config) controllers.SubReconciler {
	c.Log = c.Log.WithName("ChildService")

	return &controllers.ChildReconciler{
		ParentType:    &streamingv1alpha1.Gateway{},
		ChildType:     &corev1.Service{},
		ChildListType: &corev1.ServiceList{},

		DesiredChild: func(parent *streamingv1alpha1.Gateway) (*corev1.Service, error) {
			if len(parent.Spec.Ports) == 0 {
				return nil, nil
			}

			child := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Labels: controllers.MergeMaps(parent.Labels, map[string]string{
						streamingv1alpha1.GatewayLabelKey: parent.Name,
					}),
					Annotations:  make(map[string]string),
					Namespace:    parent.Namespace,
					GenerateName: fmt.Sprintf("%s-gateway-", parent.Name),
				},
				Spec: corev1.ServiceSpec{
					Ports: parent.Spec.Ports,
					Selector: map[string]string{
						streamingv1alpha1.GatewayLabelKey: parent.Name,
					},
				},
			}

			return child, nil
		},
		ReflectChildStatusOnParent: func(parent *streamingv1alpha1.Gateway, child *corev1.Service, err error) {
			if child == nil {
				parent.Status.ServiceRef = nil
				parent.Status.Address = nil
			} else {
				parent.Status.ServiceRef = refs.NewTypedLocalObjectReferenceForObject(child, c.Scheme)
				parent.Status.Address = &apis.Addressable{URL: fmt.Sprintf("http://%s.%s.%s", child.Name, child.Namespace, "svc.cluster.local")}
				parent.Status.PropagateServiceStatus(&child.Status)
			}
		},
		HarmonizeImmutableFields: func(current, desired *corev1.Service) {
			desired.Spec.ClusterIP = current.Spec.ClusterIP
		},
		MergeBeforeUpdate: func(current, desired *corev1.Service) {
			current.Labels = desired.Labels
			current.Spec = desired.Spec
		},
		SemanticEquals: func(a1, a2 *corev1.Service) bool {
			return equality.Semantic.DeepEqual(a1.Spec, a2.Spec) &&
				equality.Semantic.DeepEqual(a1.Labels, a2.Labels)
		},

		Config:     c,
		IndexField: ".metadata.serviceController",
		Sanitize: func(child *corev1.Service) interface{} {
			return child.Spec
		},
	}
}

func GatewayChildDeploymentReconciler(c controllers.Config) controllers.SubReconciler {
	c.Log = c.Log.WithName("ChildDeployment")

	return &controllers.ChildReconciler{
		ParentType:    &streamingv1alpha1.Gateway{},
		ChildType:     &appsv1.Deployment{},
		ChildListType: &appsv1.DeploymentList{},

		DesiredChild: func(parent *streamingv1alpha1.Gateway) (*appsv1.Deployment, error) {
			if parent.Status.ServiceRef == nil || parent.Spec.Template == nil {
				// no service, skip
				return nil, nil
			}

			labels := controllers.MergeMaps(parent.Labels, map[string]string{
				streamingv1alpha1.GatewayLabelKey: parent.Name,
			})

			template := *parent.Spec.Template.DeepCopy()
			template.Labels = controllers.MergeMaps(template.Labels, labels)

			child := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Labels:       labels,
					Annotations:  make(map[string]string),
					GenerateName: fmt.Sprintf("%s-gateway-", parent.Name),
					Namespace:    parent.Namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							streamingv1alpha1.GatewayLabelKey: parent.Name,
						},
					},
					Template: template,
				},
			}

			return child, nil
		},
		ReflectChildStatusOnParent: func(parent *streamingv1alpha1.Gateway, child *appsv1.Deployment, err error) {
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
		IndexField: ".metadata.deploymentController",
		Sanitize: func(child *appsv1.Deployment) interface{} {
			return child.Spec
		},
	}
}
