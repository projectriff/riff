// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os/exec"
	"os"
	"bufio"
)

type LogsOptions struct {
	Function        string
	Container 		string
	Tail            bool
}

var options LogsOptions

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show logs for a function resource",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Displaying logs for container %v of function %v\n", options.Container, options.Function)

		//kubectl logs ${TAIL} -c ${CONTAINER} $(kubectl get pod -l function=${FUNCTION} -o jsonpath='{.items[0].metadata.name}')

		cmdName := "kubectl"
		cmdArgs := []string{"get","pod","-l","function="+ options.Function, "-o", "jsonpath={.items[0].metadata.name}"}

		command := exec.Command(cmdName, cmdArgs...)


		output, err := command.CombinedOutput()
		if (err != nil) {
			fmt.Printf("Error getting pod %v for function %v\n", err, options.Function)
			fmt.Println(string(output))
			return
		}

		pod:= string(output)

		tail:= ""
		if options.Tail {tail = "-f"}


		cmdArgs = []string{"logs", "-c", options.Container, tail, pod}

		command = exec.Command(cmdName, cmdArgs...)
		cmdReader, err := command.StdoutPipe()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for command", err)
			return
		}

		scanner := bufio.NewScanner(cmdReader)
		go func() {
			for scanner.Scan() {
				fmt.Printf("%s\n", scanner.Text())
			}
		}()

		err = command.Start()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error starting command", err)
			return
		}

		err = command.Wait()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error waiting for command", err)
			return
		}

	},
}


func init() {
	rootCmd.AddCommand(logsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// logsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// logsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	logsCmd.Flags().StringVarP(&options.Container, "container", "c", "sidecar", "The name of the function container (sidecar or main)")
	logsCmd.Flags().StringVarP(&options.Function, "name", "n", "", "The name of the function")
	logsCmd.Flags().BoolVarP(&options.Tail,"tail","t",false, "Tail the logs" )
}
