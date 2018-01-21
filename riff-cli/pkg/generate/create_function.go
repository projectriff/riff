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
)

func CreateFunction(workDir, language string, opts options.HandlerAwareInitOptions) error {
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
		//Write Files
	}
	return nil
}
