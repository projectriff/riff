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
)

const (
	initResult       = `generate the required Dockerfile and resource definitions using sensible defaults`
	initDefinition   = `Generate`
	createResult     = `create the required Dockerfile and resource definitions, and apply the resources, using sensible defaults`
	createDefinition = `Create`
)

const baseCommandDescription = `{{.Process}} the function based on the function source code specified as the filename, using the name
  and version specified for the function image repository and tag. 

For example, from a directory named 'square' containing a function 'square.js', you can simply type :

    riff {{.Command}} node -f square

  or

    riff  {{.Command}} node

to {{.Result}}.`

const baseJavaDescription = `{{.Process}} the function based on the function source code specified as the filename, using the artifact (jar file),
  the function handler(classname), the name and version specified for the function image repository and tag. 

For example, from a maven project directory named 'greeter', type:

    riff {{.Command}} -i greetings -l java -a target/greeter-1.0.0.jar --handler=Greeter

to {{.Result}}.`

const baseShellDescription = `{{.Process}} the function based on the function script specified as the filename, using the name
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

    riff {{.Command}} -i words -l python  --n uppercase --handler=process

to {{.Result}}.`

type LongVals struct {
	Process string
	Command string
	Result  string
}

func InitCmdLong() string {
	return createCmdLong(baseCommandDescription, LongVals{Process: initDefinition, Command: "init", Result: initResult})
}

func InitJavaCmdLong() string {
	return createCmdLong(baseJavaDescription, LongVals{Process: initDefinition, Command: "init java", Result: initResult})
}

func InitShellCmdLong() string {
	return createCmdLong(baseShellDescription, LongVals{Process: initDefinition, Command: "init shell", Result: initResult})
}

func InitNodeCmdLong() string {
	return createCmdLong(baseNodeDescription, LongVals{Process: initDefinition, Command: "init node", Result: initResult})
}

func InitPythonCmdLong() string {
	return createCmdLong(basePythonDescription, LongVals{Process: initDefinition, Command: "init python", Result: initResult})
}

func CreateCmdLong() string {
	return createCmdLong(baseCommandDescription, LongVals{Process: createDefinition, Command: "create", Result: createResult})
}

func CreateJavaCmdLong() string {
	return createCmdLong(baseJavaDescription, LongVals{Process: createDefinition, Command: "create java", Result: createResult})
}

func CreateShellCmdLong() string {
	return createCmdLong(baseShellDescription, LongVals{Process: createDefinition, Command: "create shell", Result: createResult})
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
