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

package osutils

import (
	"os"
	"path/filepath"
	"os/user"
	"github.com/dturanski/riff-cli/pkg/ioutils"
	"strings"
)

func GetCWD() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return cwd
}

func GetCWDBasePath() string {
	return filepath.Base(GetCWD())
}

func GetCurrentUsername() string {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	return user.Username
}


func FileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func IsDirectory(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		ioutils.Error(err)
		return false
	}
	return fi.Mode().IsDir()
}

func Path(filename string) string {
	path := filepath.Clean(filename)
	if os.PathSeparator == '/' {
		return path
	}
	return filepath.Join(strings.Split(path,"/")...)
}