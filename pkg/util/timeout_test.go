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

package util_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/projectriff/riff/pkg/util"
)

func TestRaceUntil(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		tasks   []util.RaceTask
		err     error
	}{{
		name:    "empty",
		timeout: time.Millisecond,
		err:     context.DeadlineExceeded,
	}, {
		name:    "return immediately",
		timeout: time.Minute,
		tasks: []util.RaceTask{
			func(ctx context.Context) error {
				return nil
			},
		},
	}, {
		name:    "error immediately",
		timeout: time.Minute,
		tasks: []util.RaceTask{
			func(ctx context.Context) error {
				return fmt.Errorf("test error")
			},
		},
		err: fmt.Errorf("test error"),
	}, {
		name:    "block until canceled",
		timeout: time.Millisecond,
		tasks: []util.RaceTask{
			func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			},
		},
		err: context.DeadlineExceeded,
	}, {
		name:    "take first function to return",
		timeout: time.Minute,
		tasks: []util.RaceTask{
			func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			},
			func(ctx context.Context) error {
				return nil
			},
		},
	}, {
		name:    "take first function to error",
		timeout: time.Minute,
		tasks: []util.RaceTask{
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
			ctx := context.TODO()
			err := util.RaceUntil(ctx, test.timeout, test.tasks...)
			if expected, actual := fmt.Sprintf("%s", test.err), fmt.Sprintf("%s", err); expected != actual {
				t.Errorf("Expected error to be %q, actually %q", expected, actual)
			}
		})
	}
}
