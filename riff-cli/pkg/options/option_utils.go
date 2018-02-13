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

package options

import (
	"path/filepath"
	"github.com/projectriff/riff-cli/pkg/osutils"
	"fmt"
	"strings"
	"errors"
	"github.com/projectriff/riff-cli/pkg/functions"
)

func ImageName(opts ImageOptions) string {
	return fmt.Sprintf("%s/%s:%s",opts.GetUserAccount(),opts.GetFunctionName(),opts.GetVersion())
}

func ValidateNamePathOptions(name *string, filePath *string) error {
	*filePath = filepath.Clean(*filePath)

	if *filePath == "" {
		path, _ := filepath.Abs(".")
		*filePath = path
	}

	var err error
	if *name == "" {
		*name, err = functions.FunctionNameFromPath(*filePath)
		if err != nil {
			return err
		}
	}

	return nil;
}

/*
 * Basic sanity check that given paths exist and valid protocol given.
 * Artifact must be a regular file.
 * If artifact is given, it must be relative to the function path.
 * If function path is given as a regular file, and artifact is also given, they must reference the same path (edge case).
 * TODO: Format (regex) check on function name, input, output, version, riff_version
 */
func ValidateAndCleanInitOptions(options *InitOptions) error {

	options.FunctionPath = filepath.Clean(options.FunctionPath)
	if options.Artifact != "" {
		options.Artifact = filepath.Clean(options.Artifact)
	}

	if options.FunctionPath == "" {
		path, _ := filepath.Abs(".")
		options.FunctionPath = path
	}

	var err error
	if options.FunctionName == "" {
		options.FunctionName, err = functions.FunctionNameFromPath(options.FunctionPath)
		if err != nil {
			return err
		}
	}

	if options.Artifact != "" {

		if filepath.IsAbs(options.Artifact) {
			return errors.New(fmt.Sprintf("artifact %s must be relative to function path", options.Artifact))
		}

		absFilePath, err := filepath.Abs(options.FunctionPath)
		if err != nil {
			return err
		}

		var absArtifactPath string

		if osutils.IsDirectory(absFilePath) {
			absArtifactPath = filepath.Join(absFilePath, options.Artifact)
		} else {
			absArtifactPath = filepath.Join(filepath.Dir(absFilePath), options.Artifact)
		}

		if osutils.IsDirectory(absArtifactPath) {
			return errors.New(fmt.Sprintf("artifact %s must be a regular file", absArtifactPath))
		}

		absFilePathDir := absFilePath
		if !osutils.IsDirectory(absFilePath) {
			absFilePathDir = filepath.Dir(absFilePath)
		}

		if !strings.HasPrefix(filepath.Dir(absArtifactPath), absFilePathDir) {
			return errors.New(fmt.Sprintf("artifact %s cannot be external to filepath %", absArtifactPath, absFilePath))
		}

		if !osutils.FileExists(absArtifactPath) {
			return errors.New(fmt.Sprintf("artifact %s does not exist", absArtifactPath))
		}

		if !osutils.IsDirectory(absFilePath) && absFilePath != absArtifactPath {
			return errors.New(fmt.Sprintf("artifact %s conflicts with filepath %s", absArtifactPath, absFilePath))
		}
	}


	if options.Protocol != "" {

		supported := false
		options.Protocol = strings.ToLower(options.Protocol)
		for _, p := range SupportedProtocols {
			if options.Protocol == p {
				supported = true
			}
		}
		if (!supported) {
			return errors.New(fmt.Sprintf("protocol %s is unsupported \n", options.Protocol))
		}
	}

	return nil
}
