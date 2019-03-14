/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fileutils

import (
	"fmt"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

func ResolveTilde(path string) (string, error) {
	if !StartsWithCurrentUserDirectoryAsTilde(path, runtime.GOOS) {
		return path, nil
	}
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	homeDirectory := currentUser.HomeDir
	if homeDirectory == "" {
		return "", fmt.Errorf("current user %s has no resolvable home directory", currentUser.Name)
	}
	return filepath.Join(homeDirectory, path[2:]), nil
}

func StartsWithCurrentUserDirectoryAsTilde(path string, os string) bool {
	if strings.HasPrefix(path, "~/") {
		return true
	}
	if os != "windows" {
		return false
	}
	return strings.HasPrefix(path, `~\`)
}
