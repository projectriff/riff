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
	"github.com/projectriff/riff/pkg/test_support"
	"path/filepath"
)

var _ = Describe("Dir", func() {
	var (
		file string
		dir  string
		err  error
	)

	JustBeforeEach(func() {
		dir, err = fileutils.Dir(file)
	})

	Context("when file is a URL", func() {
		BeforeEach(func() {
			file = "file:///dir/file"
		})

		It("should return the directory of the file", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(dir).To(Equal("file:///dir"))
		})
	})

	Context("when file is a relative path", func() {
		BeforeEach(func() {
			file = filepath.Join("dir", "file")
		})

		It("should return the directory of the file", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(dir).To(Equal("dir"))
		})
	})

	Context("when file is an absolute path", func() {
		var work string
		BeforeEach(func() {
			work = test_support.CreateTempDir()
			file = filepath.Join(work, "file")
		})

		AfterEach(func() {
			test_support.CleanupDirs(GinkgoT(), work)
		})

		It("should return the directory of the file", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(dir).To(Equal(work))
		})
	})
})
