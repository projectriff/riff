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

package testing

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgotesting "k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/projectriff/system/pkg/apis"
)

type clientWrapper struct {
	client              client.Client
	scheme              *runtime.Scheme
	createActions       []objectAction
	updateActions       []objectAction
	deleteActions       []DeleteAction
	statusUpdateActions []objectAction
	genCount            int
	reactionChain       []Reactor
}

var _ client.Client = &clientWrapper{}

func newClientWrapperWithScheme(scheme *runtime.Scheme, objs ...runtime.Object) *clientWrapper {
	client := &clientWrapper{
		client:              fakeclient.NewFakeClientWithScheme(scheme, objs...),
		scheme:              scheme,
		createActions:       []objectAction{},
		updateActions:       []objectAction{},
		deleteActions:       []DeleteAction{},
		statusUpdateActions: []objectAction{},
		genCount:            0,
		reactionChain:       []Reactor{},
	}
	// generate names on create
	client.AddReactor("create", "*", func(action Action) (bool, runtime.Object, error) {
		if createAction, ok := action.(CreateAction); ok {
			obj := createAction.GetObject()
			if accessor, ok := obj.(metav1.ObjectMetaAccessor); ok {
				objmeta := accessor.GetObjectMeta()
				if objmeta.GetName() == "" && objmeta.GetGenerateName() != "" {
					client.genCount++
					// mutate the existing obj
					objmeta.SetName(fmt.Sprintf("%s%03d", objmeta.GetGenerateName(), client.genCount))
				}
			}
		}
		// never handle the action
		return false, nil, nil
	})
	return client
}

func (w *clientWrapper) AddReactor(verb, kind string, reaction ReactionFunc) {
	w.reactionChain = append(w.reactionChain, &clientgotesting.SimpleReactor{Verb: verb, Resource: kind, Reaction: reaction})
}

func (w *clientWrapper) PrependReactor(verb, kind string, reaction ReactionFunc) {
	w.reactionChain = append([]Reactor{&clientgotesting.SimpleReactor{Verb: verb, Resource: kind, Reaction: reaction}}, w.reactionChain...)
}

func (w *clientWrapper) objmeta(obj runtime.Object) (schema.GroupVersionResource, string, string, error) {
	gvks, _, err := w.scheme.ObjectKinds(obj)
	if err != nil {
		return schema.GroupVersionResource{}, "", "", err
	}
	gvk := gvks[0]
	// NOTE kind != resource, but for this purpose it's good enough
	gvr := schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: gvk.Kind,
	}

	if objmeta, ok := obj.(metav1.ObjectMetaAccessor); ok {
		return gvr, objmeta.GetObjectMeta().GetNamespace(), objmeta.GetObjectMeta().GetName(), nil
	}
	if _, ok := obj.(metav1.ListMetaAccessor); ok {
		return gvr, "", "", nil
	}

	return schema.GroupVersionResource{}, "", "", fmt.Errorf("invalid object")
}

func (w *clientWrapper) react(action Action) error {
	for _, reactor := range w.reactionChain {
		if !reactor.Handles(action) {
			continue
		}
		handled, _, err := reactor.React(action)
		if !handled {
			continue
		}
		return err
	}
	return nil
}

func (w *clientWrapper) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	gvr, namespace, name, err := w.objmeta(obj)
	if err != nil {
		return err
	}

	// call reactor chain
	err = w.react(clientgotesting.NewGetAction(gvr, namespace, name))
	if err != nil {
		return err
	}

	return w.client.Get(ctx, key, obj)
}

func (w *clientWrapper) List(ctx context.Context, list runtime.Object, opts ...client.ListOption) error {
	gvr, _, _, err := w.objmeta(list)
	if err != nil {
		return err
	}
	gvk := schema.GroupVersionKind{
		Group:   gvr.Group,
		Version: gvr.Version,
		Kind:    gvr.Resource,
	}
	listopts := &client.ListOptions{}
	for _, opt := range opts {
		opt.ApplyToList(listopts)
	}

	// call reactor chain
	err = w.react(clientgotesting.NewListAction(gvr, gvk, listopts.Namespace, metav1.ListOptions{}))
	if err != nil {
		return err
	}

	return w.client.List(ctx, list, opts...)
}

