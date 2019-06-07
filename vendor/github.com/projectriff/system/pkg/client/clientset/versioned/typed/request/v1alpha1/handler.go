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
package v1alpha1

import (
	v1alpha1 "github.com/projectriff/system/pkg/apis/request/v1alpha1"
	scheme "github.com/projectriff/system/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// HandlersGetter has a method to return a HandlerInterface.
// A group's client should implement this interface.
type HandlersGetter interface {
	Handlers(namespace string) HandlerInterface
}

// HandlerInterface has methods to work with Handler resources.
type HandlerInterface interface {
	Create(*v1alpha1.Handler) (*v1alpha1.Handler, error)
	Update(*v1alpha1.Handler) (*v1alpha1.Handler, error)
	UpdateStatus(*v1alpha1.Handler) (*v1alpha1.Handler, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Handler, error)
	List(opts v1.ListOptions) (*v1alpha1.HandlerList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Handler, err error)
	HandlerExpansion
}

// handlers implements HandlerInterface
type handlers struct {
	client rest.Interface
	ns     string
}

// newHandlers returns a Handlers
func newHandlers(c *RequestV1alpha1Client, namespace string) *handlers {
	return &handlers{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the handler, and returns the corresponding handler object, and an error if there is any.
func (c *handlers) Get(name string, options v1.GetOptions) (result *v1alpha1.Handler, err error) {
	result = &v1alpha1.Handler{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("handlers").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Handlers that match those selectors.
func (c *handlers) List(opts v1.ListOptions) (result *v1alpha1.HandlerList, err error) {
	result = &v1alpha1.HandlerList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("handlers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested handlers.
func (c *handlers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("handlers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a handler and creates it.  Returns the server's representation of the handler, and an error, if there is any.
func (c *handlers) Create(handler *v1alpha1.Handler) (result *v1alpha1.Handler, err error) {
	result = &v1alpha1.Handler{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("handlers").
		Body(handler).
		Do().
		Into(result)
	return
}

// Update takes the representation of a handler and updates it. Returns the server's representation of the handler, and an error, if there is any.
func (c *handlers) Update(handler *v1alpha1.Handler) (result *v1alpha1.Handler, err error) {
	result = &v1alpha1.Handler{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("handlers").
		Name(handler.Name).
		Body(handler).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *handlers) UpdateStatus(handler *v1alpha1.Handler) (result *v1alpha1.Handler, err error) {
	result = &v1alpha1.Handler{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("handlers").
		Name(handler.Name).
		SubResource("status").
		Body(handler).
		Do().
		Into(result)
	return
}

// Delete takes name of the handler and deletes it. Returns an error if one occurs.
func (c *handlers) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("handlers").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *handlers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("handlers").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched handler.
func (c *handlers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Handler, err error) {
	result = &v1alpha1.Handler{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("handlers").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
