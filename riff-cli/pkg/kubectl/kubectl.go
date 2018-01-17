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

package kubectl

import (
	"os/exec"
	"time"
	"context"
	"github.com/dturanski/riff-cli/pkg/ioutils"
	"bytes"
	"fmt"
)

func ExecForString(cmdArgs []string) (string, error) {
	out, err := ExecForBytes(cmdArgs)
	return string(out), err
}

func ExecForBytes(cmdArgs []string) ([]byte, error) {

	// Create a new context and add a timeout to it
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel() // The cancel should be deferred so resources are cleaned up

	cmdName := "kubectl"
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
