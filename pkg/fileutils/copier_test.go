/*
 * Copyright 2014-2018 The original author or authors
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
	"io/ioutil"
	"os"
	"path/filepath"
)

var _ = Describe("Copier", func() {
	var (
		symlinkSupported bool
		copier           fileutils.Copier
		tempDir          string
		source           string
		target           string
		err              error
	)

	// Detect whether symlinks can be created in order to determine which tests to run.
	{
		tempDir := test_support.CreateTempDir()
		err := os.Symlink("a", filepath.Join(tempDir, "symlink"))
		symlinkSupported = err == nil
	}

	BeforeEach(func() {
		checker := fileutils.NewChecker() // use a real checker to avoid mock setup
		copier = fileutils.NewCopier(ioutil.Discard, checker)

		tempDir = test_support.CreateTempDir()
	})

	JustBeforeEach(func() {
		err = copier.Copy(target, source)
	})

	AfterEach(func() {
		test_support.CleanupDirs(GinkgoT(), tempDir)
	})

	Context("when the source is a file", func() {
		BeforeEach(func() {
			source = test_support.CreateFile(tempDir, "src.file")
			target = filepath.Join(tempDir, "target.file")
		})

		It("should copy the file", func() {
			Expect(err).NotTo(HaveOccurred())

			checkFile(target, "test contents")
		})

		Context("when the target is the same file", func() {
			BeforeEach(func() {
				target = source
			})

			It("should succeed and not corrupt the file contents", func() {
				Expect(err).NotTo(HaveOccurred())

				checkFile(target, "test contents")
			})
		})

		Context("when the target is a different file", func() {
			BeforeEach(func() {
				target = test_support.CreateFile(tempDir, "target.file", "different contents")
			})

			It("should succeed and overwrite the file contents", func() {
				Expect(err).NotTo(HaveOccurred())

				checkFile(target, "test contents")
			})
		})
	})

	Context("when the source is a file with a specific mode", func() {
		var expectedPermissions string

		BeforeEach(func() {
			source = test_support.CreateFileWithMode(tempDir, "src.file", os.FileMode(0642))
			expectedPermissions = test_support.FileMode(source).String() // the permissions are surprising on Windows

			target = filepath.Join(tempDir, "target.file")
		})

		It("should copy the file mode", func() {
			Expect(err).NotTo(HaveOccurred())

			modeString := test_support.FileMode(target).String()
			Expect(modeString).To(Equal(expectedPermissions))
		})
	})

	Context("when the target does not exist", func() {
		BeforeEach(func() {
			target = filepath.Join(tempDir, "target")
		})

		Context("when the source is a directory with contents", func() {
			BeforeEach(func() {
				source = filepath.Join(tempDir, "source")
				err := os.Mkdir(source, os.FileMode(0777))
				Expect(err).NotTo(HaveOccurred())

				test_support.CreateFile(source, "file1")
				test_support.CreateFile(source, "file2")
			})

			It("should copy the directory and its contents", func() {
				Expect(err).NotTo(HaveOccurred())

				checkDirectory(target)
				checkFile(filepath.Join(target, "file1"), "test contents")
				checkFile(filepath.Join(target, "file2"), "test contents")
			})
		})

		Context("when the source is a directory with nested contents", func() {
			BeforeEach(func() {
				source = filepath.Join(tempDir, "source")
				err := os.Mkdir(source, os.FileMode(0777))
				Expect(err).NotTo(HaveOccurred())

				subDir := filepath.Join(source, "subdir")
				err = os.Mkdir(subDir, os.FileMode(0777))
				Expect(err).NotTo(HaveOccurred())

				test_support.CreateFile(subDir, "file1")
				test_support.CreateFile(subDir, "file2")
			})

			It("should copy the directory and its contents", func() {
				Expect(err).NotTo(HaveOccurred())

				checkDirectory(target)
				checkFile(filepath.Join(target, "subdir", "file1"), "test contents")
				checkFile(filepath.Join(target, "subdir", "file2"), "test contents")
			})
		})

		Context("when the source is a directory with a specific mode", func() {
			var expectedPermissions string
			BeforeEach(func() {
				source = test_support.CreateDirWithMode(tempDir, "src.dir", os.FileMode(0642))

				expectedPermissions = test_support.FileMode(source).String() // the permissions are surprising on Windows
			})

			It("should copy the mode", func() {
				Expect(err).NotTo(HaveOccurred())

				modeString := test_support.FileMode(target).String()
				Expect(modeString).To(Equal(expectedPermissions))
			})
		})

		if symlinkSupported {
			Context("when the source is a directory containing an internal symbolic link to a directory", func() {
				BeforeEach(func() {
					/*
						   Create a directory structure inside tempDir like this:

							source/ <------+
								file1      |
								dir1/      |
									link --+

					*/

					source = test_support.CreateDir(tempDir, "source")

					test_support.CreateFile(source, "file1")
					dir1 := test_support.CreateDir(source, "dir1")

					srcLink := filepath.Join(dir1, "link")
					err := os.Symlink(source, srcLink)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should copy the symbolic link", func() {
					Expect(err).NotTo(HaveOccurred())

					targetDir1 := filepath.Join(target, "dir1")
					targetLink := filepath.Join(targetDir1, "link")

					linkTarget, err := os.Readlink(targetLink)
					Expect(err).NotTo(HaveOccurred())
					Expect(linkTarget).To(Equal(".."))

					Expect(test_support.SameFile(target, filepath.Join(targetDir1, linkTarget))).To(BeTrue())
				})
			})

			Context("when the source is a directory containing an internal symbolic link to a file", func() {
				BeforeEach(func() {
					/*
						   Create a directory structure inside tempDir like this:

							source/
								file1 <----+
								link ------+

					*/

					source = test_support.CreateDir(tempDir, "source")

					file1 := test_support.CreateFile(source, "file1")
					fileLink := filepath.Join(source, "link")
					err := os.Symlink(file1, fileLink)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should copy the symbolic link", func() {
					Expect(err).NotTo(HaveOccurred())

					targetFile1 := filepath.Join(target, "file1")
					targetLink := filepath.Join(target, "link")

					linkTarget, err := os.Readlink(targetLink)
					Expect(err).NotTo(HaveOccurred())
					Expect(linkTarget).To(Equal("file1"))

					Expect(test_support.SameFile(targetFile1, filepath.Join(target, linkTarget))).To(BeTrue())
				})
			})

			Context("when the source is a directory containing an external symbolic link", func() {
				BeforeEach(func() {
					/*
						   Create a directory structure inside tempDir like this:

							source/
								  link ----> tempDir

					*/

					source = test_support.CreateDir(tempDir, "source")

					tdLink := filepath.Join(source, "link")
					err := os.Symlink(tempDir, tdLink)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should return a suitable error", func() {
					Expect(err).To(MatchError(ContainSubstring("cannot copy symbolic link")))
				})
			})

			Context("when the source is a directory containing an internal relative symbolic link", func() {
				BeforeEach(func() {
					/*
						   Create a directory structure inside tempDir like this:

							source/    <---+
										   | (internal, but via ../source)
								  link ----+

					*/

					source = test_support.CreateDir(tempDir, "source")

					tdLink := filepath.Join(source, "link")
					err := os.Symlink(filepath.Join("..", "source"), tdLink)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should copy the symbolic link", func() {
					Expect(err).NotTo(HaveOccurred())

					targetLink := filepath.Join(target, "link")

					linkTarget, err := os.Readlink(targetLink)
					Expect(err).NotTo(HaveOccurred())
					Expect(linkTarget).To(Equal("."))
				})
			})

			Context("when the source is a directory containing an external relative symbolic link", func() {
				BeforeEach(func() {
					/*
						   Create a directory structure inside tempDir like this:

							a/             <---+
								source/        |
											   | (external via ..)
									  link ----+

					*/

					aDir := test_support.CreateDir(tempDir, "a")
					source = test_support.CreateDir(aDir, "source")

					tdLink := filepath.Join(source, "link")
					err := os.Symlink("..", tdLink)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should return a suitable error", func() {
					Expect(err).To(MatchError(ContainSubstring("cannot copy symbolic link")))
				})
			})

			Context("when source is a symbolic link to a file", func() {
				BeforeEach(func() {
					/*
						   Create a directory structure inside tempDir like this:

							src.file <---+
							source ------+

					*/

					linkTarget := test_support.CreateFile(tempDir, "src.file")
					source = filepath.Join(tempDir, "link")
					err := os.Symlink(linkTarget, source)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should return a suitable error", func() {
					Expect(err).To(MatchError(ContainSubstring("cannot copy symbolic link")))
				})
			})
			Context("when source and target are the same symbolic link", func() {
				BeforeEach(func() {
					/*
					   Create a directory structure inside tempDir like this:

						src.file <---+
						source ------+

					*/
					linkTarget := test_support.CreateFile(tempDir, "src.file")
					source = filepath.Join(tempDir, "link")
					err := os.Symlink(linkTarget, source)
					Expect(err).NotTo(HaveOccurred())
					target = source
				})

				It("should succeed and not corrupt the file contents", func() {
					Expect(err).NotTo(HaveOccurred())

					linkTarget, err := os.Readlink(target)
					Expect(err).NotTo(HaveOccurred())
					Expect(linkTarget).To(Equal(filepath.Join(tempDir, "src.file")))

					checkFile(target, "test contents")
				})
			})
		}
		Context("when the source does not exist", func() {
			BeforeEach(func() {
				source = filepath.Join(tempDir, "src.file")
			})

			It("should return a suitable error", func() {
				err, ok := err.(fileutils.FileError)
				Expect(ok).To(BeTrue())
				Expect(err.ErrorId).To(Equal(fileutils.ErrFileNotFound))
			})
		})
	})

	Context("when the source and target are directories", func() {
		BeforeEach(func() {
			source = filepath.Join(tempDir, "source")
			err := os.Mkdir(source, os.FileMode(0777))
			Expect(err).NotTo(HaveOccurred())

			test_support.CreateFile(source, "file1")
			test_support.CreateFile(source, "file2")

			target = filepath.Join(tempDir, "target")
			err = os.Mkdir(target, os.FileMode(0777))
			Expect(err).NotTo(HaveOccurred())
		})

		It("should copy the source into the target", func() {
			Expect(err).NotTo(HaveOccurred())

			resultantDir := filepath.Join(target, "source")
			checkDirectory(resultantDir)
			checkFile(filepath.Join(resultantDir, "file1"), "test contents")
			checkFile(filepath.Join(resultantDir, "file2"), "test contents")
		})
	})
})

func checkDirectory(path string) {
	Expect(test_support.FileMode(path).IsDir()).To(BeTrue())
}

func checkFile(target string, expectedContents string) {
	f, err := os.Open(target)
	Expect(err).NotTo(HaveOccurred())
	defer f.Close()
	buf := make([]byte, len(expectedContents))
	n, err := f.Read(buf)
	Expect(err).NotTo(HaveOccurred())
	actualContents := string(buf[:n])
	Expect(actualContents).To(Equal(expectedContents))
}
