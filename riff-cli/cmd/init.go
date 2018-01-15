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
	"github.com/spf13/cobra"
	"github.com/dturanski/riff-cli/pkg/osutils"
	"path/filepath"
	"github.com/dturanski/riff-cli/pkg/ioutils"
	"container/heap"
	"fmt"
	"os"
)

//TODO: This command maybe should be split out into sub commands, e.g. , 'riff init <language>', e.g., classname only applies to java

type InitOptions struct {
	userAccount  string
	functionName string
	version      string
	functionPath string
	language     string
	protocol     string
	input        string
	output       string
	artifact     string
	classname    string
	riffVersion  string
	push         bool
}

var initOptions InitOptions

var fileExtenstions = map[string]string{
	"shell"		:  "sh",
	"java"		:   "java",
	"node"		:   "js",
	"js"		:   "js",
	"python"	: 	"py",
}

var functionFile string

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a function",
	Long: `Initialize the function based on the function source code specified as the functionPath, using the functionName
  and version specified for the function image repository and tag. For example, if you have a directory named 'square' containing a function 'square.js', you can cd to square and simply type :

riff init


This will generate the required Dockerfile and resource definitions using sensible defaults.`,

	PreRun: func(cmd *cobra.Command, args []string) {
		if initOptions.input == "" {
			initOptions.input = initOptions.functionName
		}

		if !osutils.FileExists(initOptions.functionPath) {
			ioutils.Errorf("File does not exist %s\n", initOptions.functionPath)
			os.Exit(1)
		}

		if osutils.IsDirectory(initOptions.functionPath) {
			if initOptions.language == "" {
				for lang, ext := range fileExtenstions {
					fileName := fmt.Sprintf("%s.%s",initOptions.functionName, ext)
					functionFile = filepath.Join(initOptions.functionPath, fileName)
					if osutils.FileExists(functionFile) {
						initOptions.language = lang
						break
					}
				}
				if initOptions.language == "" {
					ioutils.Errorf("cannot find function source for function %s in directory %s\n", initOptions.functionName, initOptions.functionPath)
					os.Exit(1)
				}
			} else {
				ext := fileExtenstions[initOptions.language]
				if ext == "" {
					ioutils.Errorf("language %s is unsupported \n", initOptions.language)
					os.Exit(1)
				}

				fileName := fmt.Sprintf("%s.%s",initOptions.functionName, ext)
				functionFile = filepath.Join(initOptions.functionPath, fileName)
				if !osutils.FileExists(functionFile) {
					ioutils.Errorf("cannot find function source for function %s\n", functionFile)
				}
			}
		}
	},

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Function file: %s\n", functionFile)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&initOptions.userAccount, "useraccount", "u", osutils.GetCurrentUsername(), "the Docker user account to be used for the image repository (defaults to current OS username")
	initCmd.Flags().StringVarP(&initOptions.functionName, "functionName", "n", osutils.GetCurrentBasePath(), "the functionName of the function (defaults to the functionName of the current directory)")
	initCmd.Flags().StringVarP(&initOptions.version, "version", "v", "0.0.1", "the version of the function (defaults to 0.0.1)")
	initCmd.Flags().StringVarP(&initOptions.functionPath, "functionPath", "f", osutils.GetCurrentBasePath(), "functionPath or directory to be used for the function resources, if a file is specified then the file's directory will be used (defaults to the current directory)")
	initCmd.Flags().StringVarP(&initOptions.language, "language", "l", "", "the language used for the function source (defaults to functionPath extension language type or 'shell' if directory specified)")
	initCmd.Flags().StringVarP(&initOptions.protocol, "protocol", "p", "", "the protocol to use for function invocations (defaults to 'stdio' for shell and python, to 'http' for java and node)")
	initCmd.Flags().StringVarP(&initOptions.input, "input", "i", "", "the functionName of the input topic (defaults to function functionName)")
	initCmd.Flags().StringVarP(&initOptions.input, "output", "o", "", "the functionName of the output topic (optional)")
	initCmd.Flags().StringVarP(&initOptions.artifact, "artifact", "a", "", "path to the function artifact, source code or jar file(defaults to function functionName with extension appended based on language:'.sh' for shell, '.jar' for java, '.js' for node and '.py' for python)")
	initCmd.Flags().StringVarP(&initOptions.classname, "classname", "", "", "the fully qualified class functionName of the Java function class (required for Java functions)")
	//TODO: Default?
	initCmd.Flags().StringVarP(&initOptions.riffVersion, "riffversion", "", "", "the version of riff to use when building containers")
	initCmd.Flags().BoolVarP(&initOptions.push, "push", "", false, "push the image to Docker registry")
}
