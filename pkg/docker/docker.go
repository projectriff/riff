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
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Docker interface {
	PushImage(name string, digest string, file string) error
	PullImage(name string, directory string) (digest string, err error)
}

// processDocker interacts with docker by spawning a process and running the `docker`
// command line tool.
type processDocker struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func (pd *processDocker) PushImage(name string, digest string, file string) error {
	if err := pd.exec(5*time.Minute, "image", "load", "-i", file); err != nil {
		return err
	}
	if err := pd.exec(1*time.Second, "image", "tag", digest, name); err != nil {
		return err
	}
	return pd.exec(10*time.Minute, "push", name)
}

func (pd *processDocker) PullImage(name string, directory string) (digest string, err error) {
	if err := pd.exec(10*time.Minute, "pull", name); err != nil {
		return "", err
	}
	b := new(bytes.Buffer)
	if err := pd.execWithStreams(pd.stdin, b, pd.stderr, 1*time.Second, "inspect", "--format='{{.Id}}'", name); err != nil {
		return "", err
	}
	if offset := strings.LastIndex(b.String(), "'sha256:"); offset == -1 {
		return "", fmt.Errorf("unable to extract digest of image %q. Command output was %q", name, b.String())
	} else {
		// chop single quote at start, single quote and \n at end
		digest = b.String()[offset+len("'") : len(b.String())-len("'\n")]
		if err := pd.exec(5*time.Minute, "image", "save", "-o", filepath.Join(directory, digest), name); err != nil {
			return "", err
		}
		return digest, nil
	}
}

func (pd *processDocker) execWithStreams(stdin io.Reader, stdout io.Writer, stderr io.Writer, timeout time.Duration, cmdArgs ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", cmdArgs...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func (pd *processDocker) exec(timeout time.Duration, cmdArgs ...string) error {
	return pd.execWithStreams(pd.stdin, pd.stdout, pd.stderr, timeout, cmdArgs...)
}

func RealDocker(stdin io.Reader, stdout io.Writer, stderr io.Writer) Docker {
	return &processDocker{stdin: stdin, stdout: stdout, stderr: stderr}
}
