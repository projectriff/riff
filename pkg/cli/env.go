/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cli

var (
	cli_name     = "riff"
	cli_version  = "unknown"
	cli_gitsha   = "unknown sha"
	cli_gitdirty = ""
)

type CompiledEnv struct {
	Name     string
	Version  string
	GitSha   string
	GitDirty bool
}

var env CompiledEnv

func init() {
	// must be created inside the init function to pickup build specific params
	env = CompiledEnv{
		Name:     cli_name,
		Version:  cli_version,
		GitSha:   cli_gitsha,
		GitDirty: cli_gitdirty != "",
	}
}
