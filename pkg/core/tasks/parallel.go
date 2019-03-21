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

package tasks

import "sync"

type CorrelatedResult struct {
	Input string
	Error error
}

func ApplyInParallel(names []string, action func(string) error) []CorrelatedResult {
	count := len(names)
	results := make([]CorrelatedResult, count)
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(count)
	for i, name := range names {
		go func(i int, name string) {
			results[i] = CorrelatedResult{
				Input: name,
				Error: action(name),
			}
			waitGroup.Done()
		}(i, name)
	}
	waitGroup.Wait()
	return results
}
