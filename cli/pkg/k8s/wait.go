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
	"strings"

	"github.com/projectriff/riff/system/pkg/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	watchclient "k8s.io/client-go/tools/watch"
)

var ErrWaitTimeout = wait.ErrWaitTimeout

type object interface {
	apis.Resource
	metav1.Object
	runtime.Object
}

// WaitUntilReady watches for mutations of the target object until the target is ready.
func WaitUntilReady(ctx context.Context, client rest.Interface, resource string, target object) error {
	lw := GetListerWatcher(ctx, client, resource, target)
	_, err := watchclient.UntilWithSync(ctx, lw, target, nil, readyCondition(target))
	return err
}

func readyCondition(target object) watchclient.ConditionFunc {
	return func(event watch.Event) (bool, error) {
		if event.Type == watch.Error {
			return false, fmt.Errorf("error waiting for ready")
		}
		obj, ok := event.Object.(object)
		if !ok || obj.GetUID() != target.GetUID() {
			// event is not for the target resource
			return false, nil
		}
		switch event.Type {
		case watch.Added, watch.Modified:
			status := obj.GetStatus()
			if status.IsReady() {
				return true, nil
			}
			readyCond := status.GetCondition(status.GetReadyConditionType())
			if readyCond != nil && readyCond.IsFalse() {
				return false, fmt.Errorf("failed to become ready: %s", readyCond.Message)
			}
			return false, nil
		case watch.Deleted:
			return false, fmt.Errorf("%s %q deleted", strings.ToLower(target.GetObjectKind().GroupVersionKind().Kind), target.GetName())
		}
		return false, nil
	}
}

type lwKey struct{}

func WithListerWatcher(ctx context.Context, lw cache.ListerWatcher) context.Context {
	return context.WithValue(ctx, lwKey{}, lw)
}

func GetListerWatcher(ctx context.Context, client rest.Interface, resource string, target object) cache.ListerWatcher {
	if lw, ok := ctx.Value(lwKey{}).(cache.ListerWatcher); ok {
		return lw
	}
	return cache.NewListWatchFromClient(client, resource, target.GetNamespace(), fields.Everything())
}
