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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// WaitUntilReady watches for mutations of the target object until the target is ready.
// Target objects must implement metav1.Object and have a Status field with an IsReady
// method. Types that implement this contract include *.projectriff.io CRDs.
func WaitUntilReady(target metav1.Object, watcher watch.Interface) error {
	if readyFunc := readyFunc(target); readyFunc.Kind() != reflect.Func {
		return fmt.Errorf("unsupported target of type %t, must have .Status.IsReady() method", target)
	}
	for {
		select {
		case ev := <-watcher.ResultChan():
			if ev.Type == watch.Error {
				return fmt.Errorf("error waiting for ready")
			}
			if obj, ok := ev.Object.(metav1.Object); !ok || obj.GetUID() != target.GetUID() {
				// event is not for the target resource
				continue
			}
			switch ev.Type {
			case watch.Added, watch.Modified:
				if readyFunc := readyFunc(ev.Object); readyFunc.Kind() == reflect.Func {
					if readyFunc.Call([]reflect.Value{})[0].Bool() {
						return nil
					}
				}
			case watch.Deleted:
				return fmt.Errorf("%s deleted", strings.ToLower(ev.Object.GetObjectKind().GroupVersionKind().Kind))
			}
		}
	}
}

func readyFunc(obj interface{}) reflect.Value {
	// use reflection since there is no common interface
	return reflect.ValueOf(obj).Elem().FieldByName("Status").Addr().MethodByName("IsReady")
}
