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
package functions

import (
	"path/filepath"
	"github.com/projectriff/riff-cli/pkg/osutils"
	"errors"
	"fmt"
)

func FunctionNameFromPath(path string) (string, error) {
	if !osutils.FileExists(path) {
		return "", errors.New(fmt.Sprintf("path %s does not exist",path));
	}
	abs,err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if osutils.IsDirectory(abs) {
		return filepath.Base(abs), nil
	}
	return filepath.Base(filepath.Dir(abs)), nil

}
