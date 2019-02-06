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

package test_support

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	. "github.com/onsi/gomega"
)

func CreateTempDir() string {
	tempDir, err := ioutil.TempDir("", "riff-test-")
	check(err)
	return tempDir
}

func CreateFile(path string, fileName string, contents ...string) string {
	return CreateFileWithMode(path, fileName, os.FileMode(0666), contents...)
}

func CreateFileWithMode(path string, fileName string, mode os.FileMode, contents ...string) string {
	fp := filepath.Join(path, fileName)
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	check(err)
	defer func() {
		err = f.Close()
		check(err)
	}()
	if len(contents) == 0 {
		_, err = f.WriteString("test contents")
		check(err)
	} else {
		for _, c := range contents {
			_, err = f.WriteString(c)
			check(err)
		}
	}
	return fp
}

func CreateDir(path string, dirName string) string {
	return CreateDirWithMode(path, dirName, os.FileMode(0755))
}

func CreateDirWithMode(path string, dirName string, mode os.FileMode) string {
	fp := filepath.Join(path, dirName)
	err := os.Mkdir(fp, mode)
	check(err)
	return fp
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func SameFile(p1 string, p2 string) bool {
	fi1, err := os.Stat(p1)
	check(err)
	fi2, err := os.Stat(p2)
	check(err)
	return os.SameFile(fi1, fi2)
}

func FileMode(path string) os.FileMode {
	fi, err := os.Lstat(path)
	check(err)
	return fi.Mode()
}

func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

type ErrorReporter interface {
	Errorf(format string, args ...interface{})
}

func CleanupDirs(t ErrorReporter, paths ...string) {
	for _, path := range paths {
		if err := os.RemoveAll(path); err != nil {
			t.Errorf("Could not delete %s", path)
		}
	}
}

func FileURL(absolutePath string) string {
	Expect(filepath.IsAbs(absolutePath)).To(BeTrue(), fmt.Sprintf("FileURL called with relative path: %s", absolutePath))
	extra := ""
	if runtime.GOOS == "windows" {
		extra = "/"
	}
	return fmt.Sprintf("file://%s%s", extra, absolutePath)
}

func AbsolutePath(path string) string {
	result, err := filepath.Abs(path)
	check(err)
	return result
}
