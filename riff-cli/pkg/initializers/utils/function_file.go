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

package utils

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/projectriff/riff-cli/pkg/options"
	"github.com/projectriff/riff-cli/pkg/osutils"
)

var supportedExtensions = []string{"js", "json", "java", "py", "sh"}

var languageForFileExtensions = map[string]string{
	"sh":   "shell",
	"jar":  "java",
	"js":   "node",
	"json": "node",
	"py":   "python",
}

//Assumes given file paths have been sanity checked and are valid
func ResolveFunctionFile(opts options.InitOptions, language string, ext string) (string, error) {

	absFilePath, err := filepath.Abs(opts.FilePath)
	if err != nil {
		return "", err
	}

	var resolvedFilePath string
	var functionDir string
	var functionFile string
	if osutils.IsDirectory(absFilePath) {
		if opts.Artifact == "" {
			functionFile = opts.FunctionName
			functionDir = absFilePath
			if ext != "" {
				resolvedFilePath = filepath.Join(functionDir, fmt.Sprintf("%s.%s", functionFile, ext))
			} else {
				functionFile, err = searchForFunctionResource(functionDir, opts.FunctionName)
				if err != nil {
					return "", err
				}
				resolvedFilePath = functionFile
			}
		} else {
			resolvedFilePath = filepath.Join(absFilePath, opts.Artifact)
		}
	} else {
		resolvedFilePath = absFilePath
	}
	if !osutils.FileExists(resolvedFilePath) {
		return "", errors.New(fmt.Sprintf("function path %s does not exist", resolvedFilePath))
	}

	if opts.Artifact != "" && language != "" && languageForFileExtensions[filepath.Ext(resolvedFilePath)[1:]] != language {
		return "", errors.New(fmt.Sprintf("language %s conflicts with artifact file extension %s", language, opts.Artifact))
	}

	return resolvedFilePath, nil
}

func searchForFunctionResource(dir string, name string) (string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return "", err
	}

	foundFile := ""
	for _, f := range files {
		if b := filepath.Base(f); b[0:len(b)-len(filepath.Ext(f))] == name {
			for _, e := range supportedExtensions {
				if filepath.Ext(f) == "."+e {
					if foundFile == "" {
						foundFile = f
					} else {
						return "", errors.New(fmt.Sprintf("function file is not unique %s, %s", filepath.Base(foundFile), filepath.Base(f)))
					}
				}
			}
		}

	}

	if foundFile == "" {
		return "", errors.New(fmt.Sprintf("no function file found in path %s", dir))
	}
	return foundFile, nil
}
