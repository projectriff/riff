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

package generate

import (
	"fmt"
	"github.com/projectriff/riff-cli/pkg/options"
	"io/ioutil"
	"path/filepath"
	"strings"
	"github.com/projectriff/riff-cli/pkg/osutils"
	"errors"
)

func CreateFunction(workdir, language string, opts options.HandlerAwareInitOptions) error {
	var functionResources FunctionResources
	var err error
	functionResources.Topics, err = createTopics(opts.InitOptions)
	if err != nil {
		return err
	}
	functionResources.Function, err = createFunction(opts.InitOptions)
	if err != nil {
		return err
	}
	functionResources.DockerFile, err = generateDockerfile(language,opts)
	if err != nil {
		return err
	}

	if (opts.DryRun) {
		fmt.Println("Generated Topics:\n")
		fmt.Printf("%s\n",functionResources.Topics)
		fmt.Println("\nGenerated Function:\n")
		fmt.Printf("%s\n",functionResources.Function)
		fmt.Println("\nGenerated Dockerfile:\n")
		fmt.Printf("%s\n",functionResources.DockerFile)
	} else {
		var err error
		err = writeFile(
				filepath.Join(workdir,
				fmt.Sprintf("%s-%s.yaml",opts.FunctionName,"topics")),
				functionResources.Topics)
		if err != nil {
			return err
		}

		err = writeFile(
			filepath.Join(workdir,
				fmt.Sprintf("%s-%s.yaml",opts.FunctionName,"topics")),
			functionResources.Topics)
		if err != nil {
			return err
		}

		err = writeFile(
			filepath.Join(workdir, "Dockerfile"),
			functionResources.DockerFile)
		if err != nil {
			return err
		}


	}
	return nil
}

func writeFile(filename string, text string) error {
	overwrite := true

	if !overwrite && osutils.FileExists(filename) {
		return errors.New(fmt.Sprintf("file %s already exists", filename))

	}
	//fmt.Printf("creating %s\n", filename)
	return ioutil.WriteFile(filename,[]byte(strings.TrimLeft(text,"\n")), 0644)

}
