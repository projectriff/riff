/*
 * Copyright 2017 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controller

import (
	"bytes"
	"log"
	"sort"
	"strconv"

	"fmt"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
)

// LagTracker is used to compute how many unprocessed messages each function needs to take care of.
type LagTracker interface {
	// Register a given function for monitoring.
	BeginTracking(Subscription) error

	// Unregister a function for monitoring.
	StopTracking(Subscription) error

	// Compute the current lags for all tracked subscriptions
	Compute() map[Subscription]PartitionedOffsets
}

// Subscription describes a tracked tuple of topic and consumer group.
type Subscription struct {
	Topic string
	Group string
}

// Offsets gives per-partition information about current and end offsets.
type Offsets struct {
	Current         int64
	End             int64
	PreviousCurrent int64
}

type PartitionedOffsets map[int32]Offsets

// Lag returns how many messages are available that haven't been handled yet.
func (o Offsets) Lag() int64 {
	return o.End - o.Current
}

// Activity returns how many messages have been handled since a previous query of the tracker.
func (o Offsets) Activity() int64 {
	if o.PreviousCurrent != -1 {
		return o.Current - o.PreviousCurrent
	} else {
		return 0
	}
}

type tracker struct {
	subscriptions map[Subscription]PartitionedOffsets
	client        *cluster.Client
}

func (t *tracker) BeginTracking(s Subscription) error {
	t.subscriptions[s] = make(PartitionedOffsets)
	parts, err := t.client.Partitions(s.Topic)
	if err != nil {
		return err
	}

	for _, part := range parts {
		t.subscriptions[s][part] = Offsets{End: -1, Current: -1, PreviousCurrent: -1}
	}
	return nil
}

func (t *tracker) StopTracking(s Subscription) error {
	delete(t.subscriptions, s)
	return nil
}

func (t *tracker) Compute() map[Subscription]PartitionedOffsets {
	for s, _ := range t.subscriptions {
		t.subscriptions[s] = t.offsetsForSubscription(s)
	}
	return t.subscriptions
}

func (t *tracker) offsetsForSubscription(s Subscription) PartitionedOffsets {

	parts, err := t.client.Partitions(s.Topic)
	if err != nil {
		log.Printf("Error reading partitions for topic %v: %v", s.Topic, err)
	}
	oldOffsets := t.subscriptions[s]
	os := make(PartitionedOffsets, len(oldOffsets))

	offsetManager, err := sarama.NewOffsetManagerFromClient(s.Group, t.client)
	if err != nil {
		log.Printf("Got error %v", err)
	}
	defer offsetManager.Close()

	for _, part := range parts {
		end, err := t.client.GetOffset(s.Topic, part, sarama.OffsetNewest)
		if err != nil {
			log.Printf("Got error %v", err)
		}

		pom, err := offsetManager.ManagePartition(s.Topic, part)
		if err != nil {
			log.Printf("Got error %v", err)
		}
		off, _ := pom.NextOffset()
		var current int64
		if off == sarama.OffsetNewest {
			current = 0
		} else {
			current = off
		}
		os[part] = Offsets{End: end, Current: current, PreviousCurrent: oldOffsets[part].Current}
		err = pom.Close()
	}
	return os
}

func NewLagTracker(brokers []string) LagTracker {
	c, _ := cluster.NewClient(brokers, nil)
	return &tracker{
		subscriptions: make(map[Subscription]PartitionedOffsets),
		client:        c,
	}
}

func (o Offsets) String() string {
	if o.PreviousCurrent != -1 {
		return fmt.Sprintf("Offsets[lag=%v (=%v-%v), Activity:%v (=%v-%v)]", o.Lag(), o.End, o.Current, o.Activity(), o.Current, o.PreviousCurrent)
	} else {
		return fmt.Sprintf("Offsets[lag=%v (=%v-%v), Activity:%v (unknown)]", o.Lag(), o.End, o.Current, o.Activity())
	}
}

func (po PartitionedOffsets) String() string {
	keys := make([]int, len(po))
	i := 0
	for k, _ := range po {
		keys[i] = int(k)
		i++
	}
	sort.Ints(keys)
	var buffer bytes.Buffer
	for _, k := range keys {
		buffer.WriteString(strconv.Itoa(k))
		buffer.WriteString(":")
		buffer.WriteString(po[int32(k)].String())
		buffer.WriteString(" ")
	}
	return buffer.String()
}
