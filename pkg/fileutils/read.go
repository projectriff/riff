/*
 * Copyright 2018 The original author or authors
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
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime"
)

// Read reads the contents of the specified file. If the file is a relative path, it is relative to base.
// Either file or base may be URLs.
func Read(file string, base string) ([]byte, error) {
	absoluteFile, err := AbsFile(file, base)
	if err != nil {
		return nil, err
	}
	return readAbsFile(absoluteFile)
}

func readAbsFile(file string) ([]byte, error) {
	u, err := url.Parse(file)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "http" || u.Scheme == "https" {
		return downloadFile(u.String())
	}
	if u.Scheme == "file" {
		file = fileURLPath(u)
	}

	if !filepath.IsAbs(file) {
		return nil, fmt.Errorf("absolute path expected instead of relative path: %s", file)
	}

	return ioutil.ReadFile(file)
}

func fileURLPath(u *url.URL) string {
	file := u.Path

	// Sanitise path on Windows. See https://github.com/golang/go/issues/6027
	if runtime.GOOS == "windows" && len(file) > 0 && file[0] == '/' {
		file = file[1:]
	}

	return file
}

func downloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
