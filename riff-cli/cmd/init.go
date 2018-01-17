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
	"github.com/spf13/cobra"
	"github.com/dturanski/riff-cli/pkg/osutils"
	"github.com/dturanski/riff-cli/pkg/ioutils"
	"path/filepath"
	"fmt"
	"errors"
	"strings"
)

var initOptions InitOptions

var initCmd = &cobra.Command{
	Use:   "init [language]",
	Short: "Initialize a function",
	Long: `Initialize the function based on the function source code specified as the filename, using the name
and version specified for the function image repository and tag. For example, if you have a directory named 'square' containing a function 'square.js', you can simply type :

riff init node -f square

or

riff init node

from the 'square' directory

to generate the required Dockerfile and resource definitions using sensible defaults.`,


	Run: func(cmd *cobra.Command, args []string) {

		initializer := NewLanguageDetectingInitializer()
		err := initializer.initialize(makeHandlerAwareOptions(cmd))
		if err != nil {
			ioutils.Error(err)
			return
		}
	},
}

var initJavaCmd = &cobra.Command{
	Use:   "java",
	Short: "Initialize a Java function",
	Long: `Initialize the function based on the function source code specified as the filename, using the artifact (jar file), the function classname, the name
  and version specified for the function image repository and tag. For example from a maven project directory named 'greeter', type:

riff init -i greetings -l java -a target/greeter-1.0.0.jar --classname=Greeter


to generate the required Dockerfile and resource definitions using sensible defaults.`,


	Run: func(cmd *cobra.Command, args []string) {
		err := validateAndCleanInitOptions(&initOptions)
		if err != nil {
			ioutils.Error(err)
			return
		}

		initializer := NewJavaInitializer()
		err = initializer.initialize(makeHandlerAwareOptions(cmd))
		if err != nil {
			ioutils.Error(err)
			return
		}
	},
}

var initNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Initialize a node.js function",
	Long: `Initialize the function based on the function source code specified as the filename, using the name
and version specified for the function image repository and tag. For example, if you have a directory named 'square' containing a function 'square.js', you can simply type :

riff init node -f square

or

riff init node

from the 'square' directory

to generate the required Dockerfile and resource definitions using sensible defaults.`,

	Run: initializeNode,
}

var initJsCmd = &cobra.Command{
	Use:   "js",
	Short: initNodeCmd.Short,
	Long:  initNodeCmd.Long,
	Run:   initNodeCmd.Run,
}

func initializeNode(cmd *cobra.Command, args []string) {
	err := validateAndCleanInitOptions(&initOptions)
	if err != nil {
		ioutils.Error(err)
		return
	}

	initializer := NewNodeInitializer()
	err = initializer.initialize(initOptions)

	if err != nil {
		ioutils.Error(err)
		return
	}
}

var initPythonCmd = &cobra.Command{
	Use:   "python",
	Short: "Initialize a Python function",
	Long: `Initialize the function based on the function source code specified as the filename, handler, name, artifact
  and version specified for the function image repository and tag. For example, type:

riff init -i words -l python  --n uppercase --handler=process


to generate the required Dockerfile and resource definitions using sensible defaults.`,


	Run: func(cmd *cobra.Command, args []string) {
		err := validateAndCleanInitOptions(&initOptions)
		if err != nil {
			ioutils.Error(err)
			return
		}

		initializer := NewPythonInitializer()

		err = initializer.initialize(makeHandlerAwareOptions(cmd))
		if err != nil {
			ioutils.Error(err)
			return
		}
	},
}

func makeHandlerAwareOptions(cmd *cobra.Command) HandlerAwareInitOptions {
	handler, _ := cmd.Flags().GetString("handler")
	return *NewHandlerAwareInitOptions(initOptions, handler)
}

/*
 * Basic sanity check, given files exist, valid protocol given
 * TODO: Format (regex) check on function name, input, output, version, riff_version
 */
func validateAndCleanInitOptions(options *InitOptions) error {

	options.functionPath = filepath.Clean(options.functionPath)
	if options.artifact != "" {
		options.artifact = filepath.Clean(options.artifact)
	}

	if options.functionPath != "" {
		if !osutils.FileExists(options.functionPath) {
			return errors.New(fmt.Sprintf("filepath %s does not exist", options.functionPath))
		}
	}

	if options.Artifact() != "" {
		absArtifactPath, err := filepath.Abs(options.artifact)
		if err != nil {
			return err
		}

		absFilePath, err := filepath.Abs(options.functionPath)
		if err != nil {
			return err
		}

		if osutils.IsDirectory(absArtifactPath) {
			return errors.New(fmt.Sprintf("artifact %s must be a regular file", absArtifactPath))
		}

		absFilePathDir := absFilePath
		if !osutils.IsDirectory(absFilePath) {
			absFilePathDir = filepath.Dir(absFilePath)
		}

		if absFilePathDir != filepath.Dir(absArtifactPath) {
			return errors.New(fmt.Sprintf("artifact %s cannot be external to filepath %", absArtifactPath, absFilePath))
		}


		if !osutils.FileExists(absArtifactPath) {
			return errors.New(fmt.Sprintf("artifact %s does not exist", absArtifactPath))
		}

		if !osutils.IsDirectory(absFilePath) && absFilePath != absArtifactPath {
			return errors.New(fmt.Sprintf("artifact %s conflicts with filepath %s", absArtifactPath, absFilePath))
		}
	}

	if options.protocol != "" {

		supported := false
		options.protocol = strings.ToLower(options.protocol)
		for _, p := range supportedProtocols {
			if options.protocol == p {
				supported = true
			}
		}
		if (!supported) {
			return errors.New(fmt.Sprintf("protocol %s is unsupported \n", options.protocol))
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.PersistentFlags().StringVarP(&initOptions.userAccount, "useraccount", "u", osutils.GetCurrentUsername(), "the Docker user account to be used for the image repository (defaults to current OS username")
	initCmd.PersistentFlags().StringVarP(&initOptions.functionName, "name", "n", "", "the functionName of the function (defaults to the functionName of the current directory)")
	initCmd.PersistentFlags().StringVarP(&initOptions.version, "version", "v", "0.0.1", "the version of the function (defaults to 0.0.1)")
	initCmd.PersistentFlags().StringVarP(&initOptions.functionPath, "filepath", "f", osutils.GetCWD(), "Path or directory to be used for the function resources, if a file is specified then the file's directory will be used (defaults to the current directory)")
	initCmd.PersistentFlags().StringVarP(&initOptions.protocol, "protocol", "p", "", "the protocol to use for function invocations (defaults to 'stdio' for shell and python, to 'http' for java and node)")
	initCmd.PersistentFlags().StringVarP(&initOptions.input, "input", "i", "", "the functionName of the input topic (defaults to function functionName)")
	initCmd.PersistentFlags().StringVarP(&initOptions.output, "output", "o", "", "the functionName of the output topic (optional)")
	initCmd.PersistentFlags().StringVarP(&initOptions.artifact, "artifact", "a", "", "path to the function artifact, source code or jar file")
	initCmd.PersistentFlags().StringVarP(&initOptions.riffVersion, "riff-version", "", "0.0.1", "the version of riff to use when building containers")
	initCmd.PersistentFlags().BoolVarP(&initOptions.push, "push", "", false, "push the image to Docker registry")

	initCmd.AddCommand(initJavaCmd)

	initJavaCmd.Flags().String("handler", "", "the fully qualified class name of the function handler")
	initJavaCmd.MarkFlagRequired("handler")

	initPythonCmd.Flags().String("handler", "", "the name of the function handler")
	initPythonCmd.MarkFlagRequired("handler")

}
