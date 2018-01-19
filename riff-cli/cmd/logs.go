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

package cmd

import (
	"fmt"
	"os/exec"
	"os"
	"bufio"

	"github.com/spf13/cobra"
	"github.com/projectriff/riff-cli/pkg/ioutils"
	"github.com/projectriff/riff-cli/pkg/kubectl"
)

type LogsOptions struct {
	function  string
	container string
	tail      bool
}

var logsOptions LogsOptions

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Display the logs for a running function",
	Long: `Display the logs for a running function For example:

riff logs -n myfunc -t

will tail the logs from the 'sidecar' container for the function 'myfunc'

`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf("Displaying logs for container %v of function %v\n", logsOptions.container, logsOptions.function)

		cmdArgs := []string{"get", "pod", "-l", "function=" + logsOptions.function, "-o", "jsonpath={.items[0].metadata.name}"}

		output, err := kubectl.ExecForString(cmdArgs)

		if err != nil {
			ioutils.Errorf("Error %v - Function %v may not be currently active\n%", err, logsOptions.function)
			return
		}

		pod := output

		tail := ""
		if logsOptions.tail {
			tail = "-f"
		}

		cmdArgs = []string{"logs", "-c", logsOptions.container, tail, pod}

		kubectlCmd := exec.Command("kubectl", cmdArgs...)
		cmdReader, err := kubectlCmd.StdoutPipe()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for kubectlCmd", err)
			return
		}

		scanner := bufio.NewScanner(cmdReader)
		go func() {
			for scanner.Scan() {
				fmt.Printf("%s\n", scanner.Text())
			}
		}()

		err = kubectlCmd.Start()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error starting kubectlCmd", err)
			return
		}

		err = kubectlCmd.Wait()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error waiting for kubectlCmd", err)
			return
		}

	},
}

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.Flags().StringVarP(&logsOptions.function, "functionName", "n", "", "The functionName of the function")
	logsCmd.Flags().StringVarP(&logsOptions.container, "container", "c", "sidecar", "The functionName of the function container (sidecar or main)")
	logsCmd.Flags().BoolVarP(&logsOptions.tail, "tail", "t", false, "Tail the logs")

	logsCmd.MarkFlagRequired("functionName")
}
