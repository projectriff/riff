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

package osutils

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

func Exec(cmdName string, cmdArgs []string, timeout time.Duration) ([]byte, error) {
	return ExecStdin(cmdName, cmdArgs, nil, timeout)
}

func ExecStdin(cmdName string, cmdArgs []string, stdin *[]byte, timeout time.Duration) ([]byte, error) {
	// Create a new context and add a timeout to it
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel() // The cancel should be deferred so resources are cleaned up

	// Create the command with our context
	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)

	if stdin != nil {
		cmd.Stdin = bytes.NewBuffer(*stdin)
	}
	// This time we can simply use Output() to get the result.
	out, err := cmd.Output()

	// We want to check the context error to see if the timeout was executed.
	// The error returned by cmd.Output() will be OS specific based on what
	// happens when a process is killed.
	if ctx.Err() == context.DeadlineExceeded {
		return nil, ctx.Err()
	}

	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.Stderr, err
	}

	return out, err
}
