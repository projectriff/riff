/*
 * Copyright 2017 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/projectriff/riff/riff-cli/pkg/functions"
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
)

func validateFunctionName(name *string, path string) error {
	var err error
	if *name == "" {
		*name, err = functions.FunctionNameFromPath(path)
	}
	if err != nil {
		return err
	}

	re := regexp.MustCompile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$")

	if !re.MatchString(*name) {
		return fmt.Errorf("function name %s is invalid. It must comply with DNS 1123 naming standards. The name must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character", *name)
	}
	return nil
}

func validateAndCleanArtifact(artifact *string, path string) error {
	if *artifact != "" {
		*artifact = filepath.Clean(*artifact)

		if filepath.IsAbs(*artifact) {
			return fmt.Errorf("artifact %s must be relative to function path", *artifact)
		}

		absFilePath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		var absArtifactPath string

		if osutils.IsDirectory(absFilePath) {
			absArtifactPath = filepath.Join(absFilePath, *artifact)
		} else {
			absArtifactPath = filepath.Join(filepath.Dir(absFilePath), *artifact)
		}

		if osutils.IsDirectory(absArtifactPath) {
			return fmt.Errorf("artifact %s must be a regular file", absArtifactPath)
		}

		absFilePathDir := absFilePath
		if !osutils.IsDirectory(absFilePath) {
			absFilePathDir = filepath.Dir(absFilePath)
		}

		if !strings.HasPrefix(filepath.Dir(absArtifactPath), absFilePathDir) {
			return fmt.Errorf("artifact %s cannot be external to filepath %s", absArtifactPath, absFilePath)
		}

		if !osutils.FileExists(absArtifactPath) {
			return fmt.Errorf("artifact %s does not exist", absArtifactPath)
		}

		if !osutils.IsDirectory(absFilePath) && absFilePath != absArtifactPath {
			return fmt.Errorf("artifact %s conflicts with filepath %s", absArtifactPath, absFilePath)
		}
	}
	return nil
}

func validateProtocol(protocol *string) error {
	supportedProtocols := []string{"http", "grpc"}
	if *protocol != "" {

		supported := false
		*protocol = strings.ToLower(*protocol)
		for _, p := range supportedProtocols {
			if *protocol == p {
				supported = true
			}
		}
		if !supported {
			return fmt.Errorf("protocol %s is unsupported \n", *protocol)
		}
	}

	return nil
}
