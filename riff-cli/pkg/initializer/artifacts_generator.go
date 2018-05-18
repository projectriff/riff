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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	projectriff_v1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1alpha1"
	"github.com/projectriff/riff/riff-cli/pkg/options"
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
)

const (
	apiVersion   = "projectriff.io/v1alpha1"
	functionKind = "Function"
	topicKind    = "Topic"
	linkKind     = "Link"
)

type resource struct {
	Path    string
	Content string
}

func generateResources(invoker projectriff_v1.Invoker, opts *options.InitOptions) error {
	var resources []resource

	// {FunctionName}-topics.yaml
	content, err := createTopicsYaml(invoker.Spec.TopicTemplate, *opts)
	if err != nil {
		return err
	}
	resources = append(resources, resource{
		Path:    fmt.Sprintf("%s-topics.yaml", opts.FunctionName),
		Content: content,
	})

	// {FunctionName}-function.yaml
	content, err = createFunctionYaml(invoker.Spec.FunctionTemplate, *opts)
	if err != nil {
		return err
	}
	resources = append(resources, resource{
		Path:    fmt.Sprintf("%s-function.yaml", opts.FunctionName),
		Content: content,
	})

	// {FunctionName}-link.yaml
	content, err = createLinkYaml(invoker.Spec.LinkTemplate, *opts)
	if err != nil {
		return err
	}
	resources = append(resources, resource{
		Path:    fmt.Sprintf("%s-link.yaml", opts.FunctionName),
		Content: content,
	})

	// Invoker defined files
	for _, file := range invoker.Spec.Files {
		content, err = generateFileContents(file.Template, file.Path, *opts)
		if err != nil {
			return err
		}
		resources = append(resources, resource{
			Path:    file.Path,
			Content: content,
		})
	}

	if opts.DryRun {
		delim := "----"
		for _, resource := range resources {
			fmt.Println(delim)
			fmt.Println(resource.Path)
			fmt.Println(delim)
			fmt.Println(resource.Content)
		}
		fmt.Println(delim)
	} else {
		workdir, err := filepath.Abs(opts.FilePath)
		if err != nil {
			return err
		}
		for _, resource := range resources {
			err = writeFile(
				filepath.Join(workdir, resource.Path),
				resource.Content,
				opts.Force)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func writeFile(filename string, text string, overwrite bool) error {
	if !overwrite && osutils.FileExists(filename) {
		fmt.Printf("Skipping existing file %s  - set --force to overwrite.\n", filename)
		return nil
	} else {
		fmt.Printf("Initializing %s\n", filename)
		return ioutil.WriteFile(filename, []byte(strings.TrimLeft(text, "\n")), 0644)
	}
}
