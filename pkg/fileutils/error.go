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

package fileutils

import "fmt"

type ErrorId string

type FileError struct {
	ErrorId ErrorId
	Cause   error
}

func newFileError(id ErrorId, cause error) FileError {
	return FileError{
		ErrorId: id,
		Cause:   cause,
	}
}

func newFileErrorf(tag ErrorId, format string, insert ...interface{}) error {
	return FileError{
		ErrorId: tag,
		Cause:   fmt.Errorf(format, insert...),
	}
}

func (fe FileError) Error() string {
	return fmt.Sprintf("fileutils error: %s: %v", fe.ErrorId, fe.Cause)
}
