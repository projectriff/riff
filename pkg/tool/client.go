/*
 * Copyright 2018 The original author or authors
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
 *
 */

package tool

import (
	eventing "github.com/knative/eventing/pkg/apis/channels/v1alpha1"
	eventing_cs "github.com/knative/eventing/pkg/client/clientset/versioned"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	serving_cs "github.com/knative/serving/pkg/client/clientset/versioned"
	"k8s.io/client-go/tools/clientcmd"
)

type Client interface {
	CreateSubscription(options CreateSubscriptionOptions) (*eventing.Subscription, error)
	CreateChannel(options CreateChannelOptions) (*eventing.Channel, error)
	CreateFunction(options CreateFunctionOptions) (*serving.Service, error)
	DeleteFunction(options DeleteFunctionOptions) error
}

type client struct {
	eventing     eventing_cs.Interface
	serving      serving_cs.Interface
	clientConfig clientcmd.ClientConfig
}

func NewClient(clientConfig clientcmd.ClientConfig, eventing eventing_cs.Interface, serving serving_cs.Interface) Client {
	return &client{clientConfig: clientConfig, eventing: eventing, serving: serving}
}
