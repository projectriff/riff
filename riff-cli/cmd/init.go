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
	"github.com/projectriff/riff-cli/pkg/osutils"
	"github.com/projectriff/riff-cli/pkg/ioutils"
	"path/filepath"
	"fmt"
	"errors"
	"os"
	"strings"
)

var initOptions InitOptions

const (
	initResult = `generate the required Dockerfile and resource definitions using sensible defaults`
	initDefinition = `Generate`
)

const initCommandDescription =  `{{.Process}} the function based on the function source code specified as the filename, using the name
and version specified for the function image repository and tag. 
For example, if you have a directory named 'square' containing a function 'square.js', you can simply type :

riff {{.Command}} node -f square

or

riff  {{.Command}} node

from the 'square' directory

to {{.Result}}.`

var initCmd = &cobra.Command{
	Use:   "init [language]",
	Short: "Initialize a function",
	Long: createCmdLong(initCommandDescription, LongVals{Process: initDefinition, Command:"init", Result:initResult}),

	Run: func(cmd *cobra.Command, args []string) {

		initializer := NewLanguageDetectingInitializer()
		err := initializer.initialize(*makeHandlerAwareOptions(cmd))
		if err != nil {
			ioutils.Error(err)
			return
		}
	},

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			if len(args) == 1 && initOptions.functionPath == "" {
				initOptions.functionPath = args[0]
			} else {
				ioutils.Errorf("Invalid argument(s) %v\n", args)
				cmd.Usage()
				os.Exit(1)
			}
		}
		if initOptions.functionPath == "" {
			initOptions.functionPath = osutils.GetCWD()
		}
		err := validateAndCleanInitOptions(&initOptions)
		if err != nil {
			ioutils.Error(err)
			os.Exit(1)
		}
	},
}

const initJavaDescription = `{{.Process}} the function based on the function source code specified as the filename, using the artifact (jar file), 
the function handler(classname), the name and version specified for the function image repository and tag. 
For example from a maven project directory named 'greeter', type:

riff {{.Command}} -i greetings -l java -a target/greeter-1.0.0.jar --handler=Greeter


to generate the required Dockerfile and resource definitions using sensible defaults.`

var initJavaCmd = &cobra.Command{
	Use:   "java",
	Short: "Initialize a Java function",
	Long: 	createCmdLong(initJavaDescription, LongVals{Process:initDefinition, Command:"init java", Result:initResult}),
	Run: func(cmd *cobra.Command, args []string) {

		initializer := NewJavaInitializer()
		err := initializer.initialize(*makeHandlerAwareOptions(cmd))
		if err != nil {
			ioutils.Error(err)
			return
		}
	},
}

const initShellDescription = `{{.Process}} the function based on the function script specified as the filename, 
using the name and version specified for the function image repository and tag. 
For example, if you have a directory named 'echo' containing a function 'echo.sh', you can simply type :

riff {{.Command}} -f echo

or

riff {{.Command}}

from the 'echo' directory

to {{.Result}}.`

var initShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Initialize a shell script function",
	Long: 	createCmdLong(initShellDescription, LongVals{Process:initDefinition, Command:"init shell", Result:initResult}),

	Run: func(cmd *cobra.Command, args []string) {
		initializer := NewShellInitializer()
		err := initializer.initialize(initOptions)
		if err != nil {
			ioutils.Error(err)
			return
		}
	},
}

const initNodeDescription = `{{.Process}} the function based on the function source code specified as the filename, using the name
and version specified for the function image repository and tag.
For example, if you have a directory named 'square' containing a function 'square.js', you can simply type :

riff {{.Command}} -f square

or

riff {{.Command}}

from the 'square' directory

to {{.Result}}.`

var initNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Initialize a node.js function",
	Long: 	createCmdLong(initNodeDescription, LongVals{Process:initDefinition, Command:"init node", Result:initResult}),

	Run: initializeNode,
}

var initJsCmd = &cobra.Command{
	Use:   "js",
	Short: initNodeCmd.Short,
	Long:  initNodeCmd.Long,
	Run:   initNodeCmd.Run,
}

