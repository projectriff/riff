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
	"github.com/dturanski/riff-cli/pkg/function"
	"github.com/dturanski/riff-cli/pkg/ioutils"
)

//TODO: This command maybe should be split out into sub commands, e.g. , 'riff init <language>', e.g., classname only applies to java

var initOptions function.InitOptions

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a function",
	Long: `Initialize the function based on the function source code specified as the functionPath, using the functionName
  and version specified for the function image repository and tag. For example, if you have a directory named 'square' containing a function 'square.js', you can cd to square and simply type :

riff init


This will generate the required Dockerfile and resource definitions using sensible defaults.`,


	Run: func(cmd *cobra.Command, args []string) {
		initializer := function.NewInitializer()
		err := initializer.Initialize(initOptions)
		if err != nil {
			ioutils.Error(err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&initOptions.UserAccount, "useraccount", "u", osutils.GetCurrentUsername(), "the Docker user account to be used for the image repository (defaults to current OS username")
	initCmd.Flags().StringVarP(&initOptions.FunctionName, "name", "n", "", "the functionName of the function (defaults to the functionName of the current directory)")
	initCmd.Flags().StringVarP(&initOptions.Version, "version", "v", "0.0.1", "the version of the function (defaults to 0.0.1)")
	initCmd.Flags().StringVarP(&initOptions.FunctionPath, "functionPath", "f", osutils.GetCWD(), "functionPath or directory to be used for the function resources, if a file is specified then the file's directory will be used (defaults to the current directory)")
	initCmd.Flags().StringVarP(&initOptions.Language, "language", "l", "", "the language used for the function source (defaults to functionPath extension language type or 'shell' if directory specified)")
	initCmd.Flags().StringVarP(&initOptions.Protocol, "protocol", "p", "", "the protocol to use for function invocations (defaults to 'stdio' for shell and python, to 'http' for java and node)")
	initCmd.Flags().StringVarP(&initOptions.Input, "input", "i", "", "the functionName of the input topic (defaults to function functionName)")
	initCmd.Flags().StringVarP(&initOptions.Output, "output", "o", "", "the functionName of the output topic (optional)")
	initCmd.Flags().StringVarP(&initOptions.Artifact, "artifact", "a", "", "path to the function artifact, source code or jar file(defaults to function functionName with extension appended based on language:'.sh' for shell, '.jar' for java, '.js' for node and '.py' for python)")
	initCmd.Flags().StringVarP(&initOptions.Classname, "classname", "", "", "the fully qualified class functionName of the Java function class (required for Java functions)")
	initCmd.Flags().StringVarP(&initOptions.RiffVersion, "riff-version", "", "0.0.1", "the version of riff to use when building containers")
	initCmd.Flags().BoolVarP(&initOptions.Push, "push", "", false, "push the image to Docker registry")
}
