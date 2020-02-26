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

package controllers

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/projectriff/system/pkg/apis"
	"github.com/projectriff/system/pkg/tracker"
)

var (
	_ reconcile.Reconciler = (*ParentReconciler)(nil)
)

// Config holds common resources for controllers. The configuration may be
// passed to sub-reconcilers.
type Config struct {
	client.Client
	APIReader client.Reader
	Recorder  record.EventRecorder
	Log       logr.Logger
	Scheme    *runtime.Scheme
	Tracker   tracker.Tracker
}

// ParentReconciler is a controller-runtime reconciler that reconciles a given
// existing resource. The ParentType resource is fetched for the reconciler
// request and passed in turn to each SubReconciler. Finally, the reconciled
// resource's status is compared with the original status, updating the API
// server if needed.
type ParentReconciler struct {
	// Type of resource to reconcile
	Type runtime.Object

	// SubReconcilers are called in order for each reconciler request. If a sub
	// reconciler errs, further sub reconcilers are skipped.
	SubReconcilers []SubReconciler

	Config
}

func (r *ParentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	bldr := ctrl.NewControllerManagedBy(mgr).For(r.Type)
	for _, reconciler := range r.SubReconcilers {
		err := reconciler.SetupWithManager(mgr, bldr)
		if err != nil {
			return err
		}
	}
	return bldr.Complete(r)
}

func (r *ParentReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := WithStash(context.Background())
	log := r.Log.WithValues("request", req.NamespacedName)

	originalParent := r.Type.DeepCopyObject().(apis.Object)

	if err := r.Get(ctx, req.NamespacedName, originalParent); err != nil {
		if apierrs.IsNotFound(err) {
			// we'll ignore not-found errors, since they can't be fixed by an immediate
			// requeue (we'll need to wait for a new notification), and we can get them
			// on deleted requests.
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch resource")
		return ctrl.Result{}, err
	}
	parent := originalParent.DeepCopyObject().(apis.Object)

	if defaulter, ok := parent.(webhook.Defaulter); ok {
		// parent.Default()
		defaulter.Default()
	}
	if initializeConditions := reflect.ValueOf(parent).Elem().FieldByName("Status").Addr().MethodByName("InitializeConditions"); initializeConditions.Kind() == reflect.Func {
		// parent.Status.InitializeConditions()
		initializeConditions.Call([]reflect.Value{})
	}

	result, err := r.reconcile(ctx, parent)

	// check if status has changed before updating
	if !equality.Semantic.DeepEqual(r.status(parent), r.status(originalParent)) && parent.GetDeletionTimestamp() == nil {
		// update status
		log.Info("updating status", "diff", cmp.Diff(r.status(originalParent), r.status(parent)))
		if updateErr := r.Status().Update(ctx, parent); updateErr != nil {
			log.Error(updateErr, "unable to update status", typeName(r.Type), parent)
			r.Recorder.Eventf(parent, corev1.EventTypeWarning, "StatusUpdateFailed",
				"Failed to update status: %v", updateErr)
			return ctrl.Result{}, updateErr
		}
		r.Recorder.Eventf(parent, corev1.EventTypeNormal, "StatusUpdated",
			"Updated status")
	}

	// return original reconcile result
	return result, err
}

func (r *ParentReconciler) reconcile(ctx context.Context, parent apis.Object) (ctrl.Result, error) {
	if parent.GetDeletionTimestamp() != nil {
		return ctrl.Result{}, nil
	}

	for _, reconciler := range r.SubReconcilers {
		if _, err := reconciler.Reconcile(ctx, parent); err != nil {
			return ctrl.Result{}, err
		}
	}

	r.copyGeneration(parent)

	return ctrl.Result{}, nil
}

func (r *ParentReconciler) copyGeneration(obj apis.Object) {
	// obj.Status.ObservedGeneration = obj.Generation
	objVal := reflect.ValueOf(obj).Elem()
	generation := objVal.FieldByName("Generation").Int()
	objVal.FieldByName("Status").FieldByName("ObservedGeneration").SetInt(generation)
}

func (r *ParentReconciler) status(obj apis.Object) interface{} {
	return reflect.ValueOf(obj).Elem().FieldByName("Status").Addr().Interface()
}

// SubReconciler are participants in a larger reconciler request. The resource
// being reconciled is passed directly to the sub reconciler. The resource's
// status can be mutated to reflect the current state.
type SubReconciler interface {
	SetupWithManager(mgr ctrl.Manager, bldr *builder.Builder) error
	Reconcile(ctx context.Context, parent apis.Object) (ctrl.Result, error)
}

var (
	_ SubReconciler = (*SyncReconciler)(nil)
	_ SubReconciler = (*ChildReconciler)(nil)
)

// SyncReconciler is a sub reconciler for custom reconciliation logic. No
// behavior is defined directly.
type SyncReconciler struct {
	// Setup performs initialization on the manager and builder this reconciler
	// will run with. It's common to setup field indexes and watch resources.
	//
	// +optional
	Setup func(mgr ctrl.Manager, bldr *builder.Builder) error

	// Sync does whatever work is necessary for the reconciler
	//
	// Expected function signature:
	//     func(ctx context.Context, parent apis.Object) error
	Sync interface{}

	Config
}

func (r *SyncReconciler) SetupWithManager(mgr ctrl.Manager, bldr *builder.Builder) error {
	if r.Setup == nil {
		return nil
	}
	return r.Setup(mgr, bldr)
}

func (r *SyncReconciler) Reconcile(ctx context.Context, parent apis.Object) (ctrl.Result, error) {
	err := r.sync(ctx, parent)
	if err != nil {
		r.Log.Error(err, "unable to sync", typeName(parent), parent)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *SyncReconciler) sync(ctx context.Context, parent apis.Object) error {
	fn := reflect.ValueOf(r.Sync)
	out := fn.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(parent),
	})
	var err error
	if !out[0].IsNil() {
		err = out[0].Interface().(error)
	}
	return err
}

