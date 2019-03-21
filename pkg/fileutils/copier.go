/*
 * Copyright 2014-2018 The original author or authors
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

package fileutils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	ErrOpeningSourceDir     ErrorId = "cannot open source directory"
	ErrCannotListSourceDir  ErrorId = "cannot list source directory"
	ErrUnexpected           ErrorId = "unexpected error"
	ErrCreatingTargetDir    ErrorId = "cannot create target directory"
	ErrOpeningSourceFile    ErrorId = "cannot open source file"
	ErrOpeningTargetFile    ErrorId = "cannot open target file"
	ErrCopyingFile          ErrorId = "file copy failed"
	ErrReadingSourceSymlink ErrorId = "cannot read symbolic link source"
	ErrWritingTargetSymlink ErrorId = "cannot write target symbolic link"
	ErrExternalSymlink      ErrorId = "external symbolic link"
)

// Copier is a helper interface for copying files and/or directories.
type Copier interface {
	/*
		Copy copies a source file to a destination file. File contents are copied. File mode and permissions
		(as described in http://golang.org/pkg/os/#FileMode) are copied.

		Directories are copied, along with their contents.

		Copying a file or directory to itself succeeds but does not modify the filesystem.

		Symbolic links are not followed and are copied provided they refer to a file or directory being copied
		(otherwise a non-nil error is returned). The only exception is copying a symbolic link to itself, which
		always succeeds.
	*/
	Copy(destPath string, srcPath string) error
}

type copier struct {
	log     io.Writer
	checker Checker
}

func NewCopier(log io.Writer, checker Checker) *copier {
	return &copier{
		log:     log,
		checker: checker,
	}
}

func (f *copier) Copy(destPath string, srcPath string) error {
	fmt.Fprintf(f.log, "copying %q to %q\n", srcPath, destPath)
	return f.doCopy(destPath, srcPath, srcPath)
}

func (f *copier) doCopy(destPath string, srcPath string, topSrcPath string) error {
	if f.sameFile(srcPath, destPath) {
		return nil
	}
	srcMode, err := f.checker.Filemode(srcPath)
	if err != nil {
		return err
	}

	if srcMode&os.ModeSymlink == os.ModeSymlink {
		return f.copySymlink(destPath, srcPath, topSrcPath)
	} else if srcMode.IsDir() {
		return f.copyDir(destPath, srcPath, topSrcPath)
	} else {
		return f.copyFile(destPath, srcPath)
	}
}

func (f *copier) copyDir(destination string, source string, topSource string) error {
	finalDestination, err := f.finalDestinationDir(destination, source)
	if err != nil {
		return err
	}

	names, err := getNames(source)
	if err != nil {
		return err
	}

	for _, name := range names {
		fmt.Fprintf(f.log, "copying %q from %q to %q\n", name, source, finalDestination)
		err = f.doCopy(filepath.Join(finalDestination, name), filepath.Join(source, name), topSource)
		if err != nil {
			return err
		}
	}

	return nil
}

func getNames(dirPath string) (names []string, err error) {
	src, err := os.Open(dirPath)
	if err != nil {
		return names, newFileError(ErrOpeningSourceDir, err)
	}
	defer func() {
		if err := src.Close(); err != nil {
			fmt.Printf("Warning: cannot close %v", src)
		}
	}()

	names, err = src.Readdirnames(-1)
	if err != nil {
		return names, newFileError(ErrCannotListSourceDir, err)
	}

	return names, nil
}

/*
	Determine the final destination directory and return an opened file referring to it.
*/
func (f *copier) finalDestinationDir(destination string, source string) (finalDestination string, err error) {
	sourceMode, err := f.checker.Filemode(source)
	if err != nil {
		return finalDestination, err
	}
	if _, err := os.Stat(destination); err != nil {
		if !os.IsNotExist(err) {
			return finalDestination, newFileError(ErrUnexpected, err)
		}
		finalDestination = destination
	} else {
		finalDestination = filepath.Join(destination, filepath.Base(source))
	}
	if err := os.MkdirAll(finalDestination, sourceMode); err != nil {
		finalDestination = ""
		return finalDestination, newFileError(ErrCreatingTargetDir, err)
	}
	return finalDestination, nil
}

func (f *copier) copyFile(destination string, source string) error {
	sourceFile, err := os.OpenFile(source, os.O_RDONLY, 0666)
	if err != nil {
		return newFileError(ErrOpeningSourceFile, err)
	}
	defer sourceFile.Close()

	mode, err := f.checker.Filemode(source)
	if err != nil {
		return err
	}

	destinationFile, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY, mode)
	if err != nil {
		return newFileError(ErrOpeningTargetFile, err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return newFileError(ErrCopyingFile, err)
	}
	return nil
}

func (f *copier) copySymlink(destLinkPath string, srcLinkPath string, topSrcPath string) error {
	linkTarget, err := os.Readlink(srcLinkPath)
	if err != nil {
		return newFileError(ErrReadingSourceSymlink, err)
	}

	// Ensure linkTarget is absolute
	if strings.HasPrefix(linkTarget, "..") {
		linkTarget = filepath.Join(filepath.Dir(srcLinkPath), linkTarget)
	}

	// check link does not point outside any directory being copied
	topRelativePath, err := filepath.Rel(topSrcPath, linkTarget)
	if err != nil {
		return newFileError(ErrUnexpected, err)
	}
	if strings.HasPrefix(topRelativePath, "..") {
		return newFileErrorf(ErrExternalSymlink,
			"cannot copy symbolic link %q with target %q which points outside the file or directory being copied %q",
			srcLinkPath, linkTarget, topSrcPath)
	}

	linkParent := filepath.Dir(srcLinkPath)
	relativePath, err := filepath.Rel(linkParent, linkTarget)
	if err != nil {
		return newFileError(ErrUnexpected, err)
	}
	fmt.Fprintf(f.log, "symbolic link %q has target %q which has path %q relative to %q (directory containing link)\n",
		srcLinkPath, linkTarget, relativePath, linkParent)
	err = os.Symlink(relativePath, destLinkPath)
	if err != nil {
		return newFileError(ErrWritingTargetSymlink, err)
	}
	fmt.Fprintf(f.log, "symbolically linked %q to %q\n", destLinkPath, relativePath)
	return nil
}

func (f *copier) sameFile(srcPath string, destPath string) bool {
	srcFi, err := os.Stat(srcPath)
	if err == nil {
		destFi, err := os.Stat(destPath)
		if err == nil {
			return os.SameFile(srcFi, destFi)
		}
	}
	return false
}
