/*
 * Copyright 2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package k8s

import (
	projectriffclientset "github.com/projectriff/system/pkg/client/clientset/versioned"
	buildv1alpha1 "github.com/projectriff/system/pkg/client/clientset/versioned/typed/build/v1alpha1"
	requestv1alpha1 "github.com/projectriff/system/pkg/client/clientset/versioned/typed/request/v1alpha1"
	streamv1alpha1 "github.com/projectriff/system/pkg/client/clientset/versioned/typed/stream/v1alpha1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type Client interface {
	DefaultNamespace() string
	Core() corev1.CoreV1Interface
	Build() buildv1alpha1.BuildV1alpha1Interface
	Request() requestv1alpha1.RequestV1alpha1Interface
	Stream() streamv1alpha1.StreamV1alpha1Interface
}

func (c *client) DefaultNamespace() string {
	return c.lazyLoadDefaultNamespaceOrDie()
}

func (c *client) Core() corev1.CoreV1Interface {
	return c.lazyLoadKubernetesClientOrDie().CoreV1()
}

func (c *client) Build() buildv1alpha1.BuildV1alpha1Interface {
	return c.lazyLoadRiffClientOrDie().BuildV1alpha1()
}

func (c *client) Request() requestv1alpha1.RequestV1alpha1Interface {
	return c.lazyLoadRiffClientOrDie().RequestV1alpha1()
}

func (c *client) Stream() streamv1alpha1.StreamV1alpha1Interface {
	return c.lazyLoadRiffClientOrDie().StreamV1alpha1()
}

func NewClient(kubeConfigFile string) Client {
	return &client{kubeConfigFile: kubeConfigFile}
}

type client struct {
	defaultNamespace string
	kubeConfigFile   string
	kubeConfig       clientcmd.ClientConfig
	restConfig       *rest.Config
	kubeClient       *kubernetes.Clientset
	riffClient       *projectriffclientset.Clientset
}

func (c *client) lazyLoadKubeConfig() clientcmd.ClientConfig {
	if c.kubeConfig == nil {
		c.kubeConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: c.kubeConfigFile},
			&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}},
		)
	}
	return c.kubeConfig
}

func (c *client) lazyLoadRestConfigOrDie() *rest.Config {
	if c.restConfig == nil {
		kubeConfig := c.lazyLoadKubeConfig()
		restConfig, err := kubeConfig.ClientConfig()
		if err != nil {
			panic(err)
		}
		c.restConfig = restConfig
	}
	return c.restConfig
}

func (c *client) lazyLoadKubernetesClientOrDie() *kubernetes.Clientset {
	if c.kubeClient == nil {
		restConfig := c.lazyLoadRestConfigOrDie()
		c.kubeClient = kubernetes.NewForConfigOrDie(restConfig)
	}
	return c.kubeClient
}

func (c *client) lazyLoadRiffClientOrDie() *projectriffclientset.Clientset {
	if c.riffClient == nil {
		restConfig := c.lazyLoadRestConfigOrDie()
		c.riffClient = projectriffclientset.NewForConfigOrDie(restConfig)
	}
	return c.riffClient
}

func (c *client) lazyLoadDefaultNamespaceOrDie() string {
	if c.defaultNamespace == "" {
		kubeConfig := c.lazyLoadKubeConfig()
		namespace, _, err := kubeConfig.Namespace()
		if err != nil {
			panic(err)
		}
		c.defaultNamespace = namespace
	}
	return c.defaultNamespace
}
