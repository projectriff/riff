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

package commands

import (
	"fmt"
	"os/user"
	"strings"

	eventing "github.com/knative/eventing/pkg/client/clientset/versioned"
	serving "github.com/knative/serving/pkg/client/clientset/versioned"
	"github.com/pivotal-cf-experimental/riff-cli/pkg/tool"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type ClientSetFactory func(kubeconfig string, masterURL string) (*rest.Config, eventing.Interface, serving.Interface, error)

var realClientSetFactory = func(kubeconfig string, masterURL string) (clientcmd.ClientConfig, eventing.Interface, serving.Interface, error) {

	kubeconfig, err := resolveHomePath(kubeconfig)
	if err != nil {
		return nil, nil, nil, err
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: masterURL}})

	cfg, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, nil, nil, err
	}
	eventingClientSet, err := eventing.NewForConfig(cfg)
	if err != nil {
		return nil, nil, nil, err
	}
	servingClientSet, err := serving.NewForConfig(cfg)

	return clientConfig, eventingClientSet, servingClientSet, err
}

func resolveHomePath(p string) (string, error) {
	if strings.HasPrefix(p, "~/") {
		u, err := user.Current()
		if err != nil {
			return "", err
		}
		home := u.HomeDir
		if home == "" {
			return "", fmt.Errorf("could not resolve user home")
		}
		return strings.Replace(p, "~/", home+"/", 1), nil
	} else {
		return p, nil
	}

}

func CreateAndWireRootCommand() *cobra.Command {

	kubeconfig := ""
	masterURL := ""
	var client tool.Client

	rootCmd := &cobra.Command{
		Use:   "riff",
		Short: "Commands for creating and managing function resources",
		Long: `riff is for functions

the riff tool is used to create and manage function resources for the riff FaaS platform https://projectriff.io/`,
		DisableAutoGenTag:          true,
		SuggestionsMinimumDistance: 2,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			clientConfig, eventingClientSet, servingClientSet, err := realClientSetFactory(kubeconfig, masterURL)
			if err != nil {
				return err
			}
			client = tool.NewClient(clientConfig, eventingClientSet, servingClientSet)
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "~/.kube/config", "`path` to a kubeconfig.")
	rootCmd.PersistentFlags().StringVar(&masterURL, "master", "", "the `address` of the Kubernetes API server. Overrides any value in kubeconfig.")

	function := Function()
	function.AddCommand(
		FunctionCreate(&client),
		FunctionSubscribe(&client),
		FunctionDelete(&client),
	)

	channel := Channel()
	channel.AddCommand(
		ChannelCreate(&client),
	)

	rootCmd.AddCommand(
		function,
		channel,
		Docs(rootCmd),
	)

	return rootCmd
}
