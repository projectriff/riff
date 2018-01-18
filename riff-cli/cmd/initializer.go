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

	"github.com/dturanski/riff-cli/pkg/osutils"
)

var supportedProtocols = []string{"stdio", "http", "grpc"}
var supportedExtensions = []string{"js", "java", "py",}



type InitOptionsAccessor interface {
	UserAccount() string

	FunctionName() string

	Artifact() string

	Version() string

	FunctionPath() string

	Protocol() string

	Input() string

	Output() string

	RiffVersion() string

	Push() bool
}

type HandlerAwareInitOptionsAccessor interface {
	InitOptionsAccessor
	Handler() string
}

type InitOptions struct {
	userAccount  string
	functionName string
	version      string
	functionPath string
	protocol     string
	input        string
	output       string
	riffVersion  string
	artifact     string
	push         bool
}

func (this InitOptions) UserAccount() string {
	return this.userAccount
}

func (this InitOptions) FunctionName() string {
	return this.functionName
}

func (this InitOptions) Version() string {
	return this.version
}

func (this InitOptions) FunctionPath() string {
	return this.functionPath
}

func (this InitOptions) Protocol() string {
	return this.protocol
}

func (this InitOptions) RiffVersion() string {
	return this.riffVersion
}

func (this InitOptions) Push() bool {
	return this.push
}

func (this InitOptions) Artifact() string {
	return this.artifact
}

func (this InitOptions) Input() string {
	return this.input
}

func (this InitOptions) Output() string {
	return this.output
}



type HandlerAwareInitOptions struct {
	InitOptions
	handler string
}

func NewHandlerAwareInitOptions(options InitOptions, handler string) *HandlerAwareInitOptions {
	handlerAwareInitOptions := &HandlerAwareInitOptions{handler:handler}
	handlerAwareInitOptions.functionName = options.functionName
	handlerAwareInitOptions.push = options.push
	handlerAwareInitOptions.protocol = options.protocol
	handlerAwareInitOptions.riffVersion = options.riffVersion
	handlerAwareInitOptions.functionPath = options.functionPath
	handlerAwareInitOptions.version = options.version
	handlerAwareInitOptions.userAccount = options.userAccount
	handlerAwareInitOptions.input = options.input
	handlerAwareInitOptions.output = options.output
	handlerAwareInitOptions.artifact = options.artifact

	return handlerAwareInitOptions
}

func (this HandlerAwareInitOptions) Handler() string {
	return this.handler
}

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
		NewJavaInitializer().initialize(options)
	case  "python":
		NewPythonInitializer().initialize(options)
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

