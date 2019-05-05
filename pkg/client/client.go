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

package client

import (
	projectriffclientset "github.com/projectriff/system/pkg/client/clientset/versioned"
	buildv1alpha1 "github.com/projectriff/system/pkg/client/clientset/versioned/typed/build/v1alpha1"
	runv1alpha1 "github.com/projectriff/system/pkg/client/clientset/versioned/typed/run/v1alpha1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type Client struct {
	DefaultNamespace string
	Core             corev1.CoreV1Interface
	Build            buildv1alpha1.BuildV1alpha1Interface
	Run              runv1alpha1.RunV1alpha1Interface
}

func NewClient(kubeCfgFile string) *Client {
	kubeConfig := getKubeConfig(kubeCfgFile)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		panic(err)
	}

	kubeClient := kubernetes.NewForConfigOrDie(config)
	riffClient := projectriffclientset.NewForConfigOrDie(config)

	return &Client{
		DefaultNamespace: getDefaultNamespaceOrDie(kubeConfig),
		Core:             kubeClient.CoreV1(),
		Build:            riffClient.BuildV1alpha1(),
		Run:              riffClient.RunV1alpha1(),
	}
}

func getKubeConfig(kubeCfgFile string) clientcmd.ClientConfig {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeCfgFile},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}},
	)
}

func getDefaultNamespaceOrDie(kubeConfig clientcmd.ClientConfig) string {
	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		panic(err)
	}
	return namespace
}
