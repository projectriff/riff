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

package knative

import (
	"context"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/source"

	buildv1alpha1 "github.com/projectriff/riff/system/pkg/apis/build/v1alpha1"
	knativev1alpha1 "github.com/projectriff/riff/system/pkg/apis/knative/v1alpha1"
	servingv1 "github.com/projectriff/riff/system/pkg/apis/thirdparty/knative/serving/v1"
	"github.com/projectriff/riff/system/pkg/controllers"
	"github.com/projectriff/riff/system/pkg/tracker"
)

// +kubebuilder:rbac:groups=knative.projectriff.io,resources=adapters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=knative.projectriff.io,resources=adapters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=build.projectriff.io,resources=applications;containers;functions,verbs=get;list;watch
// +kubebuilder:rbac:groups=serving.knative.dev,resources=configurations;services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete

func AdapterReconciler(c controllers.Config) *controllers.ParentReconciler {
	c.Log = c.Log.WithName("Adapter")

	return &controllers.ParentReconciler{
		Type: &knativev1alpha1.Adapter{},
		SubReconcilers: []controllers.SubReconciler{
			AdapterBuildRefReconciler(c),
			AdapterTargetRefReconciler(c),
		},

		Config: c,
	}
}

func AdapterBuildRefReconciler(c controllers.Config) controllers.SubReconciler {
	c.Log = c.Log.WithName("BuildRef")

	return &controllers.SyncReconciler{
		Sync: func(ctx context.Context, parent *knativev1alpha1.Adapter) error {
			build := parent.Spec.Build

			switch {
			case build.ApplicationRef != "":
				var application buildv1alpha1.Application
				key := types.NamespacedName{Namespace: parent.Namespace, Name: build.ApplicationRef}
				// track application for new images
				c.Tracker.Track(
					tracker.NewKey(application.GetGroupVersionKind(), key),
					types.NamespacedName{Namespace: parent.Namespace, Name: parent.Name},
				)
				if err := c.Get(ctx, types.NamespacedName{Namespace: parent.Namespace, Name: build.ApplicationRef}, &application); err != nil {
					if apierrs.IsNotFound(err) {
						parent.Status.MarkBuildNotFound("application", build.ApplicationRef)
						return nil
					}
					return err
				}
				if application.Status.LatestImage != "" {
					parent.Status.LatestImage = application.Status.LatestImage
					parent.Status.MarkBuildReady()
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
						parent.Status.MarkBuildNotFound("container", build.ContainerRef)
						return nil
					}
					return err
				}
				if container.Status.LatestImage != "" {
					parent.Status.LatestImage = container.Status.LatestImage
					parent.Status.MarkBuildReady()
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
						parent.Status.MarkBuildNotFound("function", build.FunctionRef)
						return nil
					}
					return err
				}
				if function.Status.LatestImage != "" {
					parent.Status.LatestImage = function.Status.LatestImage
					parent.Status.MarkBuildReady()
				}
				return nil
			}

			return fmt.Errorf("invalid adapter build")
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

func AdapterTargetRefReconciler(c controllers.Config) controllers.SubReconciler {
	c.Log = c.Log.WithName("TargetRef")

	return &controllers.SyncReconciler{
		Sync: func(ctx context.Context, parent *knativev1alpha1.Adapter) error {
			if parent.Status.LatestImage == "" {
				return nil
			}

			target := parent.Spec.Target

			switch {
			case target.ServiceRef != "":
				var actualService servingv1.Service
				key := types.NamespacedName{Namespace: parent.Namespace, Name: target.ServiceRef}
				// track service for changes
				c.Tracker.Track(
					tracker.NewKey(actualService.GetGroupVersionKind(), key),
					types.NamespacedName{Namespace: parent.Namespace, Name: parent.Name},
				)
				if err := c.Get(ctx, types.NamespacedName{Namespace: parent.Namespace, Name: target.ServiceRef}, &actualService); err != nil {
					if errors.IsNotFound(err) {
						parent.Status.MarkTargetNotFound("service", target.ServiceRef)
						return nil
					}
					return err
				}
				parent.Status.MarkTargetFound()

				if actualService.Spec.Template.Spec.Containers[0].Image == parent.Status.LatestImage {
					// already latest image
					return nil
				}

				// update service
				service := *(actualService.DeepCopy())
				service.Spec.Template.Spec.Containers[0].Image = parent.Status.LatestImage
				c.Log.Info("reconciling service", "diff", cmp.Diff(actualService.Spec, service.Spec))
				return c.Update(ctx, &service)

			case target.ConfigurationRef != "":
				var actualConfiguration servingv1.Configuration
				key := types.NamespacedName{Namespace: parent.Namespace, Name: target.ConfigurationRef}
				// track configuration for changes
				c.Tracker.Track(
					tracker.NewKey(actualConfiguration.GetGroupVersionKind(), key),
					types.NamespacedName{Namespace: parent.Namespace, Name: parent.Name},
				)
				if err := c.Get(ctx, key, &actualConfiguration); err != nil {
					if errors.IsNotFound(err) {
						parent.Status.MarkTargetNotFound("configuration", target.ConfigurationRef)
						return nil
					}
					return err
				}
				parent.Status.MarkTargetFound()

				if actualConfiguration.Spec.Template.Spec.Containers[0].Image == parent.Status.LatestImage {
					// already latest image
					return nil
				}

				// update configuration
				configuration := *(actualConfiguration.DeepCopy())
				configuration.Spec.Template.Spec.Containers[0].Image = parent.Status.LatestImage
				c.Log.Info("reconciling configuration", "diff", cmp.Diff(actualConfiguration.Spec, configuration.Spec))
				return c.Update(ctx, &configuration)

			}

			return fmt.Errorf("invalid adapter target")
		},

		Config: c,
		Setup: func(mgr controllers.Manager, bldr *controllers.Builder) error {
			bldr.Watches(&source.Kind{Type: &servingv1.Service{}}, controllers.EnqueueTracked(&servingv1.Service{}, c.Tracker, c.Scheme))
			bldr.Watches(&source.Kind{Type: &servingv1.Configuration{}}, controllers.EnqueueTracked(&servingv1.Configuration{}, c.Tracker, c.Scheme))
			return nil
		},
	}
}
