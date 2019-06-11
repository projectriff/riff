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

package util

import (
	"context"
	"time"
)

type RaceTask func(ctx context.Context) error

// RaceUntil runs each task in go routine and returns the result of the first task to
// complete (or error). A timeout is specified to limit the whole process. Each task
// should observe the context to exit once the context is done.
func RaceUntil(ctx context.Context, timeout time.Duration, tasks ...RaceTask) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	errChan := make(chan error, len(tasks)+1)
	defer close(errChan)

	for _, task := range tasks {
		go func(task RaceTask) {
			defer cancel()
			err := task(ctx)
			if ctx.Err() == nil {
				errChan <- err
			}
		}(task)
	}

	<-ctx.Done()
	errChan <- ctx.Err()

	return <-errChan
}