// ChildReconciler is a sub reconciler that manages a single child resource for
// a parent. The reconciler will ensure that exactly one child will match the
// desired state by:
// - creating a child if none exists
// - updating an existing child
// - removing an unneeded child
// - removing extra children
//
// The flow for each reconciliation request is:
// - DesiredChild
// - if child is desired:
//    - HarmonizeImmutableFields (optional)
//    - SemanticEquals
//    - MergeBeforeUpdate
// - ReflectChildStatusOnParent
//
// During setup, the child resource type is registered to watch for changes. A
// field indexer is configured for the owner on the IndexField.
type ChildReconciler struct {
	// ParentType of resource to reconcile
	ParentType apis.Object
	// ChildType is the resource being created/updated/deleted by the
	// reconciler. For example, a parent Deployment would have a ReplicaSet as a
	// child.
	ChildType apis.Object
	// ChildListType is the listing type for the child type. For example,
	// PodList is the list type for Pod
	ChildListType runtime.Object

	// Setup performs initialization on the manager and builder this reconciler
	// will run with. It's common to setup field indexes and watch resources.
	//
	// +optional
	Setup func(mgr ctrl.Manager, bldr *builder.Builder) error

	// DesiredChild returns the desired child object for the given parent
	// object, or nil if the child should not exist.
	//
	// Expected function signature:
	//     func(parent apis.Object) (apis.Object, error)
	//     func(ctx context.Context, parent apis.Object) (apis.Object, error)
	DesiredChild interface{}

	// ReflectChildStatusOnParent updates the parent object's status with values
	// from the child. Select types of error are passed, including:
	// - apierrs.IsConflict
	//
	// Expected function signature:
	//     func(parent, child apis.Object, err error)
	ReflectChildStatusOnParent interface{}

	// HarmonizeImmutableFields allows fields that are immutable on the current
	// object to be copied to the desired object in order to avoid creating
	// updates which are guaranteed to fail.
	//
	// Expected function signature:
	//     func(current, desired apis.Object)
	//
	// +optional
	HarmonizeImmutableFields interface{}

	// MergeBeforeUpdate copies desired fields on to the current object before
	// calling update. Typically fields to copy are the Spec, Labels and
	// Annotations.
	//
	// Expected function signature:
	//     func(current, desired apis.Object)
	MergeBeforeUpdate interface{}

	// SemanticEquals compares two child resources returning true if there is a
	// meaningful difference that should trigger an update.
	//
	// Expected function signature:
	//     func(a1, a2 apis.Object) bool
	SemanticEquals interface{}

	// Sanitize is called with an object before logging the value. Any value may
	// be returned. A meaningful subset of the resource is typically returned,
	// like the Spec.
	//
	// Expected function signature:
	//     func(child apis.Object) interface{}
	//
	// +optional
	Sanitize interface{}

	Config

	// IndexField is used to index objects of the child's type based on their
	// controlling owner. This field needs to be unique within the manager.
	IndexField string
}

