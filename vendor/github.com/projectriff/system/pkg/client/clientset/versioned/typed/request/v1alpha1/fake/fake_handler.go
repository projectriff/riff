/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package fake

import (
	v1alpha1 "github.com/projectriff/system/pkg/apis/request/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeHandlers implements HandlerInterface
type FakeHandlers struct {
	Fake *FakeRequestV1alpha1
	ns   string
}

var handlersResource = schema.GroupVersionResource{Group: "request.projectriff.io", Version: "v1alpha1", Resource: "handlers"}

var handlersKind = schema.GroupVersionKind{Group: "request.projectriff.io", Version: "v1alpha1", Kind: "Handler"}

// Get takes name of the handler, and returns the corresponding handler object, and an error if there is any.
func (c *FakeHandlers) Get(name string, options v1.GetOptions) (result *v1alpha1.Handler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(handlersResource, c.ns, name), &v1alpha1.Handler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Handler), err
}

// List takes label and field selectors, and returns the list of Handlers that match those selectors.
func (c *FakeHandlers) List(opts v1.ListOptions) (result *v1alpha1.HandlerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(handlersResource, handlersKind, c.ns, opts), &v1alpha1.HandlerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.HandlerList{ListMeta: obj.(*v1alpha1.HandlerList).ListMeta}
	for _, item := range obj.(*v1alpha1.HandlerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested handlers.
func (c *FakeHandlers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(handlersResource, c.ns, opts))

}

// Create takes the representation of a handler and creates it.  Returns the server's representation of the handler, and an error, if there is any.
func (c *FakeHandlers) Create(handler *v1alpha1.Handler) (result *v1alpha1.Handler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(handlersResource, c.ns, handler), &v1alpha1.Handler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Handler), err
}

// Update takes the representation of a handler and updates it. Returns the server's representation of the handler, and an error, if there is any.
func (c *FakeHandlers) Update(handler *v1alpha1.Handler) (result *v1alpha1.Handler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(handlersResource, c.ns, handler), &v1alpha1.Handler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Handler), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeHandlers) UpdateStatus(handler *v1alpha1.Handler) (*v1alpha1.Handler, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(handlersResource, "status", c.ns, handler), &v1alpha1.Handler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Handler), err
}

// Delete takes name of the handler and deletes it. Returns an error if one occurs.
func (c *FakeHandlers) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(handlersResource, c.ns, name), &v1alpha1.Handler{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeHandlers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(handlersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.HandlerList{})
	return err
}

// Patch applies the patch and returns the patched handler.
func (c *FakeHandlers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Handler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(handlersResource, c.ns, name, data, subresources...), &v1alpha1.Handler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Handler), err
}
