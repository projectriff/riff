/*
 * Copyright 2018-2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commands

import (
	"context"
	"fmt"
	"io"
	"os/user"
	"strings"
	"time"

	"github.com/pivotal/go-ape/pkg/furl"
	"github.com/projectriff/riff/pkg/kubectl"

	lcimg "github.com/buildpack/lifecycle/image"
	"github.com/buildpack/pack"
	"github.com/buildpack/pack/cache"
	"github.com/buildpack/pack/docker"
	"github.com/buildpack/pack/logging"
	kbuild "github.com/knative/build/pkg/client/clientset/versioned"
	kserving "github.com/knative/serving/pkg/client/clientset/versioned"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/core/kustomize"
	"github.com/projectriff/riff/pkg/env"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var realClientSetFactory = func(kubeconfig string, masterURL string) (clientcmd.ClientConfig, kubernetes.Interface, kserving.Interface, kbuild.Interface, error) {

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
	servingClientSet, err := kserving.NewForConfig(cfg)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	buildClientSet, err := kbuild.NewForConfig(cfg)

	return clientConfig, kubeClientSet, servingClientSet, buildClientSet, err
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

func CreateAndWireRootCommand(manifests map[string]*core.Manifest) *cobra.Command {

	var client core.Client

	rootCmd := &cobra.Command{
		Use:   env.Cli.Name,
		Short: "Commands for creating and managing function resources",
		Long: `riff is for functions.

` + env.Cli.Name + ` is a CLI for functions on Knative.
See https://projectriff.io and https://github.com/knative/docs`,
		SilenceErrors:              true, // We'll print errors ourselves (after usage rather than before)
		SilenceUsage:               true, // We'll print the *help* message instead of *usage* ourselves
		DisableAutoGenTag:          true,
		SuggestionsMinimumDistance: 2,
	}

	installAdvancedUsage(rootCmd)

	buildpackBuilder := &buildpackBuilder{}
	function := Function()
	installKubeConfigSupport(function, &client)
	function.AddCommand(
		FunctionCreate(buildpackBuilder, &client),
		FunctionUpdate(buildpackBuilder, &client),
		FunctionBuild(buildpackBuilder, &client),
	)

	service := Service()
	installKubeConfigSupport(service, &client)
	service.AddCommand(
		ServiceList(&client),
		ServiceCreate(&client),
		ServiceUpdate(&client),
		ServiceStatus(&client),
		ServiceInvoke(&client),
		ServiceDelete(&client),
	)

	namespace := Namespace()
	installKubeConfigSupport(namespace, &client)
	namespace.AddCommand(
		NamespaceInit(manifests, &client),
		NamespaceCleanup(&client),
	)

	system := System()
	installKubeConfigSupport(system, &client)
	system.AddCommand(
		SystemInstall(manifests, &client),
		SystemUninstall(&client),
	)

	credentials := Credentials()
	installKubeConfigSupport(credentials, &client)
	credentials.AddCommand(
		CredentialsSet(&client),
	)

	rootCmd.AddCommand(
		function,
		service,
		namespace,
		system,
		credentials,
		Docs(rootCmd, LocalFs{}),
		Version(),
		Completion(rootCmd),
	)

	_ = Visit(rootCmd, func(c *cobra.Command) error {
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

// installKubeConfigSupport is to be applied to commands (or parents of commands) that construct a k8s client thanks
// to a kubeconfig configuration. It adds two flags and sets up the PersistentPreRunE function so that it reads
// those configuration files. Hence, when entering the RunE function of the command, the provided clients (passed by
// reference here and to the command creation helpers) are correctly initialized.
func installKubeConfigSupport(command *cobra.Command, client *core.Client) {

	kubeConfigPath := ""
	masterURL := ""

	command.PersistentFlags().StringVar(&kubeConfigPath, "kubeconfig", "~/.kube/config", "the `path` of a kubeconfig")
	command.PersistentFlags().StringVar(&masterURL, "master", "", "the `address` of the Kubernetes API server; overrides any value in kubeconfig")

	oldPersistentPreRunE := command.PersistentPreRunE
	command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		clientConfig, kubeClientSet, servingClientSet, buildClientSet, err := realClientSetFactory(kubeConfigPath, masterURL)
		if err != nil {
			return err
		}

		configPath, err := furl.ResolveTilde(kubeConfigPath)
		if err != nil {
			return err
		}
		kubeCtl := kubectl.RealKubeCtl(configPath, masterURL)

		*client = core.NewClient(clientConfig, kubeClientSet, servingClientSet, buildClientSet, kubeCtl, kustomize.MakeKustomizer(30*time.Second))

		if oldPersistentPreRunE != nil {
			return oldPersistentPreRunE(cmd, args)
		}

		return nil
	}
}

type buildpackBuilder struct{}

func (*buildpackBuilder) Build(repoName string, options core.BuildOptions, log io.Writer) error {
	ctx := context.TODO()
	appDir := options.LocalPath
	builderImage := options.BuildpackImage
	runImage := options.RunImage
	publish := true
	clearCache := false
	outWriter := log
	errWriter := log
	// NOTE below this line is copied directly from github.com/buildpack/pack.Build, once pack offers a proper client we can consume it
	// TODO: Receive Cache as an argument of this function
	dockerClient, err := docker.New()
	if err != nil {
		return err
	}
	c, err := cache.New(repoName, dockerClient)
	if err != nil {
		return err
	}
	imageFactory, err := lcimg.NewFactory(lcimg.WithOutWriter(outWriter))
	if err != nil {
		return err
	}
	imageFetcher := &pack.ImageFetcher{
		Factory: imageFactory,
		Docker:  dockerClient,
	}
	logger := logging.NewLogger(outWriter, errWriter, true, false)
	bf, err := pack.DefaultBuildFactory(logger, c, dockerClient, imageFetcher)
	if err != nil {
		return err
	}
	b, err := bf.BuildConfigFromFlags(ctx,
		&pack.BuildFlags{
			AppDir:     appDir,
			Builder:    builderImage,
			RunImage:   runImage,
			RepoName:   repoName,
			Publish:    publish,
			ClearCache: clearCache,
			// riff: add Env support
			Env: []string{
				fmt.Sprintf("%s=%s", "RIFF", "true"),
				fmt.Sprintf("%s=%s", "RIFF_ARTIFACT", options.Artifact),
				fmt.Sprintf("%s=%s", "RIFF_HANDLER", options.Handler),
				fmt.Sprintf("%s=%s", "RIFF_OVERRIDE", options.Invoker),
			},
			// /riff
		})
	if err != nil {
		return err
	}
	return b.Run(ctx)
}
