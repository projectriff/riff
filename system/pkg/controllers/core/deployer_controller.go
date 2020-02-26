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

package core

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/projectriff/riff/system/pkg/apis"
	buildv1alpha1 "github.com/projectriff/riff/system/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/projectriff/riff/system/pkg/apis/core/v1alpha1"
	"github.com/projectriff/riff/system/pkg/controllers"
	"github.com/projectriff/riff/system/pkg/refs"
	"github.com/projectriff/riff/system/pkg/tracker"
)

// +kubebuilder:rbac:groups=core.projectriff.io,resources=deployers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.projectriff.io,resources=deployers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=build.projectriff.io,resources=applications;containers;functions,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete

func DeployerReconciler(c controllers.Config) *controllers.ParentReconciler {
	c.Log = c.Log.WithName("Deployer")

	return &controllers.ParentReconciler{
		Type: &corev1alpha1.Deployer{},
		SubReconcilers: []controllers.SubReconciler{
			DeployerBuildRefReconciler(c),
			DeployerChildDeploymentReconciler(c),
			DeployerChildServiceReconciler(c),
			DeployerChildIngressReconciler(c),
		},

		Config: c,
	}
}

func DeployerBuildRefReconciler(c controllers.Config) controllers.SubReconciler {
	c.Log = c.Log.WithName("BuildRef")

	return &controllers.SyncReconciler{
		Sync: func(ctx context.Context, parent *corev1alpha1.Deployer) error {
			build := parent.Spec.Build
			if build == nil {
				parent.Status.LatestImage = parent.Spec.Template.Spec.Containers[0].Image
				return nil
			}

			switch {
			case build.ApplicationRef != "":
				var application buildv1alpha1.Application
				key := types.NamespacedName{Namespace: parent.Namespace, Name: build.ApplicationRef}
				// track application for new images
				c.Tracker.Track(
					tracker.NewKey(application.GetGroupVersionKind(), key),
					types.NamespacedName{Namespace: parent.Namespace, Name: parent.Name},
				)
				if err := c.Get(ctx, key, &application); err != nil {
					if apierrs.IsNotFound(err) {
						return nil
					}
					return err
				}
				if application.Status.LatestImage != "" {
					parent.Status.LatestImage = application.Status.LatestImage
				}
				return nil

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

			return fmt.Errorf("invalid build")
		},

		Config: c,
		Setup: func(mgr controllers.Manager, bldr *controllers.Builder) error {
			bldr.Watches(&source.Kind{Type: &buildv1alpha1.Application{}}, controllers.EnqueueTracked(&buildv1alpha1.Application{}, c.Tracker, c.Scheme))
			bldr.Watches(&source.Kind{Type: &buildv1alpha1.Container{}}, controllers.EnqueueTracked(&buildv1alpha1.Container{}, c.Tracker, c.Scheme))
			bldr.Watches(&source.Kind{Type: &buildv1alpha1.Function{}}, controllers.EnqueueTracked(&buildv1alpha1.Function{}, c.Tracker, c.Scheme))
			return nil
		},
	}
}

func DeployerChildDeploymentReconciler(c controllers.Config) controllers.SubReconciler {
	c.Log = c.Log.WithName("ChildDeployment")

	return &controllers.ChildReconciler{
		ParentType:    &corev1alpha1.Deployer{},
		ChildType:     &appsv1.Deployment{},
		ChildListType: &appsv1.DeploymentList{},

		DesiredChild: func(parent *corev1alpha1.Deployer) (*appsv1.Deployment, error) {
			if parent.Status.LatestImage == "" {
				// no image, skip
				return nil, nil
			}

			labels := controllers.MergeMaps(parent.Labels, map[string]string{
				corev1alpha1.DeployerLabelKey: parent.Name,
			})

			template := *parent.Spec.Template.DeepCopy()
			template.Labels = controllers.MergeMaps(template.Labels, labels)
			targetPort := template.Spec.Containers[0].Ports[0]

			template.Spec.Containers[0].Env = append(template.Spec.Containers[0].Env, corev1.EnvVar{
				Name:  "PORT",
				Value: fmt.Sprintf("%d", targetPort.ContainerPort),
			})
			if template.Spec.Containers[0].ReadinessProbe == nil {
				template.Spec.Containers[0].ReadinessProbe = &corev1.Probe{
					Handler: corev1.Handler{
						TCPSocket: &corev1.TCPSocketAction{
							Port: intstr.FromInt(int(targetPort.ContainerPort)),
						},
					},
				}
			}

			child := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Labels:       labels,
					Annotations:  make(map[string]string),
					GenerateName: fmt.Sprintf("%s-deployer-", parent.Name),
					Namespace:    parent.Namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							corev1alpha1.DeployerLabelKey: parent.Name,
						},
					},
					Template: template,
				},
			}
			if child.Spec.Template.Spec.Containers[0].Image == "" {
				child.Spec.Template.Spec.Containers[0].Image = parent.Status.LatestImage
			}

			return child, nil
		},
		ReflectChildStatusOnParent: func(parent *corev1alpha1.Deployer, child *appsv1.Deployment, err error) {
			if err != nil {
				return
			}
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

func DeployerChildServiceReconciler(c controllers.Config) controllers.SubReconciler {
	c.Log = c.Log.WithName("ChildService")

	return &controllers.ChildReconciler{
		ParentType:    &corev1alpha1.Deployer{},
		ChildType:     &corev1.Service{},
		ChildListType: &corev1.ServiceList{},

		DesiredChild: func(parent *corev1alpha1.Deployer) (*corev1.Service, error) {
			if parent.Status.DeploymentRef == nil {
				// no deployment, skip
				return nil, nil
			}

			targetPort := parent.Spec.Template.Spec.Containers[0].Ports[0]

			child := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Labels: controllers.MergeMaps(parent.Labels, map[string]string{
						corev1alpha1.DeployerLabelKey: parent.Name,
					}),
					Annotations: make(map[string]string),
					Namespace:   parent.Namespace,
					Name:        parent.Name,
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{Name: targetPort.Name, Port: 80, TargetPort: intstr.FromInt(int(targetPort.ContainerPort))},
					},
					Selector: map[string]string{
						corev1alpha1.DeployerLabelKey: parent.Name,
					},
				},
			}

			return child, nil
		},
		ReflectChildStatusOnParent: func(parent *corev1alpha1.Deployer, child *corev1.Service, err error) {
			if err != nil {
				if apierrs.IsAlreadyExists(err) {
					name := err.(apierrs.APIStatus).Status().Details.Name
					parent.Status.MarkServiceNotOwned(name)
				}
				return
			}
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

func DeployerChildIngressReconciler(c controllers.Config) controllers.SubReconciler {
	c.Log = c.Log.WithName("ChildIngress")

	return &controllers.ChildReconciler{
		ParentType:    &corev1alpha1.Deployer{},
		ChildType:     &networkingv1beta1.Ingress{},
		ChildListType: &networkingv1beta1.IngressList{},

		Setup: func(mgr controllers.Manager, bldr *controllers.Builder) error {
			bldr.Watches(&source.Kind{Type: &corev1.ConfigMap{}}, controllers.EnqueueTracked(&corev1.ConfigMap{}, c.Tracker, c.Scheme))
			return nil
		},
		DesiredChild: func(parent *corev1alpha1.Deployer) (*networkingv1beta1.Ingress, error) {
			if parent.Status.ServiceRef == nil || parent.Spec.IngressPolicy == corev1alpha1.IngressPolicyClusterLocal {
				// no service, skip
				return nil, nil
			}

			coreSettings := &corev1.ConfigMap{}
			coreSettingsKey := types.NamespacedName{Namespace: systemNamespace, Name: settingsConfigMapName}

			// track config map
			c.Tracker.Track(
				tracker.NewKey(schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}, coreSettingsKey),
				types.NamespacedName{Namespace: parent.Namespace, Name: parent.Name},
			)
			if err := c.Get(context.TODO(), coreSettingsKey, coreSettings); err != nil {
				c.Log.Error(err, fmt.Sprintf("unable to fetch resource with reference: %s", coreSettingsKey.String()))
				return nil, err
			}

			domain := defaultDomain
			if d := coreSettings.Data[defaultDomainKey]; d != "" {
				domain = d
			}
			host := fmt.Sprintf("%s.%s.%s", parent.Name, parent.Namespace, domain)

			child := &networkingv1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Labels: controllers.MergeMaps(parent.Labels, map[string]string{
						corev1alpha1.DeployerLabelKey: parent.Name,
					}),
					Annotations:  make(map[string]string),
					GenerateName: fmt.Sprintf("%s-deployer-", parent.Name),
					Namespace:    parent.Namespace,
				},
				Spec: networkingv1beta1.IngressSpec{
					Rules: []networkingv1beta1.IngressRule{{
						Host: host,
						IngressRuleValue: networkingv1beta1.IngressRuleValue{
							HTTP: &networkingv1beta1.HTTPIngressRuleValue{
								Paths: []networkingv1beta1.HTTPIngressPath{{
									Path: "/",
									Backend: networkingv1beta1.IngressBackend{
										ServiceName: parent.Status.ServiceRef.Name,
										ServicePort: intstr.FromInt(80),
									},
								}},
							},
						},
					}},
				},
			}

			return child, nil
		},
		ReflectChildStatusOnParent: func(parent *corev1alpha1.Deployer, child *networkingv1beta1.Ingress, err error) {
			if err != nil {
				return
			}
			if child == nil {
				parent.Status.IngressRef = nil
				parent.Status.URL = ""
				if parent.Spec.IngressPolicy == corev1alpha1.IngressPolicyClusterLocal {
					parent.Status.MarkIngressNotRequired()
				}
			} else {
				parent.Status.IngressRef = refs.NewTypedLocalObjectReferenceForObject(child, c.Scheme)
				parent.Status.URL = fmt.Sprintf("http://%s", child.Spec.Rules[0].Host)
				parent.Status.PropagateIngressStatus(&child.Status)
			}
		},
		MergeBeforeUpdate: func(current, desired *networkingv1beta1.Ingress) {
			current.Labels = desired.Labels
			current.Spec = desired.Spec
		},
		SemanticEquals: func(a1, a2 *networkingv1beta1.Ingress) bool {
			return equality.Semantic.DeepEqual(a1.Spec, a2.Spec) &&
				equality.Semantic.DeepEqual(a1.Labels, a2.Labels)
		},

		Config:     c,
		IndexField: ".metadata.ingressController",
		Sanitize: func(child *networkingv1beta1.Ingress) interface{} {
			return child.Spec
		},
	}
}
