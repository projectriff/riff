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

package main

import (
	"os"

	"github.com/golang/glog"

	"github.com/projectriff/riff/riff-cli/pkg/initializer"
	invoker "github.com/projectriff/riff/riff-cli/pkg/invoker"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/projectriff/riff/riff-cli/pkg/options"
)

const (
	riffInvokerPaths = "RIFF_INVOKER_PATHS"
	functionArtifcat = "FUNCTION_ARTIFACT"
	functionHandler  = "FUNCTION_HANDLER"
	functionFilePath = "FUNCTION_FILE_PATH"
	functionName     = "FUNCTION_NAME"
)

func main() {
	if _, ok := os.LookupEnv(riffInvokerPaths); !ok {
		glog.Fatalf("%q env var must be defined", riffInvokerPaths)
	}

	invokerOperations := invoker.Operations(kubectl.DryRunKubeCtl())
	invokers, err := invokerOperations.List()
	if err != nil {
		glog.Fatalf("Unable to get invoker: %v", err)
	}
	if len(invokers) != 1 {
		glog.Fatalf("Expected exactly one invoker, got %d", len(invokers))
	}
	invoker := invokers[0]

	initOptions := options.InitOptions{
		Artifact:     os.Getenv(functionArtifcat),
		Handler:      os.Getenv(functionHandler),
		FilePath:     os.Getenv(functionFilePath),
		FunctionName: os.Getenv(functionName),
	}
	err = initializer.Initialize(invoker, &initOptions)
	if err != nil {
		glog.Fatalf("Unable initialize function: %v", err)
	}
}
