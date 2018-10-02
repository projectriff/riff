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
 *
 */

package fileutils_test

import (
	"github.com/projectriff/riff/pkg/fileutils"
	"github.com/projectriff/riff/pkg/test_support"
	"testing"
)

func TestExistsDir(t *testing.T) {
	f := createChecker()
	td := test_support.CreateTempDir()
	defer test_support.CleanupDirs(t, td)

	exists := f.Exists(td)
	if !exists {
		t.Fatalf("Exists failed to find existing directory %s", td)
	}
}

func TestExistsFile(t *testing.T) {
	f := createChecker()
	td := test_support.CreateTempDir()
	defer test_support.CleanupDirs(t, td)

	src := test_support.CreateFile(td, "src.file")
	exists := f.Exists(src)
	if !exists {
		t.Fatalf("Exists failed to find existing file %s", src)
	}
}

func TestExistsFalse(t *testing.T) {
	f := createChecker()
	path := "/nosuch"
	exists := f.Exists(path)
	if exists {
		t.Fatalf("Exists claimed non-existent path %s exists", path)
	}
}

func createChecker() fileutils.Checker {
	return fileutils.NewChecker()
}
