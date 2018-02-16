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
	"path/filepath"

	"github.com/projectriff/riff-cli/pkg/options"
)

func ResolveOptions(functionArtifact string, language string, opts *options.InitOptions) {

	if opts.Input == "" {
		opts.Input = opts.FunctionName
	}

	if opts.Artifact == "" {
		opts.Artifact = filepath.Base(functionArtifact)
	}

	protocolForLanguage := map[string]string{
		"shell": "grpc",
		"java":  "pipes",
		"js":    "grpc",
		"node":  "grpc",
		"python": "stdio",
	}

	if opts.Protocol == "" {
		opts.Protocol = protocolForLanguage[language]
	}
}
