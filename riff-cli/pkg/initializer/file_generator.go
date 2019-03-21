/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *        https://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package initializer

import (
	"path/filepath"

	"github.com/projectriff/riff/riff-cli/pkg/templateutils"

	"github.com/projectriff/riff/riff-cli/pkg/options"
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
)

type fileTokens struct {
	FunctionName   string
	Artifact       string
	ArtifactBase   string
	InvokerVersion string
	Handler        string
	filePath       string
}

func (t fileTokens) FileExists(fileName string) bool {
	filePath := t.filePath
	if !osutils.IsDirectory(filePath) {
		filePath = filepath.Dir(filePath)
	}
	return osutils.FileExists(filepath.Join(filePath, fileName))
}

func generateFileContents(tmpl string, name string, opts options.InitOptions) (string, error) {
	tokens := generateFileTokens(opts)
	return templateutils.Apply(tmpl, name, tokens)
}

func generateFileTokens(opts options.InitOptions) fileTokens {
	tokens := fileTokens{}
	tokens.FunctionName = opts.FunctionName
	tokens.Artifact = opts.Artifact
	tokens.ArtifactBase = filepath.Base(opts.Artifact)
	tokens.InvokerVersion = opts.InvokerVersion
	tokens.Handler = opts.Handler
	tokens.filePath = opts.FilePath
	return tokens
}
