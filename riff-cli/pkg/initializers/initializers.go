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

package initializers

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/projectriff/riff-cli/pkg/initializers/java"
	"github.com/projectriff/riff-cli/pkg/initializers/node"
	"github.com/projectriff/riff-cli/pkg/initializers/python"
	"github.com/projectriff/riff-cli/pkg/initializers/shell"
	"github.com/projectriff/riff-cli/pkg/initializers/utils"
	"github.com/projectriff/riff-cli/pkg/options"
)

var supportedExtensions = []string{"js", "json", "jar", "py", "sh"}

type Initializer struct {
	Initialize func(options.InitOptions) error
}

var languageForFileExtension = map[string]string{
	"sh":   "shell",
	"jar":  "java",
	"js":   "node",
	"json": "node",
	"py":   "python",
}

func Java() Initializer {
	return Initializer{
		Initialize: java.Initialize,
	}
}

func Python() Initializer {
	return Initializer{
		Initialize: python.Initialize,
	}
}
func Node() Initializer {
	return Initializer{
		Initialize: node.Initialize,
	}
}
func Shell() Initializer {
	return Initializer{
		Initialize: shell.Initialize,
	}
}

func Initialize(opts options.InitOptions) error {
	filePath, err := utils.ResolveFunctionFile(opts, "", "")
	if err != nil {
		return err
	}

	language := languageForFileExtension[filepath.Ext(filePath)[1:]]

	switch language {
	case "shell":
		Shell().Initialize(opts)
	case "node":
		Node().Initialize(opts)
	case "java":
		return errors.New("Java resources detected. Use 'riff init java' to specify additional required options")
	case "python":
		return errors.New("Python resources detected. Use 'riff init python' to specify additional required options")
	default:
		//TODO: Should never get here
		return errors.New(fmt.Sprintf("unsupported language %s\n", language))
	}
	return nil
}