func (w *clientWrapper) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	gvr, namespace, _, err := w.objmeta(obj)
	if err != nil {
		return err
	}

	// capture action
	w.createActions = append(w.createActions, clientgotesting.NewCreateAction(gvr, namespace, obj.DeepCopyObject()))

	// call reactor chain
	err = w.react(clientgotesting.NewCreateAction(gvr, namespace, obj))
	if err != nil {
		return err
	}

	return w.client.Create(ctx, obj, opts...)
}

func (w *clientWrapper) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error {
	gvr, namespace, name, err := w.objmeta(obj)
	if err != nil {
		return err
	}

	// capture action
	w.deleteActions = append(w.deleteActions, clientgotesting.NewDeleteAction(gvr, namespace, name))

	// call reactor chain
	err = w.react(clientgotesting.NewDeleteAction(gvr, namespace, name))
	if err != nil {
		return err
	}

	return w.client.Delete(ctx, obj, opts...)
}

func (w *clientWrapper) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	gvr, namespace, _, err := w.objmeta(obj)
	if err != nil {
		return err
	}

	// capture action
	w.updateActions = append(w.updateActions, clientgotesting.NewUpdateAction(gvr, namespace, obj.DeepCopyObject()))

	// call reactor chain
	err = w.react(clientgotesting.NewUpdateAction(gvr, namespace, obj))
	if err != nil {
		return err
	}

	return w.client.Update(ctx, obj, opts...)
}
func (w *clientWrapper) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	panic(fmt.Errorf("Patch() is not implemented"))
}

func (w *clientWrapper) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...client.DeleteAllOfOption) error {
	panic(fmt.Errorf("DeleteAllOf() is not implemented"))
}

func (w *clientWrapper) Status() client.StatusWriter {
	return &statusWriterWrapper{
		statusWriter:  w.client.Status(),
		clientWrapper: w,
	}
}

type statusWriterWrapper struct {
	statusWriter  client.StatusWriter
	clientWrapper *clientWrapper
}

var _ client.StatusWriter = &statusWriterWrapper{}

func (w *statusWriterWrapper) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	gvr, namespace, _, err := w.clientWrapper.objmeta(obj)
	if err != nil {
		return err
	}

	// capture action
	w.clientWrapper.statusUpdateActions = append(w.clientWrapper.statusUpdateActions, clientgotesting.NewUpdateSubresourceAction(gvr, "status", namespace, obj.DeepCopyObject()))

	// call reactor chain
	err = w.clientWrapper.react(clientgotesting.NewUpdateSubresourceAction(gvr, "status", namespace, obj))
	if err != nil {
		return err
	}

	return w.statusWriter.Update(ctx, obj, opts...)
}

func (w *statusWriterWrapper) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	panic(fmt.Errorf("Patch() is not implemented"))
}

// InduceFailure is used in conjunction with TableTest's WithReactors field.
// Tests that want to induce a failure in a row of a TableTest would add:
//   WithReactors: []rifftesting.ReactionFunc{
//      // Makes calls to create stream return an error.
//      rifftesting.InduceFailure("create", "Stream"),
//   },
func InduceFailure(verb, kind string, o ...InduceFailureOpts) ReactionFunc {
	var opts *InduceFailureOpts
	switch len(o) {
	case 0:
		opts = &InduceFailureOpts{}
	case 1:
		opts = &o[0]
	default:
		panic(fmt.Errorf("expected exactly zero or one InduceFailureOpts, got %v", o))
	}
	return func(action Action) (handled bool, ret runtime.Object, err error) {
		if !action.Matches(verb, kind) {
			return false, nil, nil
		}
		if opts.Namespace != "" && opts.Namespace != action.GetNamespace() {
			return false, nil, nil
		}
		if opts.Name != "" {
			switch a := action.(type) {
			case namedAction: // matches GetAction, PatchAction, DeleteAction
				if opts.Name != a.GetName() {
					return false, nil, nil
				}
			case objectAction: // matches CreateAction, UpdateAction
				obj, ok := a.GetObject().(apis.Object)
				if ok && opts.Name != obj.GetName() {
					return false, nil, nil
				}
			}
		}
		if opts.SubResource != "" && opts.SubResource != action.GetSubresource() {
			return false, nil, nil
		}
		err = opts.Error
		if err == nil {
			err = fmt.Errorf("inducing failure for %s %s", action.GetVerb(), action.GetResource().Resource)
		}
		return true, nil, err
	}
}

type namedAction interface {
	Action
	GetName() string
}

type objectAction interface {
	Action
	GetObject() runtime.Object
}

type InduceFailureOpts struct {
	Error       error
	Namespace   string
	Name        string
	SubResource string
}
