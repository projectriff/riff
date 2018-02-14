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

package core

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/projectriff/riff-cli/pkg/options"
	"github.com/projectriff/riff-cli/pkg/osutils"
)

const (
	ApiVersion = "projectriff.io/v1"
)

type FunctionResources struct {
	Topics       string
	Function     string
	DockerFile   string
	DockerIgnore string
}

type Function struct {
	ApiVersion string
	Name       string
	Input      string
	Output     string
	Image      string
	Protocol   string
}

type ArtifactsGenerator struct {
	GenerateFunction     func(options.InitOptions) (string, error)
	GenerateDockerFile   func(options.InitOptions) (string, error)
	GenerateDockerIgnore func(options.InitOptions) (string, error)
}

func GenerateFunctionArtfacts(generator ArtifactsGenerator, workdir string, opts options.InitOptions) error {
	var functionResources FunctionResources
	var err error
	functionResources.Topics, err = createTopics(opts)
	if err != nil {
		return err
	}
	functionResources.Function, err = generator.GenerateFunction(opts)
	if err != nil {
		return err
	}
	functionResources.DockerFile, err = generator.GenerateDockerFile(opts)
	if err != nil {
		return err
	}
	if generator.GenerateDockerIgnore != nil {
		// optionally generate .dockerignore
		functionResources.DockerIgnore, err = generator.GenerateDockerIgnore(opts)
		if err != nil {
			return err
		}
	}

	if opts.DryRun {
		fmt.Printf("%s-%s.yaml\n", opts.FunctionName, "topics")
		fmt.Print("----")
		fmt.Printf("%s", functionResources.Topics)
		fmt.Print("----\n")
		fmt.Printf("\n%s-%s.yaml\n", opts.FunctionName, "function")
		fmt.Print("----")
		fmt.Printf("%s", functionResources.Function)
		fmt.Print("----\n")
		fmt.Print("\nDockerfile\n")
		fmt.Print("----")
		fmt.Printf("%s", functionResources.DockerFile)
		fmt.Print("----\n")
		if generator.GenerateDockerIgnore != nil {
			fmt.Print("\n.dockerignore\n")
			fmt.Print("----")
			fmt.Printf("%s", functionResources.DockerIgnore)
			fmt.Print("----\n")
		}
		fmt.Println("")
	} else {
		var err error
		err = writeFile(
			filepath.Join(workdir,
				fmt.Sprintf("%s-%s.yaml", opts.FunctionName, "topics")),
			functionResources.Topics,
			opts.Force)
		if err != nil {
			return err
		}

		err = writeFile(
			filepath.Join(workdir,
				fmt.Sprintf("%s-%s.yaml", opts.FunctionName, "function")),
			functionResources.Function,
			opts.Force)
		if err != nil {
			return err
		}

		err = writeFile(
			filepath.Join(workdir, "Dockerfile"),
			functionResources.DockerFile,
			opts.Force)
		if err != nil {
			return err
		}

		if generator.GenerateDockerIgnore != nil {
			err = writeFile(
				filepath.Join(workdir, ".dockerignore"),
				functionResources.DockerIgnore,
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
