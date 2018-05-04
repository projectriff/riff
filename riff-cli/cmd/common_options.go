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
	"os"
	"path/filepath"
	"strings"

	"github.com/projectriff/riff/riff-cli/pkg/functions"
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
)

func validateFunctionName(name *string, path string) error {
	var err error
	if *name == "" {
		*name, err = functions.FunctionNameFromPath(path)
	}
	return err
}

func validateAndCleanArtifact(artifact *string, path string) error {
	if *artifact != "" {
		fmt.Printf("artifact: %v\n", *artifact)
		absFilePath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		/*
		 * If artifact is relative to current directory or an absolute path then use the absolute path
		 * else make it relative to the file path directory. But it must be in the file path directory.
		 */
		var absArtifactPath string
		if strings.IndexRune(*artifact, '.') == 0 || strings.IndexRune(*artifact, os.PathSeparator) == 0 {
			absArtifactPath, err = filepath.Abs(*artifact)
			if err != nil {
				return err
			}
		} else {
			if osutils.IsDirectory(absFilePath) {
				absArtifactPath = filepath.Join(absFilePath, *artifact)
			} else {
				absArtifactPath = filepath.Join(filepath.Dir(absFilePath), *artifact)
			}
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
		*artifact = strings.Replace(absArtifactPath, absFilePath+string(os.PathSeparator), "", 1)
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
