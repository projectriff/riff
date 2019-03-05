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

package kubectl

import (
	"time"

	"github.com/projectriff/riff/pkg/osutils"
)

type KubeCtl interface {
	Exec(cmdArgs []string) (string, error)
	ExecStdin(cmdArgs []string, stdin *[]byte) (string, error)
}

// processKubeCtl interacts with kubernetes by spawning a process and running the kubectl
// command line tool.
type processKubeCtl struct {
	configLocation        string
	serverAddressOverride string
}

func (kc *processKubeCtl) Exec(cmdArgs []string) (string, error) {
	args := kc.withConfigFlags(cmdArgs)
	out, err := osutils.Exec("kubectl", args, 60*time.Second)
	return string(out), err
}

func (kc *processKubeCtl) ExecStdin(cmdArgs []string, stdin *[]byte) (string, error) {
	args := kc.withConfigFlags(cmdArgs)
	out, err := osutils.ExecStdin("kubectl", args, stdin, 60*time.Second)
	return string(out), err
}

func (kc *processKubeCtl) withConfigFlags(otherArgs []string) []string {
	flags := []string{"--kubeconfig", kc.configLocation}
	if kc.serverAddressOverride != "" {
		flags = append(flags, "--server", kc.serverAddressOverride)
	}
	return append(flags, otherArgs...)
}

func RealKubeCtl(configLocation string, serverAddressOverride string) KubeCtl {
	return &processKubeCtl{
		configLocation:        configLocation,
		serverAddressOverride: serverAddressOverride,
	}
}
