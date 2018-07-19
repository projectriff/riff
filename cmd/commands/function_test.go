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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff-cli/cmd/commands"
	"github.com/spf13/cobra"
)

var _ = Describe("The function command", func() {

	Context("when using 'function create'", func() {
		var (
			fnCreate *cobra.Command
		)
		It("should complain about flags misuse", func() {
			fnCreate = commands.FunctionCreate(nil)
			fnCreate.SetArgs([]string{"node", "my-fn", "--from-image", "foo", "--to-image", "bar"})
			Expect(fnCreate.Execute()).To(MatchError("at most one of --from-image, --to-image must be set"))

			fnCreate = commands.FunctionCreate(nil)
			fnCreate.SetArgs([]string{"node", "my-fn", "--from-image", "bar", "--input", "foo"})
			Expect(fnCreate.Execute()).To(MatchError("when --input is set, at least one of --bus, --cluster-bus must be set"))

			fnCreate = commands.FunctionCreate(nil)
			fnCreate.SetArgs([]string{"node", "my-fn", "--from-image", "bar", "--input", "foo", "--bus", "b", "--cluster-bus", "cb"})
			Expect(fnCreate.Execute()).To(MatchError("when --input is set, at most one of --bus, --cluster-bus must be set"))

			fnCreate = commands.FunctionCreate(nil)
			fnCreate.SetArgs([]string{"node", "my-fn", "--from-image", "bar", "--bus", "b"})
			Expect(fnCreate.Execute()).To(MatchError("when --input is not set, --bus should not be set"))

			fnCreate = commands.FunctionCreate(nil)
			fnCreate.SetArgs([]string{"node", "my-fn", "--from-image", "bar", "--cluster-bus", "cb"})
			Expect(fnCreate.Execute()).To(MatchError("when --input is not set, --cluster-bus should not be set"))

			fnCreate = commands.FunctionCreate(nil)
			fnCreate.SetArgs([]string{"node", "my-fn", "--from-image", "bar", "--git-revision", "foo"})
			Expect(fnCreate.Execute()).To(MatchError("when --to-image is not set, --git-revision should not be set"))

			fnCreate = commands.FunctionCreate(nil)
			fnCreate.SetArgs([]string{"node", "my-fn", "--from-image", "bar", "--git-repo", "foo"})
			Expect(fnCreate.Execute()).To(MatchError("when --to-image is not set, --git-repo should not be set"))

			fnCreate = commands.FunctionCreate(nil)
			fnCreate.SetArgs([]string{"node", "my-fn", "--from-image", "bar", "--handler", "foo"})
			Expect(fnCreate.Execute()).To(MatchError("when --to-image is not set, --handler should not be set"))

			fnCreate = commands.FunctionCreate(nil)
			fnCreate.SetArgs([]string{"node", "my-fn", "--from-image", "bar", "--artifact", "foo"})
			Expect(fnCreate.Execute()).To(MatchError("when --to-image is not set, --artifact should not be set"))

		})
	})
})
