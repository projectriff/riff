/*
 * Copyright 2018 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

import (
	v1alpha1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1alpha1"
	scheme "github.com/projectriff/riff/kubernetes-crds/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// TopicBindingsGetter has a method to return a TopicBindingInterface.
// A group's client should implement this interface.
type TopicBindingsGetter interface {
	TopicBindings(namespace string) TopicBindingInterface
}

// TopicBindingInterface has methods to work with TopicBinding resources.
type TopicBindingInterface interface {
	Create(*v1alpha1.TopicBinding) (*v1alpha1.TopicBinding, error)
	Update(*v1alpha1.TopicBinding) (*v1alpha1.TopicBinding, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.TopicBinding, error)
	List(opts v1.ListOptions) (*v1alpha1.TopicBindingList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.TopicBinding, err error)
	TopicBindingExpansion
}

// topicBindings implements TopicBindingInterface
type topicBindings struct {
	client rest.Interface
	ns     string
}

// newTopicBindings returns a TopicBindings
func newTopicBindings(c *ProjectriffV1alpha1Client, namespace string) *topicBindings {
	return &topicBindings{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the topicBinding, and returns the corresponding topicBinding object, and an error if there is any.
func (c *topicBindings) Get(name string, options v1.GetOptions) (result *v1alpha1.TopicBinding, err error) {
	result = &v1alpha1.TopicBinding{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("topicbindings").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of TopicBindings that match those selectors.
func (c *topicBindings) List(opts v1.ListOptions) (result *v1alpha1.TopicBindingList, err error) {
	result = &v1alpha1.TopicBindingList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("topicbindings").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested topicBindings.
func (c *topicBindings) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("topicbindings").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a topicBinding and creates it.  Returns the server's representation of the topicBinding, and an error, if there is any.
func (c *topicBindings) Create(topicBinding *v1alpha1.TopicBinding) (result *v1alpha1.TopicBinding, err error) {
	result = &v1alpha1.TopicBinding{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("topicbindings").
		Body(topicBinding).
		Do().
		Into(result)
	return
}

// Update takes the representation of a topicBinding and updates it. Returns the server's representation of the topicBinding, and an error, if there is any.
func (c *topicBindings) Update(topicBinding *v1alpha1.TopicBinding) (result *v1alpha1.TopicBinding, err error) {
	result = &v1alpha1.TopicBinding{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("topicbindings").
		Name(topicBinding.Name).
		Body(topicBinding).
		Do().
		Into(result)
	return
}

// Delete takes name of the topicBinding and deletes it. Returns an error if one occurs.
func (c *topicBindings) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("topicbindings").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *topicBindings) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("topicbindings").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched topicBinding.
func (c *topicBindings) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.TopicBinding, err error) {
	result = &v1alpha1.TopicBinding{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("topicbindings").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
