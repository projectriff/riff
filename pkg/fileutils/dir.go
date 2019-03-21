/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fileutils

import (
	"net/url"
	"path"
	"path/filepath"
)

// Dir returns the directory portion of the given file.
func Dir(file string) (string, error) {
	if filepath.IsAbs(file) {
		return filepath.Dir(file), nil
	}
	u, err := url.Parse(file)
	if err != nil {
		return "", err
	}
	if u.IsAbs() {
		u.Path = path.Dir(u.Path)
		return u.String(), nil
	}
	return filepath.Dir(file), nil
}
