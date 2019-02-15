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
	"errors"
	"fmt"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/cmd/commands"
	. "github.com/spf13/cobra"
)

var _ = Describe("The cobra extensions", func() {

	Context("the argument validation framework", func() {

		It("should not fail if an optional argument to validate is not provided", func() {
			command := &Command{
				Use: "some-command",
				Args: commands.ArgValidationConjunction(
					MaximumNArgs(1),
					commands.OptionalAtPosition(1, func(_ *Command, _ string) error {
						return errors.New("should not be called")
					}),
				),
				RunE: func(cmd *Command, args []string) error {
					return nil
				},
			}

			Expect(command.Execute()).NotTo(HaveOccurred())
		})

		It("should fail if one argument is invalid", func() {
			command := &Command{
				Use: "some-command",
				Args: commands.ArgValidationConjunction(
					MinimumNArgs(1),
					commands.StartingAtPosition(0, evenNumberValidator),
				),
				RunE: func(cmd *Command, args []string) error {
					return nil
				},
			}

			command.SetArgs([]string{"2", "3", "7", "6"})

			Expect(command.Execute()).To(MatchError("3 should be even\n7 should be even\n"))
		})

		It("should fail if a not blank argument is not set", func() {
			var foobar string
			command := &Command{
				Use:     "some-command",
				PreRunE: commands.FlagsValidatorAsCobraRunE(commands.NotBlank("foobar")),
				RunE: func(cmd *Command, args []string) error {
					return nil
				},
			}
			command.Flags().StringVar(&foobar, "foobar", "", "some meaningful flag")

			Expect(command.Execute()).To(MatchError("flag --foobar cannot be empty"))
		})

		// note: the unset value case is handled separately by cobra
		It("should fail if a not blank argument's value is... blank", func() {
			var foobar string
			command := &Command{
				Use:     "some-command",
				PreRunE: commands.FlagsValidatorAsCobraRunE(commands.NotBlank("foobar")),
				RunE: func(cmd *Command, args []string) error {
					return nil
				},
			}
			command.Flags().StringVar(&foobar, "foobar", "", "some meaningful flag")

			command.SetArgs([]string{"--foobar", ""})

			Expect(command.Execute()).To(MatchError("flag --foobar cannot be empty"))
		})

	})
})

func evenNumberValidator(_ *Command, s string) error {
	integer, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	if integer%2 != 0 {
		return fmt.Errorf("%d should be even", integer)
	}
	return nil
}
