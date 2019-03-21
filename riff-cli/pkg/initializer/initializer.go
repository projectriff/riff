/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *        https://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package initializer

import (
	"fmt"
	"path/filepath"
	"strings"

	projectriff_v1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1alpha1"
	"github.com/projectriff/riff/riff-cli/pkg/options"
	"github.com/projectriff/riff/riff-cli/pkg/templateutils"
)

func Initialize(invoker projectriff_v1.Invoker, opts *options.InitOptions) error {
	err := resolveOptions(opts, invoker)
	if err != nil {
		return err
	}

	err = generateResources(invoker, opts)
	if err != nil {
		return err
	}

	return nil
}

type handlerOptions struct {
	FunctionName string
}

func (h handlerOptions) TitleCase(s string) string {
	return strings.Title(s)
}

func resolveOptions(opts *options.InitOptions, invoker projectriff_v1.Invoker) error {
	if opts.Input == "" {
		opts.Input = opts.FunctionName
	}

	if opts.InvokerVersion == "" {
		opts.InvokerVersion = invoker.Spec.Version
	}

	if opts.Artifact == "" {
		workdir, err := filepath.Abs(opts.FilePath)
		if err != nil {
			return err
		}
		artifacts, err := resolveArtifacts(workdir, invoker)
		if err != nil {
			return err
		}

		if len(artifacts) == 0 {
			return fmt.Errorf("No matching artifact found")
		}
		if len(artifacts) > 1 {
			// TODO attempt to find the "best" artifact
			return fmt.Errorf("Artifact must be specified")
		}

		relativePath, err := filepath.Rel(workdir, artifacts[0])
		if err != nil {
			return err
		}
		opts.Artifact = relativePath
	}

	if opts.Handler != "" {
		handler, err := templateutils.Apply(opts.Handler, "opts.Handler", handlerOptions{FunctionName: opts.FunctionName})
		if err != nil {
			return err
		}
		opts.Handler = handler
	}

	return nil
}

func resolveArtifacts(workdir string, invoker projectriff_v1.Invoker) ([]string, error) {
	artifacts := make(map[string]bool)
	for _, matcher := range invoker.Spec.Matchers {
		matches, err := filepath.Glob(filepath.Join(workdir, matcher))
		if err != nil {
			return []string{}, nil
		}
		for _, match := range matches {
			artifacts[match] = true
		}
	}
	keys := make([]string, 0, len(artifacts))
	for k := range artifacts {
		keys = append(keys, k)
	}
	return keys, nil
}
