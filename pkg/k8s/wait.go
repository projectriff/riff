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
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	requestv1alpha1 "github.com/projectriff/system/pkg/apis/request/v1alpha1"
	streamv1alpha1 "github.com/projectriff/system/pkg/apis/stream/v1alpha1"
	"k8s.io/apimachinery/pkg/watch"
)

func WaitUntilReady(watcher watch.Interface) {
	func() {
		for {
			select {
			case ev := <-watcher.ResultChan():
				switch ev.Type {
				case watch.Added, watch.Modified:
					// TODO create an interface for this
					if application, ok := ev.Object.(*buildv1alpha1.Application); ok && application.Status.IsReady() {
						return
					}
					if function, ok := ev.Object.(*buildv1alpha1.Function); ok && function.Status.IsReady() {
						return
					}
					if handler, ok := ev.Object.(*requestv1alpha1.Handler); ok && handler.Status.IsReady() {
						return
					}
					if stream, ok := ev.Object.(*streamv1alpha1.Stream); ok && stream.Status.IsReady() {
						return
					}
					if processor, ok := ev.Object.(*streamv1alpha1.Processor); ok && processor.Status.IsReady() {
						return
					}
				case watch.Deleted, watch.Error:
					return
				}
			}
		}
	}()
}
