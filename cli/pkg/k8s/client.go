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
	corev1alpha1 "github.com/projectriff/system/pkg/client/clientset/versioned/typed/core/v1alpha1"
	knativev1alpha1 "github.com/projectriff/system/pkg/client/clientset/versioned/typed/knative/v1alpha1"
	streamv1alpha1 "github.com/projectriff/system/pkg/client/clientset/versioned/typed/streaming/v1alpha1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	authv1client "k8s.io/client-go/kubernetes/typed/authorization/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client interface {
	DefaultNamespace() string
	KubeRestConfig() *rest.Config
	Core() corev1.CoreV1Interface
	Auth() authv1client.AuthorizationV1Interface
	APIExtension() apiextensionsv1beta1.ApiextensionsV1beta1Interface
	Build() buildv1alpha1.BuildV1alpha1Interface
	CoreRuntime() corev1alpha1.CoreV1alpha1Interface
	StreamingRuntime() streamv1alpha1.StreamingV1alpha1Interface
	KnativeRuntime() knativev1alpha1.KnativeV1alpha1Interface
}

func (c *client) DefaultNamespace() string {
	return c.lazyLoadDefaultNamespaceOrDie()
}

func (c *client) KubeRestConfig() *rest.Config {
	return c.lazyLoadRestConfigOrDie()
}

func (c *client) Core() corev1.CoreV1Interface {
	return c.lazyLoadKubernetesClientsetOrDie().CoreV1()
}

func (c *client) Auth() authv1client.AuthorizationV1Interface {
	return c.lazyLoadKubernetesClientsetOrDie().AuthorizationV1()
}

func (c *client) APIExtension() apiextensionsv1beta1.ApiextensionsV1beta1Interface {
	return c.lazyLoadAPIExtensionsClientsetOrDie().ApiextensionsV1beta1()
}

func (c *client) Build() buildv1alpha1.BuildV1alpha1Interface {
	return c.lazyLoadRiffClientsetOrDie().BuildV1alpha1()
}

func (c *client) CoreRuntime() corev1alpha1.CoreV1alpha1Interface {
	return c.lazyLoadRiffClientsetOrDie().CoreV1alpha1()
}

func (c *client) StreamingRuntime() streamv1alpha1.StreamingV1alpha1Interface {
	return c.lazyLoadRiffClientsetOrDie().StreamingV1alpha1()
}

func (c *client) KnativeRuntime() knativev1alpha1.KnativeV1alpha1Interface {
	return c.lazyLoadRiffClientsetOrDie().KnativeV1alpha1()
}

func NewClient(kubeConfigFile string) Client {
	return &client{kubeConfigFile: kubeConfigFile}
}

type client struct {
	defaultNamespace       string
	kubeConfigFile         string
	kubeConfig             clientcmd.ClientConfig
	restConfig             *rest.Config
	kubeClientset          *kubernetes.Clientset
	apiExtensionsClientset *apiextensionsclientset.Clientset
	riffClientset          *projectriffclientset.Clientset
}

func (c *client) lazyLoadKubeConfig() clientcmd.ClientConfig {
	if c.kubeConfig == nil {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		loadingRules.ExplicitPath = c.kubeConfigFile
		configOverrides := &clientcmd.ConfigOverrides{}
		c.kubeConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
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

func (c *client) lazyLoadKubernetesClientsetOrDie() *kubernetes.Clientset {
	if c.kubeClientset == nil {
		restConfig := c.lazyLoadRestConfigOrDie()
		c.kubeClientset = kubernetes.NewForConfigOrDie(restConfig)
	}
	return c.kubeClientset
}

func (c *client) lazyLoadAPIExtensionsClientsetOrDie() *apiextensionsclientset.Clientset {
	if c.apiExtensionsClientset == nil {
		restConfig := c.lazyLoadRestConfigOrDie()
		c.apiExtensionsClientset = apiextensionsclientset.NewForConfigOrDie(restConfig)
	}
	return c.apiExtensionsClientset

}

func (c *client) lazyLoadRiffClientsetOrDie() *projectriffclientset.Clientset {
	if c.riffClientset == nil {
		restConfig := c.lazyLoadRestConfigOrDie()
		c.riffClientset = projectriffclientset.NewForConfigOrDie(restConfig)
	}
	return c.riffClientset
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
