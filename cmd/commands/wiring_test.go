/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commands_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/cmd/commands"
	"github.com/spf13/cobra"
)

var _ = Describe("`riff` root command", func() {
	Context("should wire subcommands", func() {
		var rootCommand *cobra.Command

		BeforeEach(func() {
			rootCommand = commands.CreateAndWireRootCommand()
		})

		It("including `riff subscription`", func() {
			errMsg := "`%s` should be wired to root command"
			Expect(find(rootCommand, "subscription")).NotTo(BeNil(), fmt.Sprintf(errMsg, "subscription"))
			Expect(find(rootCommand, "subscription", "create")).NotTo(BeNil(), fmt.Sprintf(errMsg, "subscription create"))
			Expect(find(rootCommand, "subscription", "delete")).NotTo(BeNil(), fmt.Sprintf(errMsg, "subscription delete"))
			Expect(find(rootCommand, "subscription", "list")).NotTo(BeNil(), fmt.Sprintf(errMsg, "subscription list"))
		})

	})

})

func find(command *cobra.Command, names ...string) *cobra.Command {
	cmd, unmatchedArgs, err := command.Find(names)
	if err != nil || len(unmatchedArgs) > 0 {
		return nil
	}
	return cmd
}
