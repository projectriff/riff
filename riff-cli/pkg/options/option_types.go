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

type InitOptions struct {
	FunctionName   string
	Version        string
	FilePath       string
	Protocol       string
	Input          string
	Output         string
	Artifact       string
	InvokerVersion string
	UserAccount    string
	DryRun         bool
	Force          bool
	InvokerName    string
	Handler        string
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

type ImageOptions interface {
	GetFunctionName() string
	GetVersion() string
	GetUserAccount() string
}
