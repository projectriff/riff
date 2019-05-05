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
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type Client struct {
	kubeConfig       clientcmd.ClientConfig
	DefaultNamespace string
	RestClient       rest.Interface
	KubeClient       kubernetes.Interface
}

func NewClient(kubeCfgFile string) *Client {
	kubeConfig := getKubeConfig(kubeCfgFile)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		panic(err)
	}

	return &Client{
		kubeConfig:       kubeConfig,
		DefaultNamespace: getDefaultNamespaceOrDie(kubeConfig),
		KubeClient:       kubernetes.NewForConfigOrDie(config),
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

func getRestConfigOrDie(kubeConfig clientcmd.ClientConfig) *rest.Config {
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		panic(err)
	}
	return config
}
