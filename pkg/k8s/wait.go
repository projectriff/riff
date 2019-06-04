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

	"github.com/projectriff/system/pkg/apis"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	watchclient "k8s.io/client-go/tools/watch"
)

var ErrWaitTimeout = wait.ErrWaitTimeout

// WaitUntilReady watches for mutations of the target object until the target is ready.
func WaitUntilReady(ctx context.Context, client rest.Interface, resource string, target apis.Object) error {
	if client == (*rest.RESTClient)(nil) {
		return nil
	}
	lw := cache.NewListWatchFromClient(client, resource, target.GetNamespace(), fields.Everything())
	_, err := watchclient.UntilWithSync(ctx, lw, target, nil, readyCondition(target))
	return err
}

func readyCondition(target apis.Object) watchclient.ConditionFunc {
	return func(event watch.Event) (bool, error) {
		if event.Type == watch.Error {
			return false, fmt.Errorf("error waiting for ready")
		}
		obj, ok := event.Object.(apis.Object)
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
			return true, fmt.Errorf("%s %q deleted", strings.ToLower(target.GetObjectKind().GroupVersionKind().Kind), target.GetName())
		}
		return false, nil
	}
}
