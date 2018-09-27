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

type ListChannelOptions struct {
	Namespace string
}

func (c *client) ListChannels(options ListChannelOptions) (*v1alpha1.ChannelList, error) {
	ns := c.explicitOrConfigNamespace(options.Namespace)
	return c.eventing.ChannelsV1alpha1().Channels(ns).List(meta_v1.ListOptions{})
}

type CreateChannelOptions struct {
	Namespace  string
	Name       string
	Bus        string
	ClusterBus string
	DryRun     bool
}

func (c *client) CreateChannel(options CreateChannelOptions) (*v1alpha1.Channel, error) {
	ns := c.explicitOrConfigNamespace(options.Namespace)
	channel := v1alpha1.Channel{
		TypeMeta: meta_v1.TypeMeta{
			APIVersion: "channels.knative.dev/v1alpha1",
			Kind:       "Channel",
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name: options.Name,
		},
		Spec: v1alpha1.ChannelSpec{
			ClusterBus: options.ClusterBus,
			Bus:        options.Bus,
		},
	}

	if !options.DryRun {
		_, err := c.eventing.ChannelsV1alpha1().Channels(ns).Create(&channel)
		return &channel, err
	} else {
		return &channel, nil
	}
}

type DeleteChannelOptions struct {
	Namespace string
	Name      string
}

func (c *client) DeleteChannel(options DeleteChannelOptions) error {
	ns := c.explicitOrConfigNamespace(options.Namespace)

	err := c.eventing.ChannelsV1alpha1().Channels(ns).Delete(options.Name, nil)

	return err
}