func (r *ChildReconciler) SetupWithManager(mgr ctrl.Manager, bldr *builder.Builder) error {
	bldr.Owns(r.ChildType)

	if err := IndexControllersOfType(mgr, r.IndexField, r.ParentType, r.ChildType, r.Scheme); err != nil {
		return err
	}

	if r.Setup == nil {
		return nil
	}
	return r.Setup(mgr, bldr)
}

func (r *ChildReconciler) Reconcile(ctx context.Context, parent apis.Object) (ctrl.Result, error) {
	child, err := r.reconcile(ctx, parent)
	if err != nil {
		if apierrs.IsAlreadyExists(err) {
			// check if the resource blocking create is owned by the parent.
			// the created child from a previous turn may be slow to appear in the informer cache, but shouldn't appear
			// on the parent as being not ready.
			apierr := err.(apierrs.APIStatus)
			conflicted := r.ChildType.DeepCopyObject().(apis.Object)
			_ = r.APIReader.Get(ctx, types.NamespacedName{Namespace: parent.GetNamespace(), Name: apierr.Status().Details.Name}, conflicted)
			if metav1.IsControlledBy(conflicted, parent) {
				// skip updating the parent's status, fail and try again
				return ctrl.Result{}, err
			}
			r.Log.Info("unable to reconcile child, not owned", typeName(r.ParentType), parent, typeName(r.ChildType), r.sanitize(child))
			r.reflectChildStatusOnParent(parent, child, err)
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, "unable to reconcile child", typeName(r.ParentType), parent)
		return ctrl.Result{}, err
	}
	r.reflectChildStatusOnParent(parent, child, err)

	return ctrl.Result{}, nil
}

func (r *ChildReconciler) reconcile(ctx context.Context, parent apis.Object) (apis.Object, error) {
	actual := r.ChildType.DeepCopyObject().(apis.Object)
	children := r.ChildListType.DeepCopyObject().(runtime.Object)
	if err := r.List(ctx, children, client.InNamespace(parent.GetNamespace()), client.MatchingField(r.IndexField, parent.GetName())); err != nil {
		return nil, err
	}
	// TODO do we need to remove resources pending deletion?
	items := r.items(children)
	if len(items) == 1 {
		actual = items[0]
	} else if len(items) > 1 {
		// this shouldn't happen, delete everything to a clean slate
		for _, extra := range items {
			r.Log.Info("deleting extra child", typeName(r.ChildType), r.sanitize(extra))
			if err := r.Delete(ctx, extra); err != nil {
				r.Recorder.Eventf(parent, corev1.EventTypeWarning, "DeleteFailed",
					"Failed to delete %s %q: %v", typeName(r.ChildType), extra.GetName(), err)
				return nil, err
			}
			r.Recorder.Eventf(parent, corev1.EventTypeNormal, "Deleted",
				"Deleted %s %q", typeName(r.ChildType), extra.GetName())
		}
	}

	desired, err := r.desiredChild(ctx, parent)
	if err != nil {
		return nil, err
	}
	if desired != nil {
		if err := ctrl.SetControllerReference(parent, desired, r.Scheme); err != nil {
			return nil, err
		}
	}

	// delete child if no longer needed
	if desired == nil {
		if !actual.GetCreationTimestamp().Time.IsZero() {
			r.Log.Info("deleting unwanted child", typeName(r.ChildType), r.sanitize(actual))
			if err := r.Delete(ctx, actual); err != nil {
				r.Log.Error(err, "unable to delete unwanted child", typeName(r.ChildType), r.sanitize(actual))
				r.Recorder.Eventf(parent, corev1.EventTypeWarning, "DeleteFailed",
					"Failed to delete %s %q: %v", typeName(r.ChildType), actual.GetName(), err)
				return nil, err
			}
			r.Recorder.Eventf(parent, corev1.EventTypeNormal, "Deleted",
				"Deleted %s %q", typeName(r.ChildType), actual.GetName())
		}
		return nil, nil
	}

	// create child if it doesn't exist
	if actual.GetName() == "" {
		r.Log.Info("creating child", typeName(r.ChildType), r.sanitize(desired))
		if err := r.Create(ctx, desired); err != nil {
			r.Log.Error(err, "unable to create child", typeName(r.ChildType), r.sanitize(desired))
			r.Recorder.Eventf(parent, corev1.EventTypeWarning, "CreationFailed",
				"Failed to create %s %q: %v", typeName(r.ChildType), desired.GetName(), err)
			return nil, err
		}
		r.Recorder.Eventf(parent, corev1.EventTypeNormal, "Created",
			"Created %s %q", typeName(r.ChildType), desired.GetName())
		return desired, nil
	}

	// overwrite fields that should not be mutated
	r.harmonizeImmutableFields(actual, desired)

	if r.semanticEquals(desired, actual) {
		// child is unchanged
		return actual, nil
	}

	// update child with desired changes
	current := actual.DeepCopyObject().(apis.Object)
	r.mergeBeforeUpdate(current, desired)
	r.Log.Info("reconciling child", "diff", cmp.Diff(r.sanitize(actual), r.sanitize(current)))
	if err := r.Update(ctx, current); err != nil {
		r.Log.Error(err, "unable to update child", typeName(r.ChildType), r.sanitize(current))
		r.Recorder.Eventf(parent, corev1.EventTypeWarning, "UpdateFailed",
			"Failed to update %s %q: %v", typeName(r.ChildType), current.GetName(), err)
		return nil, err
	}
	r.Recorder.Eventf(parent, corev1.EventTypeNormal, "Updated",
		"Updated %s %q", typeName(r.ChildType), current.GetName())

	return current, nil
}

