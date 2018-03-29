/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/ghodss/yaml"
	projectriff_v1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1"
	"github.com/projectriff/riff/riff-cli/cmd/utils"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
	"github.com/spf13/cobra"
)

func Invokers() *cobra.Command {

	var invokersCmd = &cobra.Command{
		Use:   "invokers",
		Short: "Manage invokers in the cluster",
	}

	return invokersCmd
}

type InvokersApplyOptions struct {
	Filename string
	Name     string
	Version  string
}

func InvokersApply(kubeCtl kubectl.KubeCtl) (*cobra.Command, *InvokersApplyOptions) {

	var invokersApplyOptions = InvokersApplyOptions{}

	var invokersApplyCmd = &cobra.Command{
		Use:   "apply",
		Short: "Install or update an invoker in the cluster",
		Args:  utils.AliasFlagToSoleArg("filename"),
		RunE: func(cmd *cobra.Command, args []string) error {
			invokerURLs, err := resolveInvokerURLs(invokersApplyOptions.Filename)
			if err != nil {
				return err
			}
			if len(invokerURLs) == 0 {
				return fmt.Errorf("No invokers found at %s", invokersApplyOptions.Filename)
			}
			if len(invokerURLs) > 1 && invokersApplyOptions.Name != "" {
				return fmt.Errorf("Invoker name can only be set for a single invoker, found %d", len(invokerURLs))
			}
			invokersBytes, err := loadInvokers(invokerURLs)
			for _, invokerBytes := range invokersBytes {
				var invoker = projectriff_v1.Invoker{}
				err = yaml.Unmarshal(invokerBytes, &invoker)
				if err != nil {
					return err
				}
				if invokersApplyOptions.Name != "" {
					invoker.ObjectMeta.Name = invokersApplyOptions.Name
				}
				if invokersApplyOptions.Version != "" {
					invoker.Spec.Version = invokersApplyOptions.Version
				}

				content, err := yaml.Marshal(invoker)
				if err != nil {
					return err
				}

				out, err := kubeCtl.ExecStdin([]string{"apply", "-f", "-"}, &content)
				if err != nil {
					return err
				}
				fmt.Print(out)
			}
			return nil
		},
	}

	invokersApplyCmd.Flags().StringVarP(&invokersApplyOptions.Filename, "filename", "f", ".", "path to the invoker resource to install")
	invokersApplyCmd.Flags().StringVarP(&invokersApplyOptions.Name, "name", "n", "", "name of the invoker (defaults to the name in the invoker resource)")
	invokersApplyCmd.Flags().StringVarP(&invokersApplyOptions.Version, "version", "v", "", "version of the invoker (defaults to the version in the invoker resource)")

	return invokersApplyCmd, &invokersApplyOptions
}

func InvokersList(kubeCtl kubectl.KubeCtl) *cobra.Command {

	var invokersListCmd = &cobra.Command{
		Use:   "list",
		Short: "List invokers in the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeCtlArgs := append([]string{
				"get", "invokers.projectriff.io",
				"--sort-by=metadata.name",
				"-o=custom-columns=NAME:.metadata.name,VERSION:.spec.version",
			}, args...)
			listing, err := kubeCtl.Exec(kubeCtlArgs)
			if err != nil {
				return err
			}
			fmt.Print(listing)
			return nil
		},
	}

	return invokersListCmd
}

type InvokersDeleteOptions struct {
	All  bool
	Name string
}

func InvokersDelete(kubeCtl kubectl.KubeCtl) (*cobra.Command, *InvokersDeleteOptions) {

	var invokersDeleteOptions = InvokersDeleteOptions{}

	var invokersDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Remove an invoker from the cluster",
		Args:  utils.AliasFlagToSoleArg("name"),
		RunE: func(cmd *cobra.Command, args []string) error {
			var kubeCtlArgs = []string{"delete", "invokers.projectriff.io"}
			if invokersDeleteOptions.All {
				kubeCtlArgs = append(kubeCtlArgs, "--all")
			} else if invokersDeleteOptions.Name == "" {
				return fmt.Errorf("Invoker to delete must be specified")
			} else {
				kubeCtlArgs = append(kubeCtlArgs, invokersDeleteOptions.Name)
			}
			out, err := kubeCtl.Exec(kubeCtlArgs)
			if err != nil {
				return err
			}
			fmt.Print(out)
			return nil
		},
	}

	invokersDeleteCmd.Flags().BoolVar(&invokersDeleteOptions.All, "all", false, "remove all invokers from the cluster")
	invokersDeleteCmd.Flags().StringVarP(&invokersDeleteOptions.Name, "name", "n", "", "invoker name to remove from the cluster")

	return invokersDeleteCmd, &invokersDeleteOptions
}

func loadInvokers(urls []url.URL) ([][]byte, error) {
	var invokersBytes = [][]byte{}
	for _, url := range urls {
		if url.Scheme == "file" {
			file, err := ioutil.ReadFile(url.Path)
			if err != nil {
				return nil, err
			}
			invokersBytes = append(invokersBytes, file)
		} else if url.Scheme == "http" || url.Scheme == "https" {
			resp, err := http.Get(url.String())
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			invokersBytes = append(invokersBytes, body)
		} else {
			return nil, fmt.Errorf("Filename must be file, http or https, got %s", url.Scheme)
		}
	}
	return invokersBytes, nil
}

func resolveInvokerURLs(filename string) ([]url.URL, error) {
	u, err := url.Parse(filename)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "file"
	}
	if u.Scheme == "http" || u.Scheme == "https" {
		return []url.URL{*u}, nil
	}
	if u.Scheme == "file" {
		if osutils.IsDirectory(u.Path) {
			u.Path = filepath.Join(u.Path, "*-invoker.yaml")
		}
		filenames, err := filepath.Glob(u.Path)
		if err != nil {
			return nil, err
		}
		var urls = []url.URL{}
		for _, f := range filenames {
			urls = append(urls, url.URL{
				Scheme: u.Scheme,
				Path:   f,
			})
		}
		return urls, nil
	}
	return nil, fmt.Errorf("Filename must be file, http or https, got %s", u.Scheme)
}
