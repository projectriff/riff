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
	"testing"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"os"
	"github.com/projectriff/riff-cli/pkg/osutils"
)


func TestFunctionNameFromPathFromCurrentDirectory(t *testing.T) {
	currentDir,_ := filepath.Abs(".")
	os.Chdir(osutils.Path("../../test_data/shell/echo"))

	as := assert.New(t)

	fname, err := FunctionNameFromPath(".")
	as.NoError(err)

	as.Equal("echo",fname)

	os.Chdir(currentDir)
}

func TestFunctionNameFromRelativePath(t *testing.T) {
	as := assert.New(t)
	fname, err := FunctionNameFromPath(osutils.Path("../../test_data/shell/echo"))
	as.NoError(err)
	as.Equal("echo",fname)
}

func TestFunctionNameFromRegularFile(t *testing.T) {
	as := assert.New(t)
	fname, err := FunctionNameFromPath(osutils.Path("../../test_data/shell/echo/echo.sh"))
	as.NoError(err)
	as.Equal("echo",fname)
}

func TestFunctionNameFromInvalidPathIsEmpty(t *testing.T) {
	as := assert.New(t)
	fname, err := FunctionNameFromPath("a/b/c/d")
	as.Error(err)
	as.Empty(fname)
}


