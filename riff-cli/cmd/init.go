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
	"github.com/projectriff/riff-cli/pkg/ioutils"
	"github.com/projectriff/riff-cli/pkg/options"
	"os"
	"github.com/spf13/pflag"
	"github.com/projectriff/riff-cli/pkg/initializer"
)

const (
	initResult     = `generate the required Dockerfile and resource definitions using sensible defaults`
	initDefinition = `Generate`
)

/*
 * init Command
 * TODO: Use cmd.Example
 */
const initCommandDescription = `{{.Process}} the function based on the function source code specified as the filename, using the name
and version specified for the function image repository and tag. 
For example, if you have a directory named 'square' containing a function 'square.js', you can simply type :

riff {{.Command}} node -f square

or

riff  {{.Command}} node

from the 'square' directory

to {{.Result}}.`

var initOptions options.InitOptions

var handler string

var initCmd = &cobra.Command{
	Use:   "init [language]",
	Short: "Initialize a function",
	Long:  createCmdLong(initCommandDescription, LongVals{Process: initDefinition, Command: "init", Result: initResult}),

	RunE: func(cmd *cobra.Command, args []string) error {
		err := initializer.Initialize(*newHandlerAwareOptions(cmd))
		if err != nil {
			return err
		}
		return nil
	},

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initOptions = createOptions.InitOptions
		if !initOptions.Initialized {
			initOptions = options.InitOptions{}
			var flagset pflag.FlagSet
			if cmd.Parent() == rootCmd {
				flagset = *cmd.PersistentFlags()
			} else {
				flagset = *cmd.Parent().PersistentFlags()
			}
			mergeInitOptions(flagset, &initOptions)


			if len(args) > 0 {
				if len(args) == 1 && initOptions.FunctionPath == "" {
					initOptions.FunctionPath = args[0]
				} else {
					ioutils.Errorf("Invalid argument(s) %v\n", args)
					cmd.Usage()
					os.Exit(1)
				}
			}

			err := options.ValidateAndCleanInitOptions(&initOptions)
			if err != nil {
				os.Exit(1)
			}
			initOptions.Initialized = true
		}
	},

}

/*
 * init java Command
 */
const initJavaDescription = `{{.Process}} the function based on the function source code specified as the filename, using the artifact (jar file), 
the function handler(classname), the name and version specified for the function image repository and tag. 
For example from a maven project directory named 'greeter', type:

riff {{.Command}} -i greetings -l java -a target/greeter-1.0.0.jar --handler=Greeter


to generate the required Dockerfile and resource definitions using sensible defaults.`

var initJavaCmd = &cobra.Command{
	Use:   "java",
	Short: "Initialize a Java function",
	Long:  createCmdLong(initJavaDescription, LongVals{Process: initDefinition, Command: "init java", Result: initResult}),
	RunE: func(cmd *cobra.Command, args []string) error {

		err := initializer.InitializeJava(*newHandlerAwareOptions(cmd))
		if err != nil {
			return err
		}
		return nil
	},
}
/*
 * init shell ommand
 */
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
	Long:  createCmdLong(initShellDescription, LongVals{Process: initDefinition, Command: "init shell", Result: initResult}),

	RunE: func(cmd *cobra.Command, args []string) error {
		err := initializer.InitializeShell(initOptions)
		if err != nil {
			return err
		}
		return nil
	},

}
/*
 * init node Command
 */
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
	Long:  createCmdLong(initNodeDescription, LongVals{Process: initDefinition, Command: "init node", Result: initResult}),

	RunE: func(cmd *cobra.Command, args []string) error {
		err := initializer.InitializeNode(initOptions)
		if err != nil {
			return err
		}
		return nil
	},
	Aliases: []string{"js"},
}

/*
 * init python Command
 */
const initPythonDescription = `{{.Process}} the function based on the function source code specified as the filename, handler, name, artifact
  and version specified for the function image repository and tag. 
For example, type:

riff {{.Command}} -i words -l python  --n uppercase --handler=process


to {{.Result}}.`

var initPythonCmd = &cobra.Command{
	Use:   "python",
	Short: "Initialize a Python function",
	Long:  createCmdLong(initPythonDescription, LongVals{Process: initDefinition, Command: "init python", Result: initResult}),


	RunE: func(cmd *cobra.Command, args []string) error {
		err := initializer.InitializePython(*newHandlerAwareOptions(cmd))
		if err != nil {
			return err
		}
		return nil
	},
}

func newHandlerAwareOptions(cmd *cobra.Command) *options.HandlerAwareInitOptions {
	opts := &options.HandlerAwareInitOptions{}
	opts.InitOptions = initOptions
	if handler == "" {
		handler,_ = cmd.Flags().GetString("handler")
	}
	opts.Handler = handler
	return opts
}

func init() {

	rootCmd.AddCommand(initCmd)


	createInitFlags(initCmd.PersistentFlags())

	initCmd.AddCommand(initJavaCmd)
	initCmd.AddCommand(initNodeCmd)
	initCmd.AddCommand(initPythonCmd)
	initCmd.AddCommand(initShellCmd)

	initJavaCmd.Flags().String("handler", "", "the fully qualified class name of the function handler")
	initJavaCmd.MarkFlagRequired("handler")

	initPythonCmd.Flags().String("handler", "", "the name of the function handler")
	initPythonCmd.MarkFlagRequired("handler")

}
