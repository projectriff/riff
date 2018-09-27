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
	"fmt"

	"github.com/knative/eventing/pkg/apis/channels/v1alpha1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CreateSubscriptionOptions struct {
	Namespace  string
	Name       string
	Channel    string
	Subscriber string
	ReplyTo    string
	DryRun     bool
}

type DeleteSubscriptionOptions struct {
	Namespace string
	Name      string
}

type ListSubscriptionsOptions struct {
	Namespace string
}

func (c *client) CreateSubscription(options CreateSubscriptionOptions) (*v1alpha1.Subscription, error) {
	ns := c.explicitOrConfigNamespace(options.Namespace)
	if options.ReplyTo != "" {
		options.ReplyTo = fmt.Sprintf("%s-channel", options.ReplyTo)
	}
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
			ReplyTo:    options.ReplyTo,
		},
	}

	if !options.DryRun {
		_, e := c.eventing.ChannelsV1alpha1().Subscriptions(ns).Create(&s)
		return &s, e
	} else {
		return &s, nil
	}

}

func (c *client) DeleteSubscription(options DeleteSubscriptionOptions) error {
	ns := c.explicitOrConfigNamespace(options.Namespace)
	return c.eventing.ChannelsV1alpha1().Subscriptions(ns).Delete(options.Name, nil)
}

func (c *client) ListSubscriptions(options ListSubscriptionsOptions) (*v1alpha1.SubscriptionList, error) {
	ns := c.explicitOrConfigNamespace(options.Namespace)

	list, err := c.eventing.ChannelsV1alpha1().Subscriptions(ns).List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list, nil
}
