/*
Copyright 2018 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/knative/eventing/pkg/apis/feeds/v1alpha1"
	"github.com/knative/eventing/pkg/client/clientset/versioned/scheme"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	rest "k8s.io/client-go/rest"
)

type FeedsV1alpha1Interface interface {
	RESTClient() rest.Interface
	ClusterEventSourcesGetter
	ClusterEventTypesGetter
	EventSourcesGetter
	EventTypesGetter
	FeedsGetter
}

// FeedsV1alpha1Client is used to interact with features provided by the feeds.knative.dev group.
type FeedsV1alpha1Client struct {
	restClient rest.Interface
}

func (c *FeedsV1alpha1Client) ClusterEventSources() ClusterEventSourceInterface {
	return newClusterEventSources(c)
}

func (c *FeedsV1alpha1Client) ClusterEventTypes() ClusterEventTypeInterface {
	return newClusterEventTypes(c)
}

func (c *FeedsV1alpha1Client) EventSources(namespace string) EventSourceInterface {
	return newEventSources(c, namespace)
}

func (c *FeedsV1alpha1Client) EventTypes(namespace string) EventTypeInterface {
	return newEventTypes(c, namespace)
}

func (c *FeedsV1alpha1Client) Feeds(namespace string) FeedInterface {
	return newFeeds(c, namespace)
}

// NewForConfig creates a new FeedsV1alpha1Client for the given config.
func NewForConfig(c *rest.Config) (*FeedsV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &FeedsV1alpha1Client{client}, nil
}

// NewForConfigOrDie creates a new FeedsV1alpha1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *FeedsV1alpha1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new FeedsV1alpha1Client for the given RESTClient.
func New(c rest.Interface) *FeedsV1alpha1Client {
	return &FeedsV1alpha1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1alpha1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FeedsV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
