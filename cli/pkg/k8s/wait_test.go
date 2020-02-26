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

package k8s_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/projectriff/riff/cli/pkg/k8s"
	rifftesting "github.com/projectriff/riff/cli/pkg/testing"
	"github.com/projectriff/system/pkg/apis"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	cachetesting "k8s.io/client-go/tools/cache/testing"
)

func TestWaitUntilReady(t *testing.T) {
	// using Application, but any type will work
	application := &buildv1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "build.projectriff.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "my-application",
			UID:       "c6acbbab-87dd-11e9-807c-42010a80011d",
		},
		Status: buildv1alpha1.ApplicationStatus{
			Status: apis.Status{
				Conditions: apis.Conditions{
					{
						Type:   apis.ConditionReady,
						Status: corev1.ConditionUnknown,
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		resource *buildv1alpha1.Application
		events   []watch.Event
		err      error
	}{{
		name:     "transitions true",
		resource: application.DeepCopy(),
		events: []watch.Event{
			updateReady(application, corev1.ConditionTrue, ""),
		},
	}, {
		name:     "transitions false",
		resource: application.DeepCopy(),
		events: []watch.Event{
			updateReady(application, corev1.ConditionFalse, "test not ready"),
		},
		err: fmt.Errorf("failed to become ready: %s", "test not ready"),
	}, {
		name:     "ignore other resources",
		resource: application.DeepCopy(),
		events: []watch.Event{
			updateReadyOther(application, corev1.ConditionFalse, "not my app"),
			updateReady(application, corev1.ConditionTrue, ""),
		},
	}, {
		name:     "bail on delete",
		resource: application.DeepCopy(),
		events: []watch.Event{
			updateReady(application, corev1.ConditionUnknown, ""),
			watch.Event{Type: watch.Deleted, Object: application.DeepCopy()},
		},
		err: fmt.Errorf("%s %q deleted", "application", "my-application"),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lw := cachetesting.NewFakeControllerSource()
			defer lw.Shutdown()
			ctx := k8s.WithListerWatcher(context.Background(), lw)

			client := rifftesting.NewClient(application)
			done := make(chan error, 1)
			defer close(done)
			go func() {
				done <- k8s.WaitUntilReady(ctx, client.Build().RESTClient(), "applications", application)
			}()

			time.Sleep(5 * time.Millisecond)
			for _, event := range test.events {
				lw.Change(event, 1)
			}

			err := <-done
			if expected, actual := fmt.Sprintf("%s", test.err), fmt.Sprintf("%s", err); expected != actual {
				t.Errorf("expected error %v, actually %v", expected, actual)
			}
		})
	}
}

func updateReady(application *buildv1alpha1.Application, status corev1.ConditionStatus, message string) watch.Event {
	application = application.DeepCopy()
	application.Status.Conditions[0].Status = status
	application.Status.Conditions[0].Message = message
	return watch.Event{Type: watch.Modified, Object: application}
}

func updateReadyOther(application *buildv1alpha1.Application, status corev1.ConditionStatus, message string) watch.Event {
	application = application.DeepCopy()
	application.UID = "not-a-uid"
	application.Status.Conditions[0].Status = status
	application.Status.Conditions[0].Message = message
	return watch.Event{Type: watch.Modified, Object: application}
}
