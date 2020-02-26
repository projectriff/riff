/*
Copyright 2018 The Knative Authors

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

// modified from https://github.com/knative/pkg/tree/master/tracker

package tracker

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// Tracker defines the interface through which an object can register
// that it is tracking another object by reference.
type Tracker interface {
	// Track tells us that "obj" is tracking changes to the
	// referenced object.
	Track(ref Key, obj types.NamespacedName)

	// Lookup returns actively tracked objects for the reference.
	Lookup(ref Key) []types.NamespacedName
}

func NewKey(gvk schema.GroupVersionKind, namespacedName types.NamespacedName) Key {
	return Key{
		GroupKind:      gvk.GroupKind(),
		NamespacedName: namespacedName,
	}
}

type Key struct {
	GroupKind      schema.GroupKind
	NamespacedName types.NamespacedName
}

func (k *Key) String() string {
	return fmt.Sprintf("%s/%s", k.GroupKind, k.NamespacedName)
}

// New returns an implementation of Tracker that lets a Reconciler
// register a particular resource as watching a resource for
// a particular lease duration.  This watch must be refreshed
// periodically (e.g. by a controller resync) or it will expire.
func New(lease time.Duration, log logr.Logger) Tracker {
	return &impl{
		log:           log,
		leaseDuration: lease,
	}
}

type impl struct {
	log logr.Logger
	m   sync.Mutex

	// mapping maps from an object reference to the set of
	// keys for objects watching it.
	mapping map[string]set

	// The amount of time that an object may watch another
	// before having to renew the lease.
	leaseDuration time.Duration
}

// Check that impl implements Tracker.
var _ Tracker = (*impl)(nil)

// set is a map from keys to expirations
type set map[types.NamespacedName]time.Time

// Track implements Tracker.
func (i *impl) Track(ref Key, obj types.NamespacedName) {
	i.m.Lock()
	defer i.m.Unlock()
	if i.mapping == nil {
		i.mapping = make(map[string]set)
	}

	l, ok := i.mapping[ref.String()]
	if !ok {
		l = set{}
	}
	// Overwrite the key with a new expiration.
	l[obj] = time.Now().Add(i.leaseDuration)

	i.mapping[ref.String()] = l

	i.log.Info("tracking resource", "ref", ref.String(), "obj", obj.String(), "ttl", l[obj].UTC().Format(time.RFC3339))
}

func isExpired(expiry time.Time) bool {
	return time.Now().After(expiry)
}

// Lookup implements Tracker.
func (i *impl) Lookup(ref Key) []types.NamespacedName {
	items := []types.NamespacedName{}

	// TODO(mattmoor): Consider locking the mapping (global) for a
	// smaller scope and leveraging a per-set lock to guard its access.
	i.m.Lock()
	defer i.m.Unlock()
	s, ok := i.mapping[ref.String()]
	if !ok {
		i.log.V(2).Info("no tracked items found", "ref", ref.String())
		return items
	}

	for key, expiry := range s {
		// If the expiration has lapsed, then delete the key.
		if isExpired(expiry) {
			delete(s, key)
			continue
		}
		items = append(items, key)
	}

	if len(s) == 0 {
		delete(i.mapping, ref.String())
	}

	i.log.V(1).Info("found tracked items", "ref", ref.String(), "items", items)

	return items
}