func initializeNode(cmd *cobra.Command, args []string) {

	initializer := NewNodeInitializer()
	err := initializer.initialize(initOptions)
	if err != nil {
		ioutils.Error(err)
		return
	}
}

const initPythonDescription = `{{.Process}} the function based on the function source code specified as the filename, handler, name, artifact
  and version specified for the function image repository and tag. 
For example, type:

riff {{.Command}} -i words -l python  --n uppercase --handler=process


to {{.Result}}.`

var initPythonCmd = &cobra.Command{
	Use:   "python",
	Short: "Initialize a Python function",
	Long: createCmdLong(initPythonDescription, LongVals{Process:initDefinition, Command:"init python", Result:initResult}),


	Run: func(cmd *cobra.Command, args []string) {

		initializer := NewPythonInitializer()

		err := initializer.initialize(*makeHandlerAwareOptions(cmd))
		if err != nil {
			ioutils.Error(err)
			return
		}
	},
}

func makeHandlerAwareOptions(cmd *cobra.Command) *HandlerAwareInitOptions {
	handler, _ := cmd.Flags().GetString("handler")
	options := NewHandlerAwareInitOptions(initOptions, handler)
	return options
}

/*
 * Basic sanity check that given paths exist and valid protocol given.
 * Artifact must be a regular file.
 * If artifact is given, it must be relative to the function path.
 * If function path is given as a regular file, and artifact is also given, they must reference the same path (edge case).
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

		if filepath.IsAbs(options.artifact) {
			return errors.New(fmt.Sprintf("artifact %s must be relative to function path", options.artifact))
		}

		absFilePath, err := filepath.Abs(options.functionPath)
		if err != nil {
			return err
		}

		var absArtifactPath string

		if osutils.IsDirectory(absFilePath) {
			absArtifactPath = filepath.Join(absFilePath, options.artifact)
		} else {
			absArtifactPath = filepath.Join(filepath.Dir(absFilePath), options.artifact)
		}

		if osutils.IsDirectory(absArtifactPath) {
			return errors.New(fmt.Sprintf("artifact %s must be a regular file", absArtifactPath))
		}

		absFilePathDir := absFilePath
		if !osutils.IsDirectory(absFilePath) {
			absFilePathDir = filepath.Dir(absFilePath)
		}

		if !strings.HasPrefix(filepath.Dir(absArtifactPath), absFilePathDir) {
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

	initCmd.PersistentFlags().StringVarP(&initOptions.functionName, "name", "n", "", "the functionName of the function (defaults to the functionName of the current directory)")
	initCmd.PersistentFlags().StringVarP(&initOptions.version, "version", "v", "0.0.1", "the version of the function (defaults to 0.0.1)")
	initCmd.PersistentFlags().StringVarP(&initOptions.functionPath, "filepath", "f", "", "path or directory to be used for the function resources, if a file is specified then the file's directory will be used (defaults to the current directory)")
	initCmd.PersistentFlags().StringVarP(&initOptions.protocol, "protocol", "p", "", "the protocol to use for function invocations (defaults to 'stdio' for shell and python, to 'http' for java and node)")
	initCmd.PersistentFlags().StringVarP(&initOptions.input, "input", "i", "", "the functionName of the input topic (defaults to function functionName)")
	initCmd.PersistentFlags().StringVarP(&initOptions.output, "output", "o", "", "the functionName of the output topic (optional)")
	initCmd.PersistentFlags().StringVarP(&initOptions.artifact, "artifact", "a", "", "path to the function artifact, source code or jar file")


	initCmd.AddCommand(initJavaCmd)
	initCmd.AddCommand(initJsCmd)
	initCmd.AddCommand(initNodeCmd)
	initCmd.AddCommand(initPythonCmd)
	initCmd.AddCommand(initShellCmd)

	initJavaCmd.Flags().String("handler", "", "the fully qualified class name of the function handler")
	initJavaCmd.MarkFlagRequired("handler")

	initPythonCmd.Flags().String("handler", "", "the name of the function handler")
	initPythonCmd.MarkFlagRequired("handler")

}
