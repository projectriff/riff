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

package docker

import (
	"fmt"
	"strings"
	"bufio"
	"os/exec"
	"os"
)

//go:generate mockery -name=Docker -inpkg

// Docker abstracts away interaction via a docker client.
type Docker interface {
	// Exec requests that the given docker command be executed, possibly with additional args.
	Exec(command string, cmdArgs ... string) error
}

// processDocker interacts with docker by spawning a process and using the 'docker' command line tool for real.
type processDocker struct {
}

// dryRunDocker only prints out the docker commands that *would* run.
type dryRunDocker struct {
}

func RealDocker() Docker {
	return &processDocker{}
}

func DryRunDocker() Docker {
	return &dryRunDocker{}
}

func (d *processDocker) Exec(command string, cmdArgs ... string) error {
	cmd := createDockerCommand(command, cmdArgs...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	cmd.Start()
	print(bufio.NewScanner(stdout), "[STDOUT]")
	print(bufio.NewScanner(stderr), "[STDERR]")
	return cmd.Wait()
}

func createDockerCommand(command string, cmdArgs ...string) *exec.Cmd {
	commandAndArgs := append([]string{command}, cmdArgs...)
	cmd := exec.Command("docker", commandAndArgs...)
	cmd.Env = createEnv("HOME", "PATH", "DOCKER_HOST", "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH", "DOCKER_API_VERSION")
	return cmd
}

// create a process environment based on the environment variables with the given keys
func createEnv(keys ... string) []string {
	env := []string{}
	for _, key := range keys {
		if value, set := os.LookupEnv(key); set {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
	}
	return env
}

// to print the processed information when stdout gets a new line
func print(scanner *bufio.Scanner, prefix string) {
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("%s %s\n", prefix, line)
	}
}

func (d *dryRunDocker) Exec(command string, cmdArgs ... string) error {
	fmt.Printf("%s command: docker %s %s\n", strings.Title(command), command, strings.Join(cmdArgs, " "))
	return nil
}
