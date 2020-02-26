/*
Copyright 2019 the original author or authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testing

import (
	"fmt"
	"testing"

	controllerruntime "sigs.k8s.io/controller-runtime"
)

func AssertErrorEqual(expected error) VerifyFunc {
	return func(t *testing.T, result controllerruntime.Result, err error) {
		if err != expected {
			t.Errorf("Unexpected error: expected %v, actual %v", expected, err)
		}
	}
}

func AssertErrorMessagef(message string, a ...interface{}) VerifyFunc {
	return func(t *testing.T, result controllerruntime.Result, err error) {
		expected := fmt.Sprintf(message, a...)
		actual := ""
		if err != nil {
			actual = err.Error()
		}
		if actual != expected {
			t.Errorf("Unexpected error message: expected %v, actual %v", expected, actual)
		}
	}
}
