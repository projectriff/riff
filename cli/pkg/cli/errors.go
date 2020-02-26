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

package cli

var SilentError = &silentError{}

type silentError struct {
	err error
}

func (e *silentError) Error() string {
	return e.err.Error()
}

func (e *silentError) Unwrap() error {
	return e.err
}

func (e *silentError) Is(err error) bool {
	_, ok := err.(*silentError)
	return ok
}

func SilenceError(err error) error {
	return &silentError{err: err}
}
