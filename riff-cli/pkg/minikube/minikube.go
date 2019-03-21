/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *        https://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package minikube

import (
	"os/exec"
	"strings"
	"fmt"
	"regexp"
)

//go:generate mockery -name=Minikube -inpkg

// Type Minikube abtracts away interaction with the minikube command line tool (may be used in development mode,
// in which case the address for the http gateway using a NodePort is not localhost, but rather `minikube ip`)
type Minikube interface {
	QueryIp() (string, error)
}

// realMinikube spawns a new process and runs the `minikube` command for real when asked to do so.
type realMinikube struct {
}

func (*realMinikube) QueryIp() (string, error) {
	cmdName := "minikube"

	cmd := exec.Command(cmdName, "ip")

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// minikube can print status messages as part of other commands (eg "There is a newer version available").
	// Look for ip on the last line only
	lines := strings.Split(string(output), "\n")
	// Last line is carriage return alone
	if len(lines) < 2 {
		return "", fmt.Errorf("Unable to parse minikube ip in command output:\n%s", output)
	}
	line := lines[len(lines)-2]
	if match, err := regexp.MatchString("^\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}$", line) ; !match || err != nil {
		return "", fmt.Errorf("Unable to parse minikube ip in command output:\n%s", output)
	}
	return line, nil
}

func RealMinikube() Minikube {
	return &realMinikube{}
}
