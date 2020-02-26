/*
Copyright 2020 the original author or authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testing

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	"github.com/projectriff/riff/system/pkg/apis"
)

type Event struct {
	metav1.TypeMeta
	types.NamespacedName
	Type    string
	Reason  string
	Message string
}

func NewEvent(factory Factory, scheme *runtime.Scheme, eventtype, reason, messageFormat string, a ...interface{}) Event {
	obj := factory.CreateObject()
	gvks, _, _ := scheme.ObjectKinds(obj)
	apiVersion, kind := gvks[0].ToAPIVersionAndKind()

	return Event{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiVersion,
			Kind:       kind,
		},
		NamespacedName: types.NamespacedName{
			Namespace: obj.GetNamespace(),
			Name:      obj.GetName(),
		},
		Type:    eventtype,
		Reason:  reason,
		Message: fmt.Sprintf(messageFormat, a...),
	}
}

type eventRecorder struct {
	events []Event
	scheme *runtime.Scheme
}

var (
	_ record.EventRecorder = (*eventRecorder)(nil)
)

func (r *eventRecorder) Event(object runtime.Object, eventtype, reason, message string) {
	o := object.(apis.Object)
	gvks, _, _ := r.scheme.ObjectKinds(o)
	apiVersion, kind := gvks[0].ToAPIVersionAndKind()
	r.events = append(r.events, Event{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiVersion,
			Kind:       kind,
		},
		NamespacedName: types.NamespacedName{
			Namespace: o.GetNamespace(),
			Name:      o.GetName(),
		},
		Type:    eventtype,
		Reason:  reason,
		Message: message,
	})
}

func (r *eventRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
	r.Event(object, eventtype, reason, fmt.Sprintf(messageFmt, args...))
}

func (r *eventRecorder) PastEventf(object runtime.Object, timestamp metav1.Time, eventtype, reason, messageFmt string, args ...interface{}) {
	panic("not implemented")
}

func (r *eventRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
	r.Eventf(object, eventtype, reason, messageFmt, args...)
}
