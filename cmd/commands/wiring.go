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
	"github.com/projectriff/riff/pkg/core"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var realClientSetFactory = func(kubeconfig string, masterURL string) (clientcmd.ClientConfig, kubernetes.Interface, eventing.Interface, serving.Interface, error) {

	kubeconfig, err := resolveHomePath(kubeconfig)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: masterURL}})

	cfg, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	kubeClientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	eventingClientSet, err := eventing.NewForConfig(cfg)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	servingClientSet, err := serving.NewForConfig(cfg)

	return clientConfig, kubeClientSet, eventingClientSet, servingClientSet, err
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
	var client core.Client
	var kc core.KubectlClient

	rootCmd := &cobra.Command{
		Use:   "riff",
		Short: "Commands for creating and managing function resources",
		Long: `riff is for functions.

riff is a CLI for functions on Knative.
See https://projectriff.io and https://github.com/knative/docs`,
		SilenceErrors:              true, // We'll print errors ourselves (after usage rather than before)
		DisableAutoGenTag:          true,
		SuggestionsMinimumDistance: 2,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			clientConfig, kubeClientSet, eventingClientSet, servingClientSet, err := realClientSetFactory(kubeconfig, masterURL)
			if err != nil {
				return err
			}
			client = core.NewClient(clientConfig, kubeClientSet, eventingClientSet, servingClientSet)
			kc = core.NewKubectlClient(kubeClientSet)
			return nil
		},
	}

	installAdvancedUsage(rootCmd)

	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "~/.kube/config", "the `path` of a kubeconfig")
	rootCmd.PersistentFlags().StringVar(&masterURL, "master", "", "the `address` of the Kubernetes API server; overrides any value in kubeconfig")

	function := Function()
	function.AddCommand(
		FunctionCreate(&client),
		FunctionBuild(&client),
	)

	service := Service()
	service.AddCommand(
		ServiceList(&client),
		ServiceCreate(&client),
		ServiceStatus(&client),
		ServiceInvoke(&client),
		ServiceSubscribe(&client),
		ServiceUnsubscribe(&client),
		ServiceDelete(&client),
	)

	channel := Channel()
	channel.AddCommand(
		ChannelList(&client),
		ChannelCreate(&client),
		ChannelDelete(&client),
	)

	namespace := Namespace()
	namespace.AddCommand(
		NamespaceInit(&kc),
	)

	system := System()
	system.AddCommand(
		SystemInstall(&kc),
		SystemUninstall(&kc),
	)

	rootCmd.AddCommand(
		function,
		service,
		channel,
		namespace,
		system,
		Docs(rootCmd),
		Version(),
	)

	Visit(rootCmd, func(c *cobra.Command) error {
		// Disable usage printing as soon as we enter RunE(), as errors that happen from then on
		// are not mis-usage error, but "regular" runtime errors
		exec := c.RunE
		if exec != nil {
			c.RunE = func(cmd *cobra.Command, args []string) error {
				c.SilenceUsage = true
				return exec(cmd, args)
			}
		}
		return nil
	})

	return rootCmd
}
