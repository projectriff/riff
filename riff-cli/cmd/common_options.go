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
	"path/filepath"
	"github.com/projectriff/riff/riff-cli/pkg/functions"
)

func validateFilepath(path *string) error {
	*path = filepath.Clean(*path)

	if *path == "" {
		*path, _ = filepath.Abs(".")
	}

	return nil
}

func validateFunctionName(name *string, path string) error {
	var err error
	if *name == "" {
		*name, err = functions.FunctionNameFromPath(path)
	}
	return err
}
