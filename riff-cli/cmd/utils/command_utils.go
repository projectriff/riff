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

package utils

import (
	"bytes"
	"text/template"
	"github.com/projectriff/riff/riff-cli/global"
	"github.com/spf13/cobra"
	"fmt"
)

type Defaults struct {
	InvokerVersion string
	UserAccount    string
	Force          bool
	DryRun         bool
	Push           bool
	Version        string
}

var DefaultValues = Defaults{
	InvokerVersion: global.INVOKER_VERSION,
	UserAccount:    "current OS user",
	Force:          false,
	DryRun:         false,
	Push:           false,
	Version:        "0.0.1",
}


const (
	initResult       = `generate the required Dockerfile and resource definitions using sensible defaults`
	initDefinition   = `Generate`
	createResult     = `create the required Dockerfile and resource definitions, and apply the resources, using sensible defaults`
	createDefinition = `Create`
)

const baseDescription = `{{.Process}} the function based on the function source code specified as the filename, using the name
and version specified for the function image repository and tag. 

For example, from a directory named 'square' containing a function 'square.js', you can simply type :

    riff {{.Command}} node -f square

  or

    riff {{.Command}} node

to {{.Result}}.`

const baseJavaDescription = `{{.Process}} the function based on the function source code specified as the filename, using the artifact (jar file),
the function handler(classname), the name and version specified for the function image repository and tag. 

For example, from a maven project directory named 'greeter', type:

    riff {{.Command}} -i greetings -l java -a target/greeter-1.0.0.jar --handler=Greeter

to {{.Result}}.`

const baseCommandDescription = `{{.Process}} the function based on the executable command specified as the filename, using the name
and version specified for the function image repository and tag. 

For example, from a directory named 'echo' containing a function 'echo.sh', you can simply type :

    riff {{.Command}} -f echo

  or

    riff {{.Command}}

to {{.Result}}.`

const baseNodeDescription = `{{.Process}} the function based on the function source code specified as the filename, using the name
and version specified for the function image repository and tag.  

For example, from a directory  named 'square' containing a function 'square.js', you can simply type :

    riff {{.Command}} -f square

  or

    riff {{.Command}}

to {{.Result}}.`

const basePythonDescription = `{{.Process}} the function based on the function source code specified as the filename, handler, 
name, artifact and version specified for the function image repository and tag. 

For example, type:

    riff {{.Command}} -i words -l python -n uppercase --handler=process

to {{.Result}}.`

const baseGoDescription = `{{.Process}} the function based on a shared '.so' library file specified as the filename
and exported symbol name specified as the handler.

For example, type:

    riff {{.Command}} -i words -l go -n rot13 --handler=Encode

to {{.Result}}.`

type LongVals struct {
	Process string
	Command string
	Result  string
}

func InitCmdLong() string {
	return createCmdLong(baseDescription, LongVals{Process: initDefinition, Command: "init", Result: initResult})
}

func InitJavaCmdLong() string {
	return createCmdLong(baseJavaDescription, LongVals{Process: initDefinition, Command: "init java", Result: initResult})
}

func InitCommandCmdLong() string {
	return createCmdLong(baseCommandDescription, LongVals{Process: initDefinition, Command: "init command", Result: initResult})
}

func InitNodeCmdLong() string {
	return createCmdLong(baseNodeDescription, LongVals{Process: initDefinition, Command: "init node", Result: initResult})
}

func InitPythonCmdLong() string {
	return createCmdLong(basePythonDescription, LongVals{Process: initDefinition, Command: "init python", Result: initResult})
}

func InitGoCmdLong() string {
	return createCmdLong(baseGoDescription, LongVals{Process: initDefinition, Command: "init go", Result: initResult})
}

func CreateCmdLong() string {
	return createCmdLong(baseDescription, LongVals{Process: createDefinition, Command: "create", Result: createResult})
}

func CreateJavaCmdLong() string {
	return createCmdLong(baseJavaDescription, LongVals{Process: createDefinition, Command: "create java", Result: createResult})
}

func CreateCommandCmdLong() string {
	return createCmdLong(baseCommandDescription, LongVals{Process: createDefinition, Command: "create command", Result: createResult})
}

func CreateNodeCmdLong() string {
	return createCmdLong(baseNodeDescription, LongVals{Process: createDefinition, Command: "create node", Result: createResult})
}

func CreatePythonCmdLong() string {
	return createCmdLong(basePythonDescription, LongVals{Process: createDefinition, Command: "create python", Result: createResult})
}

func createCmdLong(longDescr string, vals LongVals) string {
	tmpl, err := template.New("longDescr").Parse(longDescr)
	if err != nil {
		panic(err)
	}

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, vals)
	if err != nil {
		panic(err)
	}

	return tpl.String()
}


// AliasFlagToSoleArg returns a cobra.PositionalArgs args validator that populates the given flag if it hasn't been yet,
// from an arg that must be set and be the only one. No args must be present if the flag has already been set.
func AliasFlagToSoleArg(flag string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		f := cmd.Flag(flag)
		if len(args) > 0 {
			if len(args) == 1 {
				if !f.Changed {
					f.Value.Set(args[0])
				} else {
					return fmt.Errorf("value for %v has already been set via the --%v flag to '%v'. " +
						"Can't set it via an argument (to '%v') as well", flag, flag, f.Value.String(), args[0])
				}
			} else {
				return fmt.Errorf("command %v expects exactly one argument", cmd.Name())
			}
		}
		return nil
	}
}

func And(functions ... cobra.PositionalArgs) cobra.PositionalArgs {
	if len(functions) == 0 {
		return nil
	}
	return func(cmd *cobra.Command, args []string) error {
		for _, f := range functions {
			if err := f(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}
