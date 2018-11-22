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

package fileutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/pkg/fileutils"
	"os"
	"path/filepath"
)

var _ = Describe("Read", func() {

	var (
		file    string
		base    string
		content []byte
		err     error
	)

	JustBeforeEach(func() {
		content, err = fileutils.Read(file, base)
	})

	Context("when file is a URL", func() {
		BeforeEach(func() {
			file = getwdAsURL() + "/fixtures/file.txt"

			base = "" // irrelevant when file is absolute
		})

		It("should read the file content", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("contents"))
		})
	})

	Context("when file is an absolute path", func() {
		BeforeEach(func() {
			cwd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			file = filepath.Join(cwd, "fixtures", "file.txt")

			base = "" // irrelevant when file is absolute
		})

		It("should read the file content", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("contents"))
		})
	})

	Context("when file is a relative path", func() {
		BeforeEach(func() {
			file = filepath.Join("fixtures", "file.txt")

			base = "" // irrelevant when file is absolute
		})

		Context("when base is empty", func() {
			BeforeEach(func() {
				base = ""
			})

			It("should read the file content", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal("contents"))
			})
		})

		Context("when base is a URL", func() {
			BeforeEach(func() {
				base = getwdAsURL()
			})

			It("should read the file content", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal("contents"))
			})
		})

		Context("when base is an absolute file path", func() {
			BeforeEach(func() {
				var err error
				base, err = os.Getwd()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should read the file content", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal("contents"))
			})
		})

		Context("when base is a relative file path", func() {
			BeforeEach(func() {
				base = "fixtures"

				file = "file.txt"
			})

			It("should read the file content", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal("contents"))
			})
		})

		Context("when base is a relative file path expressed using dot", func() {
			BeforeEach(func() {
				base = "./fixtures"

				file = "file.txt"
			})

			It("should read the file content", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal("contents"))
			})
		})
	})

})

func getwdAsURL() string {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	return "file:///" + filepath.ToSlash(cwd)
}
