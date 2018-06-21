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

package invokers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	projectriff_v1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1alpha1"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
)

type invokerOperations struct {
	kubeCtl kubectl.KubeCtl
}

func Operations(kubeCtl kubectl.KubeCtl) invokerOperations {
	return invokerOperations{kubeCtl}
}

type ApplyOptions struct {
	Filename string
	Name     string
	Version  string
}

func (operations *invokerOperations) Apply(opts ApplyOptions) (string, error) {
	invokerURLs, err := resolveInvokerURLs(opts.Filename)
	if err != nil {
		return "", err
	}
	if len(invokerURLs) == 0 {
		return "", fmt.Errorf("No invokers found at %s", opts.Filename)
	}
	if len(invokerURLs) > 1 && opts.Name != "" {
		return "", fmt.Errorf("Invoker name can only be set for a single invoker, found %d", len(invokerURLs))
	}
	invokersBytes, err := loadInvokers(invokerURLs)
	buf := bytes.Buffer{}
	for _, invokerBytes := range invokersBytes {
		out, err := applyInvoker(operations.kubeCtl, invokerBytes, opts)
		if err != nil {
			return buf.String(), err
		}
		buf.WriteString(out)
	}
	return buf.String(), nil
}

func applyInvoker(kubeCtl kubectl.KubeCtl, invokerBytes []byte, opts ApplyOptions) (string, error) {
	var invoker = projectriff_v1.Invoker{}
	err := yaml.Unmarshal(invokerBytes, &invoker)
	if err != nil {
		return "", err
	}
	if opts.Name != "" {
		invoker.ObjectMeta.Name = opts.Name
	}
	if opts.Version != "" {
		invoker.Spec.Version = opts.Version
	}

	content, err := yaml.Marshal(invoker)
	if err != nil {
		return "", err
	}

	return kubeCtl.ExecStdin([]string{"apply", "-f", "-"}, &content)
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

func (operations *invokerOperations) Table(args ...string) (string, error) {
	kubeCtlArgs := append([]string{
		"get", "invokers.projectriff.io",
		"--sort-by=metadata.name",
		"-o=custom-columns=INVOKER:.metadata.name,VERSION:.spec.version",
	}, args...)
	return operations.kubeCtl.Exec(kubeCtlArgs)
}

func (operations *invokerOperations) List() ([]projectriff_v1.Invoker, error) {
	if invokerPaths, ok := os.LookupEnv("RIFF_INVOKER_PATHS"); ok {
		return listFromPaths(invokerPaths)
	}
	return listFromKubeCtl(operations.kubeCtl)
}

func listFromPaths(invokerPaths string) ([]projectriff_v1.Invoker, error) {
	invokerURLs, err := resolveInvokerURLs(invokerPaths)
	if err != nil {
		return nil, err
	}
	if len(invokerURLs) == 0 {
		return nil, fmt.Errorf("No invokers found at %s", invokerPaths)
	}
	invokersBytes, err := loadInvokers(invokerURLs)

	invokers := []projectriff_v1.Invoker{}
	for _, bytes := range invokersBytes {
		invoker := projectriff_v1.Invoker{}
		err = yaml.Unmarshal(bytes, &invoker)
		if err != nil {
			return nil, err
		}
		invokers = append(invokers, invoker)
	}
	return invokers, nil
}

func listFromKubeCtl(kubeCtl kubectl.KubeCtl) ([]projectriff_v1.Invoker, error) {
	str, err := kubeCtl.Exec([]string{"get", "Invokers", "-o", "json"})
	if err != nil {
		return nil, err
	}
	var invokerList projectriff_v1.InvokerList
	err = json.Unmarshal([]byte(str), &invokerList)
	if err != nil {
		return nil, err
	}
	return invokerList.Items, nil
}

func (operations *invokerOperations) Delete(name string) (string, error) {
	var kubeCtlArgs = []string{"delete", "invokers.projectriff.io", name}
	return operations.kubeCtl.Exec(kubeCtlArgs)
}

func (operations *invokerOperations) DeleteAll() (string, error) {
	var kubeCtlArgs = []string{"delete", "invokers.projectriff.io", "--all"}
	return operations.kubeCtl.Exec(kubeCtlArgs)
}
