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

package k8s

import (
	"fmt"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/watch"
)

func WaitUntilReady(watcher watch.Interface) error {
	for {
		select {
		case ev := <-watcher.ResultChan():
			switch ev.Type {
			case watch.Added, watch.Modified:
				// use reflection since there is no common interface
				if readyFunc := reflect.ValueOf(ev.Object).Elem().FieldByName("Status").Addr().MethodByName("IsReady"); readyFunc.Kind() == reflect.Func {
					if readyFunc.Call([]reflect.Value{})[0].Bool() {
						return nil
					}
				}
			case watch.Deleted:
				return fmt.Errorf("%s deleted", strings.ToLower(ev.Object.GetObjectKind().GroupVersionKind().Kind))
			case watch.Error:
				return fmt.Errorf("error waiting for ready")
			}
		}
	}
}
