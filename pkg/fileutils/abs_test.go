/*
 * Copyright 2019 The original author or authors
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
	"os"
	"path/filepath"

	"github.com/projectriff/riff/pkg/test_support"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/pkg/fileutils"
)

var _ = Describe("IsAbsFile", func() {

	var (
		path          string
		absolute      bool
		canonicalPath string
		err           error
	)

	JustBeforeEach(func() {
		absolute, canonicalPath, err = fileutils.IsAbsFile(path)
	})

	checkAbsolute := func() {
		It("should return successfully", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("should treat the path as absolute", func() {
			Expect(absolute).To(BeTrue())
		})

		It("should return the path as the canonical form", func() {
			Expect(canonicalPath).To(Equal(path))
		})
	}

	Context("when the input path is a http URL", func() {
		BeforeEach(func() {
			path = "http://projectriff.io"
		})

		checkAbsolute()
	})

	Context("when the input path is a https URL", func() {
		BeforeEach(func() {
			path = "https://projectriff.io"
		})

		checkAbsolute()
	})

	Context("when the input path is a URL with an unsupported scheme", func() {
		BeforeEach(func() {
			path = "ftp://projectriff.io"
		})

		It("should return a suitable error", func() {
			Expect(err).To(MatchError("unsupported URL scheme ftp in ftp://projectriff.io"))
		})
	})

	Context("when the input path is an absolute file path", func() {
		BeforeEach(func() {
			var err error
			path, err = os.Getwd()
			Expect(err).NotTo(HaveOccurred())
		})

		checkAbsolute()
	})

	Context("when the input path is a file URL", func() {
		var baseDir string

		BeforeEach(func() {
			var err error
			baseDir, err = os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			path = test_support.FileURL(baseDir)
		})

		It("should return successfully", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("should treat the path as absolute", func() {
			Expect(absolute).To(BeTrue())
		})

		It("should return the file path of the URL as the canonical form", func() {
			Expect(canonicalPath).To(Equal(baseDir))
		})
	})

	Context("when the input path is a relative file path", func() {
		BeforeEach(func() {
			path = "a/b"
		})

		It("should return successfully", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("should treat the path as relative", func() {
			Expect(absolute).To(BeFalse())
		})
	})
})

var _ = Describe("AbsFile", func() {

	var (
		path    string
		baseDir string
		absFile string
		err     error
	)

	BeforeEach(func() {
		var err error
		baseDir, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		absFile, err = fileutils.AbsFile(path, baseDir)
	})

	checkAbsolute := func() {
		It("should return successfully", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return the path as the absolute path", func() {
			Expect(absFile).To(Equal(path))
		})
	}

	Context("when the input path is a http URL", func() {
		BeforeEach(func() {
			path = "http://projectriff.io"
		})

		checkAbsolute()
	})

	Context("when the input path is a https URL", func() {
		BeforeEach(func() {
			path = "https://projectriff.io"
		})

		checkAbsolute()
	})

	Context("when the input path is a URL with an unsupported scheme", func() {
		BeforeEach(func() {
			path = "ftp://projectriff.io"
		})

		It("should return a suitable error", func() {
			Expect(err).To(MatchError("unsupported URL scheme ftp in ftp://projectriff.io"))
		})
	})

	Context("when the input path is an absolute file path", func() {
		BeforeEach(func() {
			var err error
			path, err = os.Getwd()
			Expect(err).NotTo(HaveOccurred())
		})

		checkAbsolute()
	})

	Context("when the input path is a file URL", func() {
		BeforeEach(func() {
			path = test_support.FileURL(baseDir)
		})

		It("should return successfully", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return the file path of the URL", func() {
			Expect(absFile).To(Equal(baseDir))
		})
	})

	Context("when the input path is a relative file path", func() {
		BeforeEach(func() {
			path = filepath.Join("a", "b")
		})

		It("should return successfully", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return the absolute path of the file", func() {
			Expect(absFile).To(Equal(filepath.Join(baseDir, "a", "b")))
		})
	})
})
