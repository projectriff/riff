/*
 * Copyright 2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cli_test

import (
	"fmt"
	"testing"

	"github.com/projectriff/riff/pkg/cli"
)

func TestSilenceError(t *testing.T) {
	err := fmt.Errorf("test error")
	silentErr := cli.SilenceError(err)

	if cli.IsSilent(err) {
		t.Errorf("expected error to not be silent, got %#v", err)
	}
	if !cli.IsSilent(silentErr) {
		t.Errorf("expected error to be silent, got %#v", err)
	}
	if expected, actual := err.Error(), silentErr.Error(); expected != actual {
		t.Errorf("errors expected to match, expected %q, actually %q", expected, actual)
	}
}
