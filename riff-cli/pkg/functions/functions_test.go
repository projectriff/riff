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

package functions

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Working with file paths", func() {
	It("Should return current directory name when given '.'", func() {
		fname, err := FunctionNameFromPath(".")
		Expect(err).NotTo(HaveOccurred())
		Expect(fname).To(Equal("functions"))
	})

	It("Should cope with relative dir paths", func() {
		fname, err := FunctionNameFromPath("../../test_data/command/echo")
		Expect(err).NotTo(HaveOccurred())
		Expect(fname).To(Equal("echo"))
	})

	It("Should cope with relative file paths", func() {
		fname, err := FunctionNameFromPath("../../test_data/command/echo/echo.sh")
		Expect(err).NotTo(HaveOccurred())
		Expect(fname).To(Equal("echo"))
	})

	It("Should error on invalid path", func() {
		fname, err := FunctionNameFromPath("a/b/c/d")
		Expect(err).To(MatchError("path 'a/b/c/d' does not exist"))
		Expect(fname).To(Equal(""))
	})
})
