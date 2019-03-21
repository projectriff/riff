/*
 * Copyright 2018-2019 The original author or authors
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
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

// AbsFile returns an absolute file path for a given file path. If the file is a relative path, it is relative to the
// given base path. Either file or base may be URLs. Rather than returning a file URL, an absolute file path is
// returned instead.
func AbsFile(file string, base string) (string, error) {
	absolute, canonicalFile, err := IsAbsFile(file)
	if err != nil {
		return "", err
	}

	if absolute {
		return canonicalFile, nil
	}

	if filepath.IsAbs(base) {
		return filepath.Join(base, file), nil
	}

	b, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	if b.Scheme == "http" || b.Scheme == "https" || b.Scheme == "file" {
		b.Path = path.Join(b.Path, filepath.ToSlash(file))
		return b.String(), nil
	}
	if b.IsAbs() {
		return "", fmt.Errorf("unsupported URL scheme %s in %s", b.Scheme, base)
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, base, file), nil
}

// IsAbsFile returns true if and only if the given file string is either a URL or an absolute file path.
// If it returns true, it also returns a canonical form of the file string which is either a http or https URL or an
// absolute file path. In other words, the canonical form of a file URL is the absolute file path of the URL.
func IsAbsFile(file string) (bool, string, error) {
	if filepath.IsAbs(file) {
		return true, file, nil
	}

	u, err := url.Parse(file)
	if err != nil {
		return false, "", err
	}

	if u.Scheme == "http" || u.Scheme == "https" {
		return true, file, nil
	}

	if u.Scheme == "file" {
		return true, fileURLPath(u), nil
	}

	if u.IsAbs() {
		return false, "", fmt.Errorf("unsupported URL scheme %s in %s", u.Scheme, file)
	}

	return false, "", nil
}
