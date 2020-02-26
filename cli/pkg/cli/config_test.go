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

package cli_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"
	"github.com/projectriff/cli/pkg/cli"
)

func TestNewDefaultConfig_Stdio(t *testing.T) {
	config := cli.NewDefaultConfig()

	if expected, actual := os.Stdin, config.Stdin; expected != actual {
		t.Errorf("Expected stdin to be %v, actually %v", expected, actual)
	}
	if expected, actual := os.Stdout, config.Stdout; expected != actual {
		t.Errorf("Expected stdout to be %v, actually %v", expected, actual)
	}
	if expected, actual := os.Stderr, config.Stderr; expected != actual {
		t.Errorf("Expected stderr to be %v, actually %v", expected, actual)
	}
}

func TestNewDefaultConfig_CompiledEnv(t *testing.T) {
	expected := cli.NewDefaultConfig().CompiledEnv
	actual := cli.CompiledEnv{
		Name:     "riff",
		Version:  "unknown",
		GitSha:   "unknown sha",
		GitDirty: false,
		Runtimes: map[string]bool{
			"core": true,
		},
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("Unexpected env (-expected, +actual): %s", diff)
	}
}

func TestConfig_Print(t *testing.T) {
	noColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = noColor }()

	config := cli.NewDefaultConfig()

	tests := []struct {
		name    string
		format  string
		args    []interface{}
		printer func(format string, a ...interface{}) (n int, err error)
		stdout  string
		stderr  string
	}{{
		name:    "Printf",
		format:  "%s",
		args:    []interface{}{"hello"},
		printer: config.Printf,
		stdout:  "hello",
	}, {
		name:    "Eprintf",
		format:  "%s",
		args:    []interface{}{"hello"},
		printer: config.Eprintf,
		stderr:  "hello",
	}, {
		name:    "Infof",
		format:  "%s",
		args:    []interface{}{"hello"},
		printer: config.Infof,
		stdout:  cli.InfoColor.Sprint("hello"),
	}, {
		name:    "Einfof",
		format:  "%s",
		args:    []interface{}{"hello"},
		printer: config.Einfof,
		stderr:  cli.InfoColor.Sprint("hello"),
	}, {
		name:    "Successf",
		format:  "%s",
		args:    []interface{}{"hello"},
		printer: config.Successf,
		stdout:  cli.SuccessColor.Sprint("hello"),
	}, {
		name:    "Esuccessf",
		format:  "%s",
		args:    []interface{}{"hello"},
		printer: config.Esuccessf,
		stderr:  cli.SuccessColor.Sprint("hello"),
	}, {
		name:    "Errorf",
		format:  "%s",
		args:    []interface{}{"hello"},
		printer: config.Errorf,
		stdout:  cli.ErrorColor.Sprint("hello"),
	}, {
		name:    "Eerrorf",
		format:  "%s",
		args:    []interface{}{"hello"},
		printer: config.Eerrorf,
		stderr:  cli.ErrorColor.Sprint("hello"),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			config.Stdout = stdout
			config.Stderr = stderr

			_, err := test.printer(test.format, test.args...)

			if err != nil {
				t.Errorf("Expected no error, actually %q", err)
			}
			if expected, actual := test.stdout, stdout.String(); expected != actual {
				t.Errorf("Expected stdout to be %q, actually %q", expected, actual)
			}
			if expected, actual := test.stderr, stderr.String(); expected != actual {
				t.Errorf("Expected stderr to be %q, actually %q", expected, actual)
			}
		})
	}
}
