/*
 * Copyright 2018-Present the original author or authors.
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

package kafka

import (
	"github.com/bsm/sarama-cluster"
	"github.com/Shopify/sarama"
	"log"
)

// NewInspector creates an inspector for a given collection of Kafka brokers.
func NewInspector(brokers []string) (*inspector, error) {
	client, err := cluster.NewClient(brokers, nil)
	if err != nil {
		return nil, err
	}

	return &inspector{
		client: client,
	}, nil
}

type inspector struct {
	client *cluster.Client
}

func (i *inspector) QueueLength(topic string, groupID string) (int64, error) {
	parts, err := i.client.Partitions(topic)
	if err != nil {
		return -1, err
	}

	offsetManager, err := sarama.NewOffsetManagerFromClient(groupID, i.client)
	if err != nil {
		return -1, err
	}

	// Compute the queue length as the sum, across all the partitions, of how far the consumer's committed offset is
	// behind the last produced offset.
	length := int64(0)
	for _, part := range parts {
		end, err := i.client.GetOffset(topic, part, sarama.OffsetNewest)
		if err != nil {
			return -1, err
		}

		pom, err := offsetManager.ManagePartition(topic, part)
		if err != nil {
			return -1, err
		}

		off, _ := pom.NextOffset()
		var current int64
		if off == sarama.OffsetNewest {
			// No offset has been committed yet, so take the current position to be the start of the partition
			current = 0
		} else {
			current = off
		}

		length += end - current
		err = pom.Close()
		if err != nil {
			log.Printf("Failed to close partition offset manager: %v", err)
		}
	}

	return length, nil
}
