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
	"path/filepath"
	"fmt"
	"errors"
	"os"

	"github.com/projectriff/riff-cli/pkg/osutils"
)

var supportedProtocols = []string{"stdio", "http", "grpc"}
var supportedExtensions = []string{"js", "java", "py",}

//
type Initializer struct {
	initOptions InitOptionsAccessor

	functionFile string
	language     string
	extension    string
}

func NewNodeInitializer() *Initializer {
	return &Initializer{language: "node", extension: "js"}
}

func NewShellInitializer() *Initializer {
	return &Initializer{language: "shell", extension: "sh"}
}

//
type PythonInitializer struct {
	Initializer
	initOptions HandlerAwareInitOptionsAccessor
}

func NewPythonInitializer() *PythonInitializer {
	pythonInitializer := &PythonInitializer{}
	pythonInitializer.language = "python"
	pythonInitializer.extension = "py"
	return pythonInitializer
}
func (this *PythonInitializer) initialize(options HandlerAwareInitOptionsAccessor) error {
	fmt.Println("language: " + this.language)
	return nil
}

//
type JavaInitializer struct {
	Initializer
	initOptions HandlerAwareInitOptionsAccessor
}

func NewJavaInitializer() *JavaInitializer {
	javaInitializer := &JavaInitializer{}
	javaInitializer.language="java"
	javaInitializer.extension="java"
	return javaInitializer
}

func (this *JavaInitializer) initialize(options HandlerAwareInitOptionsAccessor) error {
	fmt.Println("language: " + this.language)
	return nil
}

//
type LanguageDetectingInitializer struct {
	Initializer
}

func NewLanguageDetectingInitializer() *LanguageDetectingInitializer {
	return &LanguageDetectingInitializer{}
}

func (this *LanguageDetectingInitializer) initialize(options HandlerAwareInitOptionsAccessor) error {
	functionPath, err := resolveFunctionPath(options, "")
	if err != nil {
		return err
	}

	var languageForFileExtenstion = map[string]string{
		"sh"	:  	"sh",
		"java"	: 	"java",
		"js"	:   "node",
		"py"	: 	"python",
	}

	language := languageForFileExtenstion[filepath.Ext(functionPath)[1:]]

	switch language {
	case "shell":
		NewShellInitializer().initialize(options)
	case "node":
		NewNodeInitializer().initialize(options)
	case "java":
		fmt.Println("java resources detected. Use 'riff init java' to provide additional required options")
		return nil
	case  "python":
		fmt.Println("python resources detected. Use 'riff init python' to provide additional required options")
		return nil
	default:
		//TODO: Should never get here
		return errors.New(fmt.Sprintf("unsupported language %s\n",language))
	}

	return nil

}

func (this Initializer) initialize(opts InitOptionsAccessor) error {
	fmt.Println("language: " + this.language)
	functionPath, err := resolveFunctionPath(opts, this.extension)
	if err != nil {
		return err
	}

	// Create function resources in function Path
	workDir := filepath.Dir(functionPath)
	fmt.Println(functionPath, workDir, opts.FunctionName())
	os.Chdir(workDir)
	return nil
}


func createDockerfile(workDir string, opts InitOptionsAccessor) error {
	return nil
}

//Assumes given file paths have been sanity checked and are valid
func resolveFunctionPath(options InitOptionsAccessor, ext string) (string, error) {

	functionName := options.FunctionName()
	if functionName == "" {
		functionName = filepath.Base(options.FunctionPath())
	}
	absFilePath, err := filepath.Abs(options.FunctionPath())
	if err != nil {
		return "", err
	}

	var resolvedFunctionPath string
	var functionDir string
	var functionFile string
	if osutils.IsDirectory(absFilePath) {
		if options.Artifact() == "" {
			functionFile = functionName
			functionDir = absFilePath
			if ext != "" {
				resolvedFunctionPath = filepath.Join(functionDir, fmt.Sprintf("%s.%s", functionFile, ext))
			} else {
				functionFile, err = searchForFunctionResource(functionDir, functionName)
				if err != nil {
					return "", err
				}
				resolvedFunctionPath = functionFile
			}
		} else {
			resolvedFunctionPath = filepath.Join(absFilePath, options.Artifact())
		}
	} else {
		resolvedFunctionPath = absFilePath
	}
	if !osutils.FileExists(resolvedFunctionPath) {
		return "", errors.New(fmt.Sprintf("function path %s does not exist", resolvedFunctionPath))
	}
	return resolvedFunctionPath, nil

}

func searchForFunctionResource(dir string, functionName string) (string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return "", err
	}

	foundFile := ""
	for _, f := range (files) {
		if b := filepath.Base(f); b[0:len(b)-len(filepath.Ext(f))] == functionName {
			for _, e := range supportedExtensions {
				if filepath.Ext(f) == "."+e {
					if foundFile == "" {
						foundFile = f
					} else {
						return "", errors.New(fmt.Sprintf("function file is not unique %s, %s", filepath.Base(foundFile), filepath.Base(f)))
					}
				}
			}
		}

	}

	if foundFile == "" {
		return "", errors.New(fmt.Sprintf("no function file found in path %s", dir))
	}
	return foundFile, nil
}

func (this Initializer) FunctionPath() string {
	return this.initOptions.FunctionPath()
}

