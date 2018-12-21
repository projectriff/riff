/*
 * Copyright 2014-2018 The original author or authors
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
	"os"
)

const ErrFileNotFound ErrorId = "file not found"

// Checker is a helper interface for querying files and directories.
type Checker interface {
	/*
		Tests the existence of a file or directory at a given path. Returns true if and only if the file or
		directory exists.
	*/
	Exists(path string) bool

	/*
		Filemode returns the os.FileMode of the file with the given path. If the file does not exist, returns
		an error with tag ErrFileNotFound.
	*/
	Filemode(path string) (os.FileMode, error)
}

type checker struct{}

func NewChecker() *checker {
	return &checker{}
}

// Exists returns true if and only if a file exists at the given path.
func (c *checker) Exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// Filemode returns the file mode of the file at the given path.
func (c *checker) Filemode(path string) (os.FileMode, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return os.FileMode(0), newFileError(ErrFileNotFound, err)
	}
	return fi.Mode(), nil
}
