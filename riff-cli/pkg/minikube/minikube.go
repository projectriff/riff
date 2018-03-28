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

package minikube

import (
	"os/exec"
	"strings"
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

	output, err := cmd.CombinedOutput()

	return strings.TrimRight(string(output), "\n"), err
}

func RealMinikube() Minikube {
	return &realMinikube{}
}
