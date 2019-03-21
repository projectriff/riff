/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *        https://www.apache.org/licenses/LICENSE-2.0
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
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
)

// FunctionNameFromPath returns a name for a function derived from the owning directory of the provided path.
// That is, if path is a directory, its last element is returned. If path denotes a file, the name of its parent directory
// is returned.
func FunctionNameFromPath(path string) (string, error) {
	abs,err := osutils.AbsPath(path)
	if err != nil {
		return "", err
	}
	if osutils.IsDirectory(abs) {
		return filepath.Base(abs), nil
	}
	return filepath.Base(filepath.Dir(abs)), nil
}

