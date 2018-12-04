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
	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CreateSubscriptionOptions struct {
	Namespace  string
	Name       string
	Channel    string
	Subscriber string
	Reply      string
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
	s := v1alpha1.Subscription{
		TypeMeta: meta_v1.TypeMeta{
			APIVersion: "eventing.knative.dev/v1alpha1",
			Kind:       "Subscription",
		},

		ObjectMeta: meta_v1.ObjectMeta{
			Name: options.Name,
		},
		Spec: v1alpha1.SubscriptionSpec{
			Channel: corev1.ObjectReference{
				APIVersion: "eventing.knative.dev/v1alpha1",
				Kind:       "Channel",
				Name:       options.Channel,
			},
			Subscriber: c.makeSubscriptionSubscriber(options.Subscriber),
			Reply:      c.makeSubscriptionReplyStrategy(options.Reply),
		},
	}

	if !options.DryRun {
		_, e := c.eventing.EventingV1alpha1().Subscriptions(ns).Create(&s)
		return &s, e
	} else {
		return &s, nil
	}

}

func (c *client) DeleteSubscription(options DeleteSubscriptionOptions) error {
	ns := c.explicitOrConfigNamespace(options.Namespace)
	return c.eventing.EventingV1alpha1().Subscriptions(ns).Delete(options.Name, nil)
}

func (c *client) ListSubscriptions(options ListSubscriptionsOptions) (*v1alpha1.SubscriptionList, error) {
	ns := c.explicitOrConfigNamespace(options.Namespace)

	list, err := c.eventing.EventingV1alpha1().Subscriptions(ns).List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (c *client) makeSubscriptionSubscriber(subscriber string) *v1alpha1.SubscriberSpec {
	if subscriber == "" {
		return nil
	}
	// TODO add support for DNSName as alternative to Ref
	return &v1alpha1.SubscriberSpec{
		Ref: &corev1.ObjectReference{
			APIVersion: "serving.knative.dev/v1alpha1",
			Kind:       "Service",
			Name:       subscriber,
		},
	}
}

func (c *client) makeSubscriptionReplyStrategy(reply string) *v1alpha1.ReplyStrategy {
	if reply == "" {
		return nil
	}
	return &v1alpha1.ReplyStrategy{
		Channel: &corev1.ObjectReference{
			APIVersion: "eventing.knative.dev/v1alpha1",
			Kind:       "Channel",
			Name:       reply,
		},
	}
}
