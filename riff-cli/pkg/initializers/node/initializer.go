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

package node

import (
	"path/filepath"

	"github.com/projectriff/riff-cli/pkg/initializers/core"
	"github.com/projectriff/riff-cli/pkg/initializers/utils"
	"github.com/projectriff/riff-cli/pkg/options"
)

const (
	language  = "node"
	extension = "js"
)

func Initialize(opts options.InitOptions) error {
	functionfile, err := utils.ResolveFunctionFile(opts, language, extension)
	if err != nil {
		return err
	}
	utils.ResolveOptions(functionfile, language, &opts)

	workdir := filepath.Dir(functionfile)

	generator := core.ArtifactsGenerator{
		GenerateFunction:     core.DefaultGenerateFunction,
		GenerateDockerFile:   generateNodeFunctionDockerFile,
		GenerateDockerIgnore: generateNodeFunctionDockerIgnore,
	}

	return core.GenerateFunctionArtifacts(generator, workdir, opts)
}
