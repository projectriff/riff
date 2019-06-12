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

package race

import (
	"context"
	"time"
)

// Task is a function which performs some kind of activity and returns an error value. It also observes the given
// context and, once the context is done, terminates the action and returns (the returned error is ignored in this case).
type Task func(ctx context.Context) error

// Run runs each task in its own go routine and returns the result of the first task to complete (or error). A timeout
// is specified to limit the whole process. Each task should observe the context and return once the context is done.
func Run(ctx context.Context, timeout time.Duration, tasks ...Task) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	errChan := make(chan error, len(tasks)+1)
	defer close(errChan)

	for _, task := range tasks {
		go func(task Task) {
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
