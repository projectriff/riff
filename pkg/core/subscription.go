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
 */

package core

import (
	"github.com/knative/eventing/pkg/apis/channels/v1alpha1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CreateSubscriptionOptions struct {
	Namespaced
	Name       string
	Channel    string
	Subscriber string
}

func (c *client) CreateSubscription(options CreateSubscriptionOptions) (*v1alpha1.Subscription, error) {
	ns := c.explicitOrConfigNamespace(options.Namespaced)

	s := v1alpha1.Subscription{
		TypeMeta: meta_v1.TypeMeta{
			APIVersion: "channels.knative.dev/v1alpha1",
			Kind:       "Subscription",
		},

		ObjectMeta: meta_v1.ObjectMeta{
			Name: options.Name,
		},
		Spec: v1alpha1.SubscriptionSpec{
			Channel:    options.Channel,
			Subscriber: options.Subscriber,
		},
	}

	_, e := c.eventing.ChannelsV1alpha1().Subscriptions(ns).Create(&s)

	return &s, e
}
