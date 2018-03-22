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

package initializer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	projectriff_v1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/projectriff/riff/riff-cli/pkg/options"
	"github.com/projectriff/riff/riff-cli/pkg/templateutils"
)

func Initialize(invokers []projectriff_v1.Invoker, opts *options.InitOptions) error {
	invoker, err := resolveInvoker(invokers, opts)
	if err != nil {
		return err
	}

	err = resolveOptions(opts, invoker)
	if err != nil {
		return err
	}

	err = generateFunctionArtifacts(invoker, opts)
	if err != nil {
		return err
	}

	return nil
}

func LoadInvokers(kubeCtl kubectl.KubeCtl) ([]projectriff_v1.Invoker, error) {
	if _, disableKubeCtl := os.LookupEnv("RIFF_DISABLE_KUBECTL"); disableKubeCtl {
		fmt.Println("No invoker lookup due to RIFF_DISABLE_KUBECTL")
		return []projectriff_v1.Invoker{}, nil
	}

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

func resolveInvoker(invokers []projectriff_v1.Invoker, opts *options.InitOptions) (projectriff_v1.Invoker, error) {
	var resolvedInvoker projectriff_v1.Invoker

	// look for an exact invoker
	if opts.InvokerName != "" {
		for _, invoker := range invokers {
			if opts.InvokerName == invoker.ObjectMeta.Name {
				resolvedInvoker = invoker
			}
		}
		if resolvedInvoker.ObjectMeta.Name == "" {
			return projectriff_v1.Invoker{}, fmt.Errorf("Invoker %s not found", opts.InvokerName)
		}
	}

	if opts.Artifact == "" {
		// look for a matching artifact

		// This will get slower as more invokers are introduced, more complex
		// matching patterns are used and run in directory with more files.
		// Considering the search is non-deterministic between calls if the
		// invokers are updated, it may not be worth the effort. Forcing the
		// user to specify the artifact will produce stable results

		workdir, err := filepath.Abs(opts.FilePath)
		if err != nil {
			return projectriff_v1.Invoker{}, err
		}
		artifacts, err := resolveArtifacts(workdir, invokers)
		if err != nil {
			return projectriff_v1.Invoker{}, err
		}

		if len(artifacts) == 0 {
			var registeredInvokers []string
			for _, element := range invokers {
				registeredInvokers = append(registeredInvokers, element.Name)
			}
			return projectriff_v1.Invoker{}, fmt.Errorf("No matching artifact found (using registered invokers: %v)", registeredInvokers)
		}
		if len(artifacts) > 1 {
			// TODO MAYBE attempt to find the "best" artifact
			return projectriff_v1.Invoker{}, fmt.Errorf("Artifact must be specified")
		}

		relativePath, err := filepath.Rel(workdir, artifacts[0])
		if err != nil {
			return projectriff_v1.Invoker{}, err
		}
		opts.Artifact = relativePath
	}

	if resolvedInvoker.ObjectMeta.Name != "" {
		return resolvedInvoker, nil
	}

	// look for a matching invoker
	var matchingInvokers []projectriff_v1.Invoker
	for _, invoker := range invokers {
		matched := false
		for _, matcher := range invoker.Spec.Matchers {
			if matched {
				continue
			}
			match, err := filepath.Match(matcher, opts.Artifact)
			if err != nil {
				return projectriff_v1.Invoker{}, err
			}
			if match {
				matchingInvokers = append(matchingInvokers, invoker)
				matched = true
			}
		}
	}
	if len(matchingInvokers) > 1 {
		// TODO MAYBE attempt to find a clear "best" match
		var names []string
		for _, matchingInvoker := range matchingInvokers {
			names = append(names, matchingInvoker.ObjectMeta.Name)
		}
		return projectriff_v1.Invoker{}, fmt.Errorf("Multiple matching invokers found, pick one of: %s: ", strings.Join(names, ", "))
	}
	if len(matchingInvokers) == 0 {
		return projectriff_v1.Invoker{}, fmt.Errorf("No invoker found matching %s", opts.Artifact)
	}
	return matchingInvokers[0], nil
}

func resolveArtifacts(workdir string, invokers []projectriff_v1.Invoker) ([]string, error) {
	artifacts := make(map[string]bool)
	for _, invoker := range invokers {
		for _, matcher := range invoker.Spec.Matchers {
			matches, err := filepath.Glob(filepath.Join(workdir, matcher))
			if err != nil {
				return []string{}, nil
			}
			for _, match := range matches {
				artifacts[match] = true
			}
		}
	}
	keys := make([]string, 0, len(artifacts))
	for k := range artifacts {
		keys = append(keys, k)
	}
	return keys, nil
}

func resolveOptions(opts *options.InitOptions, invoker projectriff_v1.Invoker) error {
	if opts.Input == "" {
		opts.Input = opts.FunctionName
	}

	if opts.InvokerVersion == "" {
		opts.InvokerVersion = invoker.Spec.Version
	}

	// if opts.Artifact == "" {
	// 	opts.Artifact = filepath.Base(functionArtifact)
	// }

	if opts.Handler != "" {
		handler, err := templateutils.Apply(opts.Handler, "opts.Handler", opts)
		if err != nil {
			return err
		}
		opts.Handler = handler
	}

	return nil
}
