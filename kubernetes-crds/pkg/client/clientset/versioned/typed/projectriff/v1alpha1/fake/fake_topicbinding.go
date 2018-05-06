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

package fake

import (
	v1alpha1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeTopicBindings implements TopicBindingInterface
type FakeTopicBindings struct {
	Fake *FakeProjectriffV1alpha1
	ns   string
}

var topicbindingsResource = schema.GroupVersionResource{Group: "projectriff.io", Version: "v1alpha1", Resource: "topicbindings"}

var topicbindingsKind = schema.GroupVersionKind{Group: "projectriff.io", Version: "v1alpha1", Kind: "TopicBinding"}

// Get takes name of the topicBinding, and returns the corresponding topicBinding object, and an error if there is any.
func (c *FakeTopicBindings) Get(name string, options v1.GetOptions) (result *v1alpha1.TopicBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(topicbindingsResource, c.ns, name), &v1alpha1.TopicBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.TopicBinding), err
}

// List takes label and field selectors, and returns the list of TopicBindings that match those selectors.
func (c *FakeTopicBindings) List(opts v1.ListOptions) (result *v1alpha1.TopicBindingList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(topicbindingsResource, topicbindingsKind, c.ns, opts), &v1alpha1.TopicBindingList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.TopicBindingList{}
	for _, item := range obj.(*v1alpha1.TopicBindingList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested topicBindings.
func (c *FakeTopicBindings) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(topicbindingsResource, c.ns, opts))

}

// Create takes the representation of a topicBinding and creates it.  Returns the server's representation of the topicBinding, and an error, if there is any.
func (c *FakeTopicBindings) Create(topicBinding *v1alpha1.TopicBinding) (result *v1alpha1.TopicBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(topicbindingsResource, c.ns, topicBinding), &v1alpha1.TopicBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.TopicBinding), err
}

// Update takes the representation of a topicBinding and updates it. Returns the server's representation of the topicBinding, and an error, if there is any.
func (c *FakeTopicBindings) Update(topicBinding *v1alpha1.TopicBinding) (result *v1alpha1.TopicBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(topicbindingsResource, c.ns, topicBinding), &v1alpha1.TopicBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.TopicBinding), err
}

// Delete takes name of the topicBinding and deletes it. Returns an error if one occurs.
func (c *FakeTopicBindings) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(topicbindingsResource, c.ns, name), &v1alpha1.TopicBinding{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeTopicBindings) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(topicbindingsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.TopicBindingList{})
	return err
}

// Patch applies the patch and returns the patched topicBinding.
func (c *FakeTopicBindings) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.TopicBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(topicbindingsResource, c.ns, name, data, subresources...), &v1alpha1.TopicBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.TopicBinding), err
}
