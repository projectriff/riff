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
	FilePath     string
	Protocol     string
	Input        string
	Output       string
	Artifact     string
	RiffVersion  string
	UserAccount  string
	Initialized  bool
	DryRun       bool
	Force        bool
	Handler      string
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
	FilePath     string
	FunctionName string
	Version      string
	RiffVersion  string
	UserAccount  string
	Namespace    string
	Push         bool
	DryRun       bool
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

func (this BuildOptions) GetNamespace() string {
	return this.Namespace
}

type ApplyOptions struct {
	FilePath  string
	Namespace string
	DryRun    bool
}

type DeleteOptions struct {
	FilePath     string
	FunctionName string
	Namespace    string
	DryRun       bool
	All          bool
}

func (this DeleteOptions) GetAll() bool {
	return this.All
}

type CreateOptions struct {
	InitOptions
	Namespace    string
	Push        bool
}

type DeleteAllOptions struct {
	FunctionName string
	FilePath     string
	Namespace    string
	DryRun       bool
	All          bool
	Initialized  bool
}

type ImageOptions interface {
	GetFunctionName() string
	GetVersion()      string
	GetUserAccount()  string
}

func GetApplyOptions(opts CreateOptions) ApplyOptions {
	return ApplyOptions{FilePath:opts.FilePath, Namespace:opts.Namespace, DryRun:opts.DryRun}
}

func GetDeleteOptions(opts DeleteAllOptions) DeleteOptions {
	return DeleteOptions{
		FilePath:     opts.FilePath,
		FunctionName: opts.FunctionName,
		Namespace:    opts.Namespace,
		DryRun:       opts.DryRun, All: opts.All}
}

func GetBuildOptions(opts CreateOptions) BuildOptions {
	return BuildOptions{
		FilePath:     opts.FilePath,
		FunctionName: opts.FunctionName,
		Version:      opts.Version,
		RiffVersion:  opts.RiffVersion,
		UserAccount:  opts.UserAccount,
		Push:         opts.Push,
		DryRun:       opts.DryRun,
	}
}



