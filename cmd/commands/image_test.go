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
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/core/mocks"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("The riff image relocate command", func() {
	Context("when given wrong args or flags", func() {
		var (
			mockClient core.ImageClient
			cc         *cobra.Command
		)

		BeforeEach(func() {
			mockClient = nil
			cc = commands.ImageRelocate(&mockClient)
		})

		It("should fail with a positional argument", func() {
			cc.SetArgs([]string{"a", "--file=x", "--images=x", "--output=x", "--registry=x", "--registry-user=x"})
			err := cc.Execute()
			Expect(err).To(MatchError("the `image relocate` command does not support positional arguments"))
		})

		It("should fail with invalid registry hostname", func() {
			cc.SetArgs([]string{"--file=x", "--images=x", "--output=x", "--registry=.host", "--registry-user=x"})
			err := cc.Execute()
			Expect(err).To(MatchError(HavePrefix("invalid registry hostname: ")))
			Expect(err).To(MatchError(ContainSubstring("must start and end with an alphanumeric character")))
		})

		It("should fail without required flags", func() {
			cc.SetArgs([]string{})
			err := cc.Execute()
			Expect(err).To(MatchError(`required flag(s) "images", "output", "registry", "registry-user" not set`))
		})

		It("should fail when both file and manifest are set", func() {
			cc.SetArgs([]string{"--file=x", "--manifest=x", "--images=x", "--output=x", "--registry=host", "--registry-user=x"})
			err := cc.Execute()
			Expect(err).To(MatchError("when --manifest is set, --file should not be set"))
		})
	})

	Context("when given suitable flags", func() {
		var (
			client core.ImageClient
			asMock *mocks.ImageClient
			sc     *cobra.Command
			err    error
		)

		BeforeEach(func() {
			client = new(mocks.ImageClient)
			asMock = client.(*mocks.ImageClient)

			sc = commands.ImageRelocate(&client)
		})

		JustBeforeEach(func() {
			err = sc.Execute()
		})

		AfterEach(func() {
			asMock.AssertExpectations(GinkgoT())

		})

		Context("when a single file is specified", func() {
			BeforeEach(func() {
				sc.SetArgs([]string{"--file=file", "--images=images", "--output=output", "--registry=registry", "--registry-user=registryuser"})

				o := core.RelocateImagesOptions{
					SingleFile:   "file",
					Manifest:     "",
					Images:       "images",
					Output:       "output",
					Registry:     "registry",
					RegistryUser: "registryuser",
				}

				asMock.On("RelocateImages", o).Return(nil)
			})

			It("should invoke core.Client correctly and succeed", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when a manifest is specified", func() {
			BeforeEach(func() {
				sc.SetArgs([]string{"--manifest=manifest", "--images=images", "--output=output", "--registry=registry", "--registry-user=registryuser"})

				o := core.RelocateImagesOptions{
					Manifest:     "manifest",
					SingleFile:   "",
					Images:       "images",
					Output:       "output",
					Registry:     "registry",
					RegistryUser: "registryuser",
				}

				asMock.On("RelocateImages", o).Return(nil)
			})

			It("should invoke core.Client correctly and succeed", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when neither file nor manifest is specified", func() {
			BeforeEach(func() {
				sc.SetArgs([]string{"--images=images", "--output=output", "--registry=registry", "--registry-user=registryuser"})

				o := core.RelocateImagesOptions{
					Manifest:     "manifest.yaml",
					SingleFile:   "",
					Images:       "images",
					Output:       "output",
					Registry:     "registry",
					RegistryUser: "registryuser",
				}

				asMock.On("RelocateImages", o).Return(nil)
			})

			It("should invoke core.Client correctly and succeed", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when core.Client returns an error", func() {
			var testErr error

			BeforeEach(func() {
				sc.SetArgs([]string{"--file=file", "--images=images", "--output=output", "--registry=registry", "--registry-user=registryuser"})

				testErr = fmt.Errorf("some error")
				asMock.On("RelocateImages", mock.Anything).Return(testErr)
			})

			It("should propagate the error", func() {
				Expect(err).To(MatchError(testErr))
			})
		})
	})
})
