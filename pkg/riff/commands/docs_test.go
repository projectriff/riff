/*
 * Copyright 2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commands_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/riff/commands"
	rifftesting "github.com/projectriff/riff/pkg/testing"
)

func TestDocsOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "valid",
			Options: &commands.DocsOptions{
				Directory: "docs",
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid",
			Options: &commands.DocsOptions{
				Directory: "",
			},
			ExpectFieldError: cli.ErrMissingField(cli.DirectoryFlagName),
		},
	}

	table.Run(t)
}

func TestDocsCommand(t *testing.T) {
	dir, err := ioutil.TempDir("", "docs")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	table := rifftesting.CommandTable{
		{
			Name: "generate docs",
			Args: []string{cli.DirectoryFlagName, dir},
			Prepare: func(t *testing.T, c *cli.Config) error {
				// ensure the directory is empty
				os.RemoveAll(dir)
				return nil
			},
			CleanUp: func(t *testing.T, c *cli.Config) error {
				files, err := ioutil.ReadDir(dir)
				if err != nil {
					t.Error(err)
				}
				// expect a single file because the docs command is currently the root command
				if expected, actual := 1, len(files); expected != actual {
					t.Errorf("expected %d file, found %d files", expected, actual)
				} else if expected, actual := "docs.md", files[0].Name(); expected != actual {
					t.Errorf("expected file name %q, found %q", expected, actual)
				}
				return os.RemoveAll(dir)
			},
		},
	}

	table.Run(t, commands.NewDocsCommand)
}
