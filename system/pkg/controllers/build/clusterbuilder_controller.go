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

package build

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	kpackbuildv1alpha1 "github.com/projectriff/riff/system/pkg/apis/thirdparty/kpack/build/v1alpha1"
)

const buildersConfigMap = "builders"

// ClusterBuilderReconciler reconciles a ClusterBuilder object
type ClusterBuilderReconciler struct {
	client.Client
	Recorder  record.EventRecorder
	Log       logr.Logger
	Namespace string
}

// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=build.pivotal.io,resources=clusterbuilders,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete

func (r *ClusterBuilderReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("configmap", req.NamespacedName)

	if req.Namespace != r.Namespace || req.Name != buildersConfigMap {
		// ignore other configmaps, should never get here
		return ctrl.Result{}, nil
	}

	var originalConfigMap corev1.ConfigMap
	if err := r.Get(ctx, req.NamespacedName, &originalConfigMap); err != nil && !apierrs.IsNotFound(err) {
		log.Error(err, "unable to fetch ConfigMap")
		return ctrl.Result{}, err
	}

	// Don't modify the informers copy
	var configMap corev1.ConfigMap
	if &originalConfigMap != nil {
		configMap = *(originalConfigMap.DeepCopy())
	}

	return r.reconcile(ctx, log, &configMap)
}

func (r *ClusterBuilderReconciler) reconcile(ctx context.Context, log logr.Logger, configMap *corev1.ConfigMap) (ctrl.Result, error) {
	if configMap != nil && configMap.GetDeletionTimestamp() != nil {
		return ctrl.Result{}, nil
	}

	var clusterBuilders kpackbuildv1alpha1.ClusterBuilderList
	if err := r.List(ctx, &clusterBuilders); err != nil {
		log.Error(err, "Failed to get ClusterBuilders", "configmap", configMap)
		return ctrl.Result{Requeue: true}, err
	}

	builderImages := make(map[string]string)
	for _, builder := range clusterBuilders.Items {
		if isTargetClusterBuilder(builder.Name) {
			builderImages[builder.Name] = builder.Status.LatestImage
		}
	}

	if configMap.Name == "" {
		configMap, err := r.createConfigMap(ctx, log, builderImages)
		if err != nil {
			log.Error(err, "Failed to create ConfigMap", "configmap", configMap)
			return ctrl.Result{}, err
		}
	} else {
		configMap, err := r.reconcileConfigMap(ctx, log, configMap, builderImages)
		if err != nil {
			log.Error(err, "Failed to reconcile ConfigMap", "configmap", configMap)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ClusterBuilderReconciler) reconcileConfigMap(ctx context.Context, log logr.Logger, existingConfigMap *corev1.ConfigMap, builderImages map[string]string) (*corev1.ConfigMap, error) {
	configMap := existingConfigMap.DeepCopy()
	configMap.Data = builderImages

	if configMapSemanticEquals(configMap, existingConfigMap) {
		return existingConfigMap, nil
	}

	log.Info("reconciling builders configmap", "diff", cmp.Diff(existingConfigMap.Data, configMap.Data))
	return configMap, r.Update(ctx, configMap)
}

func (r *ClusterBuilderReconciler) createConfigMap(ctx context.Context, log logr.Logger, builderImages map[string]string) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildersConfigMap,
			Namespace: r.Namespace,
		},
		Data: builderImages,
	}
	log.Info("creating builders configmap", "data", configMap.Data)
	return configMap, r.Create(ctx, configMap)
}

func configMapSemanticEquals(desiredConfigMap, configMap *corev1.ConfigMap) bool {
	return equality.Semantic.DeepEqual(desiredConfigMap.Data, configMap.Data)
}

func isTargetClusterBuilder(name string) bool {
	return strings.HasPrefix(name, "riff-")
}

func (r *ClusterBuilderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	enqueueConfigMap := &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {
			if !isTargetClusterBuilder(a.Meta.GetName()) {
				return []reconcile.Request{}
			}
			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Namespace: r.Namespace,
						Name:      buildersConfigMap,
					},
				},
			}
		}),
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				cm, ok := e.Object.(*corev1.ConfigMap)
				if !ok {
					// not a configmap, allow
					return true
				}
				// filter configmap accounts to only be builders in the system namespace
				return cm.Namespace == r.Namespace && cm.Name == buildersConfigMap
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				cm, ok := e.ObjectNew.(*corev1.ConfigMap)
				if !ok {
					// not a configmap, allow
					return true
				}
				// filter configmap accounts to only be builders in the system namespace
				return cm.Namespace == r.Namespace && cm.Name == buildersConfigMap
			},
		}).
		// watch for ClusterBuilder mutations to distil into ConfigMap
		Watches(&source.Kind{Type: &kpackbuildv1alpha1.ClusterBuilder{}}, enqueueConfigMap).
		Complete(r)
}
