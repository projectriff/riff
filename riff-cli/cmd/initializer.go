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

	"github.com/dturanski/riff-cli/pkg/osutils"
)

var supportedProtocols = []string{"stdio", "http", "grpc"}
var supportedExtensions = []string{"js", "java", "py",}

type InitOptionsAccessor interface {
	UserAccount() string

	FunctionName() string

	Version() string

	FunctionPath() string

	Protocol() string

	Input() string

	Output() string

	RiffVersion() string

	Push() bool

	Artifact() string
}

type HandlerAwareInitOptionsAccessor interface {
	InitOptionsAccessor
	Handler() string
}

type InitOptions struct {
	InitOptionsAccessor
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

type HandlerAwareInitOptions struct {
	InitOptions
	handler string
}

func NewHandlerAwareInitOptions(options InitOptions, handler string) *HandlerAwareInitOptions {
	handlerAwareInitOptions := &HandlerAwareInitOptions{}
	handlerAwareInitOptions.functionName = options.functionName
	handlerAwareInitOptions.push = options.push
	handlerAwareInitOptions.protocol = options.protocol
	handlerAwareInitOptions.riffVersion = options.riffVersion
	handlerAwareInitOptions.functionPath = options.functionPath
	handlerAwareInitOptions.version = options.version
	handlerAwareInitOptions.userAccount = options.userAccount
	handlerAwareInitOptions.input = options.input
	handlerAwareInitOptions.output = options.output
	handlerAwareInitOptions.handler = handler

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
	return &PythonInitializer{}
}
func (this *PythonInitializer) initialize(options HandlerAwareInitOptions) error {
	return nil
}

//
type JavaInitializer struct {
	Initializer
	initOptions HandlerAwareInitOptionsAccessor
}

func NewJavaInitializer() *JavaInitializer {
	return &JavaInitializer{}
}

func (this *JavaInitializer) initialize(options HandlerAwareInitOptions) error {
	return nil
}

//
type LanguageDetectingInitializer struct {
	Initializer
}

func NewLanguageDetectingInitializer() *LanguageDetectingInitializer {
	return &LanguageDetectingInitializer{}
}

func (this *LanguageDetectingInitializer) initialize(options HandlerAwareInitOptions) error {
	return nil
}

//
//	var fileExtenstions = map[string]string{
//		"shell":  "sh",
//		"java":   "java",
//		"node":   "js",
//		"js":     "js",
//		"python": "py",
//	}
//
//
//
//
//
////	language := ""
//	for lang, ext := range fileExtenstions {
////		fileName := fmt.Sprintf("%s.%s", options.FunctionName(), ext)
////		functionFile := filepath.Join(absPath, fileName)
////		if osutils.FileExists(functionFile) {
////			language = lang
//			break
//		}
//	}
//
//
//	if (language == "") {
//		return errors.New(fmt.Sprintf("cannot find function source for function %s in directory %s", options.FunctionName(), absPath))
//	}
//
//	return nil
//}

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

//} else {
//ext := fileExtenstions[opts.Language]
//if ext == "" {
//return errors.New(fmt.Sprintf("language %s is unsupported", opts.Language))
//}
//this.options.Language = opts.Language
//
//fileName := fmt.Sprintf("%s.%s", this.options.FunctionName, ext)
//this.functionFile = filepath.Join(this.options.FunctionPath, fileName)
//if !osutils.FileExists(this.functionFile) {
//return errors.New(fmt.Sprintf("cannot find function source for function %s", this.functionFile))
//}
//}
//}

/////////////////////////

func (this Initializer) FunctionPath() string {
	return this.initOptions.FunctionPath()
}

//func (this * Initializer) SetFunctionPath(path string) error {
//	if !osutils.FileExists(path){
//		return errors.New(fmt.Sprintf("File does not exist %s", path))
//	}
//
//	this.initOptions.functionPath, _ = filepath.Abs(path)
//	return nil
//}

func (this Initializer) initialize(opts InitOptions) error {
	return nil
}

//
//		err := this.deriveOptionsFromFunctionPath(opts)
//		if err != nil{
//		return err
//	}
//
//		err = this.resolveArtifact(opts.Artifact)
//		if err != nil{
//		return err
//	}
//
//		err = this.resolveProtocol(opts.Protocol)
//		if err != nil{
//		return err
//	}
//
//		if this.options.Language == "java"{
//		if opts.Classname == ""{
//		return errors.New("'classname is required for java")
//	}
//	}
//
//
//		if opts.Input == ""{
//		this.options.Input = this.options.FunctionName
//	}
//
//		this.options.Output = opts.Output
//		this.options.UserAccount = opts.UserAccount
//		this.options.Push = opts.Push
//		this.options.RiffVersion = opts.RiffVersion
//		this.options.Version = opts.Version
//
//		fmt.Printf("function file: %s\noptions: %+v\n", this.functionFile, this.options)
//
//		return nil
//	}
//
//	func(this *Initializer) deriveOptionsFromFunctionPath(opts InitOptions) error{
//		var fileExtenstions = map[string]string{
//		"shell":    "sh",
//		"java":   "java",
//		"node":   "js",
//		"js":   "js",
//		"python":    "py",
//	}
//
//
//
//		err := this.SetFunctionPath(opts.FunctionPath)
//		if err != nil {
//			return err
//		}
//
//		if osutils.IsDirectory(this.options.FunctionPath){
//		if opts.FunctionName == ""{
//		this.options.FunctionName = filepath.Base(this.options.FunctionPath)
//	} else{
//		this.options.FunctionName = opts.FunctionName
//	}
//
//		if opts.Language == ""{
//		for lang, ext := range fileExtenstions{
//		fileName := fmt.Sprintf("%s.%s", this.options.FunctionName, ext)
//		functionFile := filepath.Join(this.options.FunctionPath, fileName)
//		if osutils.FileExists(functionFile){
//		this.options.Language = lang
//		this.functionFile = functionFile
//		break
//	}
//	}
//		if this.options.Language == ""{
//		return errors.New(fmt.Sprintf("cannot find function source for function %s in directory %s", this.options.FunctionName, this.options.FunctionPath))
//	}
//	} else{
//		ext := fileExtenstions[opts.Language]
//		if ext == ""{
//		return errors.New(fmt.Sprintf("language %s is unsupported", opts.Language))
//	}
//		this.options.Language = opts.Language
//
//		fileName := fmt.Sprintf("%s.%s", this.options.FunctionName, ext)
//		this.functionFile = filepath.Join(this.options.FunctionPath, fileName)
//		if !osutils.FileExists(this.functionFile){
//		return errors.New(fmt.Sprintf("cannot find function source for function %s", this.functionFile))
//	}
//	}
//	} else{
//		//regular file given
//		ext := filepath.Ext(this.options.FunctionPath)
//		if opts.Language == ""{
//		for lang, e := range fileExtenstions{
//		if e == ext{
//		this.options.Language = lang
//		break
//	}
//	}
//		if this.options.Language == ""{
//		return errors.New(fmt.Sprintf("cannot find function source for function %s in directory %s", this.options.FunctionName, this.options.FunctionPath))
//	}
//	} else{
//		this.options.Language = opts.Language
//		if fileExtenstions[this.options.Language] != ext{
//		fmt.Printf("WARNING non standard extension %s given for language %s. We'll see what we can do", ext, this.options.Language)
//	}
//	}
//	}
//
//
//		return nil
//	}
//
//	func(this *Initializer) resolveProtocol(protocol
//	string) error{
//		var defaultProtocols = map[string]string{
//		"shell":    "stdio",
//		"java":   "http",
//		"node":   "http",
//		"js":   "http",
//		"python":    "stdio",
//	}
//
//		var supportedProtocols = []string{"stdio", "http", "grpc"}
//
//		if protocol == ""{
//		this.options.Protocol = defaultProtocols[this.options.Language]
//	} else{
//		supported := false
//		for _, p := range supportedProtocols{
//		if protocol == p{
//		supported = true
//	}
//	}
//		if (!supported){
//		return errors.New(fmt.Sprintf("protocol %s is unsupported \n", protocol))
//	}
//		this.options.Protocol = protocol
//	}
//		return nil
//	}
//
//	func(this *Initializer) resolveArtifact(artifact
//	string) error{
//		if artifact == ""{
//		////TODO: Needs work...
//		this.options.Artifact = filepath.Base(this.functionFile)
//		return nil
//	}
//
//		//TODO: What if the artifact ext doesn't match the language?
//		if !osutils.FileExists(artifact){
//		return errors.New(fmt.Sprintf("Artifact does not exist %s", artifact))
//	}
//		return nil
//	}
