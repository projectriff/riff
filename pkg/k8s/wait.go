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
	"context"
	"fmt"
	"reflect"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	watchclient "k8s.io/client-go/tools/watch"
)

type object interface {
	runtime.Object
	metav1.Object
}

// WaitUntilReady watches for mutations of the target object until the target is ready.
// Target objects must implement metav1.Object and have a Status field with an IsReady
// method. Types that implement this contract include *.projectriff.io CRDs.
func WaitUntilReady(ctx context.Context, client rest.Interface, resource string, target object) error {
	if readyFunc := readyFunc(target); readyFunc.Kind() != reflect.Func {
		return fmt.Errorf("unsupported target of type %t, must have .Status.IsReady() method", target)
	}
	if client == (*rest.RESTClient)(nil) {
		return nil
	}
	lw := cache.NewListWatchFromClient(client, resource, target.GetNamespace(), fields.Everything())
	_, err := watchclient.UntilWithSync(ctx, lw, target, nil, readyCondition(target))
	return err
}

func readyCondition(target object) watchclient.ConditionFunc {
	return func(event watch.Event) (bool, error) {
		if event.Type == watch.Error {
			return false, fmt.Errorf("error waiting for ready")
		}
		if obj, ok := event.Object.(metav1.Object); !ok || obj.GetUID() != target.GetUID() {
			// event is not for the target resource
			return false, nil
		}
		switch event.Type {
		case watch.Added, watch.Modified:
			if readyFunc := readyFunc(event.Object); readyFunc.Kind() == reflect.Func {
				ready := readyFunc.Call([]reflect.Value{})[0].Bool()
				if ready {
					return true, nil
				}
			}
		case watch.Deleted:
			return true, fmt.Errorf("%s %q deleted", strings.ToLower(target.GetObjectKind().GroupVersionKind().Kind), target.GetName())
		}
		return false, nil
	}
}

func readyFunc(obj interface{}) reflect.Value {
	// use reflection since there is no common interface
	return reflect.ValueOf(obj).Elem().FieldByName("Status").Addr().MethodByName("IsReady")
}
