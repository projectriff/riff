/*
Copyright 2019 the original author or authors.

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
	"time"

	"github.com/go-logr/logr/testing"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/projectriff/riff/system/pkg/tracker"
)

// TrackRequest records that one object is tracking another object.
type TrackRequest struct {
	// Tracker is the object doing the tracking
	Tracker types.NamespacedName
	// Tracked is the object being tracked
	Tracked tracker.Key
}

type trackBy func(trackingObjNamespace, trackingObjName string) TrackRequest

func (t trackBy) By(trackingObjNamespace, trackingObjName string) TrackRequest {
	return t(trackingObjNamespace, trackingObjName)
}

func CreateTrackRequest(trackedObjGroup, trackedObjKind, trackedObjNamespace, trackedObjName string) trackBy {
	return func(trackingObjNamespace, trackingObjName string) TrackRequest {
		return TrackRequest{
			Tracked: tracker.Key{GroupKind: schema.GroupKind{Group: trackedObjGroup, Kind: trackedObjKind}, NamespacedName: types.NamespacedName{Namespace: trackedObjNamespace, Name: trackedObjName}},
			Tracker: types.NamespacedName{Namespace: trackingObjNamespace, Name: trackingObjName},
		}
	}
}

func NewTrackRequest(t, b Factory, scheme *runtime.Scheme) TrackRequest {
	tracked, by := t.CreateObject(), b.CreateObject()
	gvks, _, err := scheme.ObjectKinds(tracked)
	if err != nil {
		panic(err)
	}
	return TrackRequest{
		Tracked: tracker.Key{GroupKind: schema.GroupKind{Group: gvks[0].Group, Kind: gvks[0].Kind}, NamespacedName: types.NamespacedName{Namespace: tracked.GetNamespace(), Name: tracked.GetName()}},
		Tracker: types.NamespacedName{Namespace: by.GetNamespace(), Name: by.GetName()},
	}
}

const maxDuration = time.Duration(1<<63 - 1)

func createTracker() *mockTracker {
	return &mockTracker{Tracker: tracker.New(maxDuration, testing.NullLogger{}), reqs: []TrackRequest{}}
}

type mockTracker struct {
	tracker.Tracker
	reqs []TrackRequest
}

var _ tracker.Tracker = &mockTracker{}

func (t *mockTracker) Track(ref tracker.Key, obj types.NamespacedName) {
	t.Tracker.Track(ref, obj)
	t.reqs = append(t.reqs, TrackRequest{Tracked: ref, Tracker: obj})
}

func (t *mockTracker) getTrackRequests() []TrackRequest {
	result := []TrackRequest{}
	for _, req := range t.reqs {
		result = append(result, req)
	}
	return result
}
