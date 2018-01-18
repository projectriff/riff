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

type InitOptionsAccessor interface {
	FunctionName() string

	Artifact() string

	Version() string

	FunctionPath() string

	Protocol() string

	Input() string

	Output() string
}

//
type InitOptions struct {
	functionName string
	version      string
	functionPath string
	protocol     string
	input        string
	output       string
	artifact     string
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

func (this InitOptions) Artifact() string {
	return this.artifact
}

func (this InitOptions) Input() string {
	return this.input
}

func (this InitOptions) Output() string {
	return this.output
}

//

type HandlerAwareInitOptionsAccessor interface {
	InitOptionsAccessor
	Handler() string
}

type HandlerAwareInitOptions struct {
	InitOptions
	handler string
}

func NewHandlerAwareInitOptions(options InitOptionsAccessor, handler string) *HandlerAwareInitOptions {
	handlerAwareInitOptions := &HandlerAwareInitOptions{handler: handler}
	handlerAwareInitOptions.functionName = options.FunctionName()
	handlerAwareInitOptions.protocol = options.Protocol()
	handlerAwareInitOptions.functionPath = options.FunctionPath()
	handlerAwareInitOptions.version = options.Version()
	handlerAwareInitOptions.input = options.Input()
	handlerAwareInitOptions.output = options.Output()
	handlerAwareInitOptions.artifact = options.Artifact()

	return handlerAwareInitOptions
}

func (this HandlerAwareInitOptions) Handler() string {
	return this.handler
}

//
type BuildOptionsAccessor interface {
	UserAccount() string
	RiffVersion() string
	Push() bool
}

type BuildOptions struct {
	userAccount string
	push        bool
	riffVersion string
}

func (this BuildOptions) UserAccount() string {
	return this.userAccount
}

func (this BuildOptions) RiffVersion() string {
	return this.riffVersion
}

func (this BuildOptions) Push() bool {
	return this.push
}

//
type CreateOptionsAccessor interface {
	InitOptionsAccessor
	BuildOptionsAccessor
}
type CreateOptions struct {
	InitOptions
	BuildOptions
}

func NewCreateOptions(init InitOptionsAccessor, build BuildOptionsAccessor) *CreateOptions {
	createOptions := &CreateOptions{}
	createOptions.riffVersion = build.RiffVersion()
	createOptions.push = build.Push()
	createOptions.userAccount = build.UserAccount()
	createOptions.input = init.Input()
	createOptions.output = init.Output()
	createOptions.functionPath = init.FunctionPath()
	createOptions.functionName = init.FunctionName()
	createOptions.artifact = init.Artifact()
	createOptions.protocol = init.Protocol()
	createOptions.version = init.Version()

	return createOptions
}

//
type HandlerAwareCreateOptionsAccessor interface {
	HandlerAwareInitOptionsAccessor
	BuildOptionsAccessor
}

type HandlerAwareCreateOptions struct {
	HandlerAwareInitOptions
	BuildOptions
}

func NewHandlerAwareCreateOptions(init HandlerAwareInitOptionsAccessor, build BuildOptionsAccessor) *HandlerAwareCreateOptions {
	createOptions := &HandlerAwareCreateOptions{}
	createOptions.riffVersion = build.RiffVersion()
	createOptions.push = build.Push()
	createOptions.userAccount = build.UserAccount()
	createOptions.input = init.Input()
	createOptions.output = init.Output()
	createOptions.functionPath = init.FunctionPath()
	createOptions.functionName = init.FunctionName()
	createOptions.artifact = init.Artifact()
	createOptions.protocol = init.Protocol()
	createOptions.version = init.Version()
	createOptions.handler = init.Handler()

	return createOptions
}