func (r *ChildReconciler) semanticEquals(a1, a2 apis.Object) bool {
	fn := reflect.ValueOf(r.SemanticEquals)
	out := fn.Call([]reflect.Value{
		reflect.ValueOf(a1),
		reflect.ValueOf(a2),
	})
	return out[0].Bool()
}

func (r *ChildReconciler) desiredChild(ctx context.Context, parent apis.Object) (apis.Object, error) {
	fn := reflect.ValueOf(r.DesiredChild)
	args := []reflect.Value{}
	if fn.Type().NumIn() == 2 {
		// optional first argument
		args = append(args, reflect.ValueOf(ctx))
	}
	args = append(args, reflect.ValueOf(parent))
	out := fn.Call(args)
	var obj apis.Object
	if !out[0].IsNil() {
		obj = out[0].Interface().(apis.Object)
	}
	var err error
	if !out[1].IsNil() {
		err = out[1].Interface().(error)
	}
	return obj, err
}

func (r *ChildReconciler) reflectChildStatusOnParent(parent, child apis.Object, err error) {
	fn := reflect.ValueOf(r.ReflectChildStatusOnParent)
	args := []reflect.Value{
		reflect.ValueOf(parent),
		reflect.ValueOf(child),
		reflect.ValueOf(err),
	}
	if parent == nil {
		args[0] = reflect.New(fn.Type().In(0)).Elem()
	}
	if child == nil {
		args[1] = reflect.New(fn.Type().In(1)).Elem()
	}
	if err == nil {
		args[2] = reflect.New(fn.Type().In(2)).Elem()
	}
	fn.Call(args)
}

func (r *ChildReconciler) harmonizeImmutableFields(current, desired apis.Object) {
	if r.HarmonizeImmutableFields == nil {
		return
	}
	fn := reflect.ValueOf(r.HarmonizeImmutableFields)
	fn.Call([]reflect.Value{
		reflect.ValueOf(current),
		reflect.ValueOf(desired),
	})
}

func (r *ChildReconciler) mergeBeforeUpdate(current, desired apis.Object) {
	fn := reflect.ValueOf(r.MergeBeforeUpdate)
	fn.Call([]reflect.Value{
		reflect.ValueOf(current),
		reflect.ValueOf(desired),
	})
}

func (r *ChildReconciler) sanitize(child apis.Object) interface{} {
	if r.Sanitize == nil {
		return child
	}
	if child == nil {
		return nil
	}
	fn := reflect.ValueOf(r.Sanitize)
	out := fn.Call([]reflect.Value{
		reflect.ValueOf(child),
	})
	var sanitized interface{}
	if !out[0].IsNil() {
		sanitized = out[0].Interface()
	}
	return sanitized
}

func (r *ChildReconciler) items(children runtime.Object) []apis.Object {
	childrenValue := reflect.ValueOf(children).Elem()
	itemsValue := childrenValue.FieldByName("Items")
	items := make([]apis.Object, itemsValue.Len())
	for i := range items {
		items[i] = itemsValue.Index(i).Addr().Interface().(apis.Object)
	}
	return items
}

func typeName(i interface{}) string {
	t := reflect.TypeOf(i)
	// TODO do we need this?
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// MergeMaps flattens a sequence of maps into a single map. Keys in latter maps
// overwrite previous keys. None of the arguments are mutated.
func MergeMaps(maps ...map[string]string) map[string]string {
	out := map[string]string{}
	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}
