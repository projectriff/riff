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
 *
 */

package docker

import (
	"context"
	"io"
	"os/exec"
	"time"
)

type Docker interface {
	PushImage(name string, digest string, file string) error
}

// processDocker interacts with docker by spawning a process and running the `docker`
// command line tool.
type processDocker struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func (kc *processDocker) PushImage(name string, digest string, file string) error {
	if err := kc.exec(5*time.Minute, "image", "load", "-i", file); err != nil {
		return err
	}
	if err := kc.exec(1*time.Second, "image", "tag", digest, name); err != nil {
		return err
	}
	return kc.exec(10*time.Minute, "push", name)
}

func (kc *processDocker) exec(timeout time.Duration, cmdArgs ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", cmdArgs...)
	cmd.Stdin = kc.stdin
	cmd.Stdout = kc.stdout
	cmd.Stderr = kc.stderr
	return cmd.Run()
}

func RealDocker(stdin io.Reader, stdout io.Writer, stderr io.Writer) Docker {
	return &processDocker{stdin: stdin, stdout: stdout, stderr: stderr}
}
