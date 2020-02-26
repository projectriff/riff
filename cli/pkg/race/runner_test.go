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

package race_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/projectriff/riff/cli/pkg/race"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		tasks   []race.Task
		err     error
	}{{
		name:    "empty",
		timeout: time.Millisecond,
	}, {
		name:    "return immediately",
		timeout: time.Minute,
		tasks: []race.Task{
			func(ctx context.Context) error {
				return nil
			},
		},
	}, {
		name:    "error immediately",
		timeout: time.Minute,
		tasks: []race.Task{
			func(ctx context.Context) error {
				return fmt.Errorf("test error")
			},
		},
		err: fmt.Errorf("test error"),
	}, {
		name:    "block until canceled",
		timeout: time.Millisecond,
		tasks: []race.Task{
			func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			},
		},
		err: context.DeadlineExceeded,
	}, {
		name:    "take first task to return",
		timeout: time.Minute,
		tasks: []race.Task{
			func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			},
			func(ctx context.Context) error {
				return nil
			},
		},
	}, {
		name:    "take first task to error",
		timeout: time.Minute,
		tasks: []race.Task{
			func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			},
			func(ctx context.Context) error {
				return fmt.Errorf("test error")
			},
		},
		err: fmt.Errorf("test error"),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			err := race.Run(ctx, test.timeout, test.tasks...)
			if expected, actual := fmt.Sprintf("%s", test.err), fmt.Sprintf("%s", err); expected != actual {
				t.Errorf("Expected error to be %q, actually %q", expected, actual)
			}
		})
	}
}
