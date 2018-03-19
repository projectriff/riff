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
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
	"time"
	"fmt"
	"strings"
)

var EXEC_FOR_STRING = ExecForString
var EXEC_FOR_BYTES = ExecForBytes

// go:generate mockery -name=KubeCtl -inpkg

type KubeCtl interface {
	Exec(cmdArgs []string) (string, error)
}

// processKubeCtl interacts with kubernetes by spawning a process and running the kubectl
// command line tool.
type processKubeCtl struct {
}

func (kc *processKubeCtl) Exec(cmdArgs []string) (string, error) {
	out, err := osutils.Exec("kubectl", cmdArgs, 20*time.Second)
	return string(out), err
}

// dryRunKubeCtl only prints out the kubectl commands that *would* run.
type dryRunKubeCtl struct {
}

func (kc *dryRunKubeCtl) Exec(cmdArgs []string) (string, error) {
	fmt.Printf("%s command: kubectl %s\n", strings.Title(cmdArgs[0]), strings.Join(cmdArgs, " "))
	return "", nil
}

func RealKubeCtl() KubeCtl {
	return &processKubeCtl{}
}

func DryRunKubeCtl() KubeCtl {
	return &dryRunKubeCtl{}
}

func ExecForString(cmdArgs []string) (string, error) {
	out, err := EXEC_FOR_BYTES(cmdArgs)
	return string(out), err
}

func ExecForBytes(cmdArgs []string) ([]byte, error) {
	return osutils.Exec("kubectl", cmdArgs, 20*time.Second)
}
