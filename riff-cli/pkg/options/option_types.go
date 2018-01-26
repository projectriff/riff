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
package options

var SupportedProtocols = []string{"stdio", "http", "grpc"}

type InitOptions struct {
	FunctionName string
	Version      string
	FunctionPath string
	Protocol     string
	Input        string
	Output       string
	Artifact     string
	RiffVersion  string
	UserAccount  string
	Initialized  bool
	DryRun		 bool
	Force		 bool
	Handler 	string
}

func (this InitOptions) GetFunctionName() string {
	return this.FunctionName
}

func (this InitOptions) GetVersion() string {
	return this.Version
}

func (this InitOptions) GetUserAccount() string {
	return this.UserAccount
}

type BuildOptions struct {
	FunctionPath string
	FunctionName string
	Version      string
	RiffVersion  string
	UserAccount  string
	Push         bool
	DryRun		 bool
}

func (this BuildOptions) GetFunctionName() string {
	return this.FunctionName
}

func (this BuildOptions) GetVersion() string {
	return this.Version
}

func (this BuildOptions) GetUserAccount() string {
	return this.UserAccount
}

type ApplyOptions struct {
	FunctionPath string
	DryRun		 bool
}

type DeleteOptions struct {
	FunctionPath string
	FunctionName string
	DryRun		 bool
	All		     bool
}

func (this DeleteOptions) GetAll() bool {
	return this.All
}

type CreateOptions struct {
	InitOptions
	Push        bool
}

type AllOptions struct {
	InitOptions
	Push         bool
	All		     bool
}

type ImageOptions interface {
	GetFunctionName() string
	GetVersion()      string
	GetUserAccount()  string
}

func GetApplyOptions(opts CreateOptions) ApplyOptions {
	return ApplyOptions{FunctionPath:opts.FunctionPath, DryRun:opts.DryRun}
}

func GetDeleteOptions(opts AllOptions) DeleteOptions {
	return DeleteOptions{FunctionPath:opts.FunctionPath, FunctionName:opts.FunctionName, DryRun:opts.DryRun, All:opts.All}
}

func GetBuildOptions(opts CreateOptions) BuildOptions {
	return BuildOptions{
		FunctionPath:opts.FunctionPath,
		FunctionName:opts.FunctionName,
		Version:opts.Version,
		RiffVersion:opts.RiffVersion,
		UserAccount:opts.UserAccount,
		Push:opts.Push,
		DryRun:opts.DryRun,
	}
}



