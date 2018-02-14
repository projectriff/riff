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
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/projectriff/riff-cli/pkg/ioutils"
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

func FindRiffResourceDefinitionPaths(path string) ([]string, error) {
	functions, err := filepath.Glob(filepath.Join(path, "*-function.yaml"))
	if err != nil {
		return nil, err
	}
	topics, err := filepath.Glob(filepath.Join(path, "*-topics.yaml"))
	if err != nil {
		return nil, err
	}
	return append(functions, topics...), nil
}

func IsDirectory(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.Mode().IsDir()
}

func Path(filename string) string {
	path := filepath.Clean(filename)
	if os.PathSeparator == '/' {
		return path
	}
	return filepath.Join(strings.Split(path, "/")...)
}

func Exec(cmdName string, cmdArgs []string, timeout time.Duration) ([]byte, error) {
	// Create a new context and add a timeout to it
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel() // The cancel should be deferred so resources are cleaned up

	// Create the command with our context
	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	// This time we can simply use Output() to get the result.
	out, err := cmd.Output()
	if err != nil {
		ioutils.Error(fmt.Sprint(err) + ": " + stderr.String())
	}

	// We want to check the context error to see if the timeout was executed.
	// The error returned by cmd.Output() will be OS specific based on what
	// happens when a process is killed.
	if ctx.Err() == context.DeadlineExceeded {
		ioutils.Error("Command timed out")
		return nil, ctx.Err()
	}

	return out, err
}
