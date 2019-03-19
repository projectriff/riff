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

package fileutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/pkg/fileutils"
	"os"
	"strings"
)

var _ = Describe("StartsWithHomeDirAsTilde", func() {

	It("returns true when starting with ~/", func() {
		result := fileutils.StartsWithCurrentUserDirectoryAsTilde("~/", "darwin")

		Expect(result).To(BeTrue(), "tilde+forward slash should work")
	})

	It(`returns false when starting with ~\ on Mac OS`, func() {
		result := fileutils.StartsWithCurrentUserDirectoryAsTilde(`~\`, "darwin")

		Expect(result).To(BeFalse(), "tilde+backslash should not work on Mac OS")
	})

	It(`returns true when starting with ~\ on Windows`, func() {
		result := fileutils.StartsWithCurrentUserDirectoryAsTilde(`~\`, "windows")

		Expect(result).To(BeTrue(), "tilde+backslash on Windows should work")
	})

	It(`returns true when starting with ~/ on Windows`, func() {
		result := fileutils.StartsWithCurrentUserDirectoryAsTilde(`~/`, "windows")

		Expect(result).To(BeTrue(), "tilde+forward slash on Windows should work")
	})
})

var _ = Describe("ResolveTilde", func() {

	It("resolves ~/ against current user's home directory", func() {
		initialPath := "~/some/location"

		path, err := fileutils.ResolveTilde(initialPath)

		Expect(err).NotTo(HaveOccurred())
		Expect(path).NotTo(ContainSubstring("~"))
		// on windows the path separator might have changed so replace / with the OS specific separator
		Expect(path).To(HaveSuffix(strings.Replace( initialPath[2:], "/", string(os.PathSeparator), -1)))
	})

	It("returns path without tilde as is", func() {
		initialPath := "look/matilde/no/tilde"

		path, err := fileutils.ResolveTilde(initialPath)

		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(Equal(initialPath))
	})

	It("returns path with embedded tilde as is", func() {
		initialPath := "look/matilde/thereisa/~"

		path, err := fileutils.ResolveTilde(initialPath)

		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(Equal(initialPath))
	})
})
