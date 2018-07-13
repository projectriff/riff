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
	"k8s.io/client-go/tools/clientcmd"
)

type EventingInterfaceFactory func(kubeconfig string, masterURL string) (eventing.Interface, error)

type ServingInterfaceFactory func(kubeconfig string, masterURL string) (serving.Interface, error)

var realChannelsInterfaceFactory = func(kubeconfig string, masterURL string) (eventing.Interface, error) {

	kubeconfig, err := resolveHomePath(kubeconfig)
	if err != nil {
		return nil, err
	}
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		return nil, err
	}
	clientset, err := eventing.NewForConfig(cfg)

	return clientset, err
}

var realServingInterfaceFactory = func(kubeconfig string, masterURL string) (serving.Interface, error) {

	kubeconfig, err := resolveHomePath(kubeconfig)
	if err != nil {
		return nil, err
	}
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		return nil, err
	}
	clientset, err := serving.NewForConfig(cfg)

	return clientset, err
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
			eventingClientSet, err := realChannelsInterfaceFactory(kubeconfig, masterURL)
			if err != nil {
				return err
			}
			servingClientSet, err := realServingInterfaceFactory(kubeconfig, masterURL)
			if err != nil {
				return err
			}
			client = tool.NewClient(eventingClientSet, servingClientSet)
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "~/.kube/config", "`path` to a kubeconfig.")
	rootCmd.PersistentFlags().StringVar(&masterURL, "master", "", "the `address` of the Kubernetes API server. Overrides any value in kubeconfig.")

	function := Function()
	function.AddCommand(
		FunctionCreate(&client),
		FunctionSubscribe(&client),
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
