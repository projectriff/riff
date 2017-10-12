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

package io.sk8s.event.dispatcher;

import org.apache.kafka.clients.consumer.Consumer;
import org.apache.kafka.clients.consumer.ConsumerConfig;
import org.apache.kafka.clients.consumer.KafkaConsumer;
import org.apache.kafka.clients.consumer.OffsetAndMetadata;
import org.apache.kafka.common.PartitionInfo;
import org.apache.kafka.common.TopicPartition;
import org.apache.kafka.common.serialization.StringDeserializer;

import java.io.Closeable;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.Properties;
import java.util.Set;
import java.util.UUID;
import java.util.function.Function;
import java.util.stream.Collectors;

/**
 * Tracks current and end offset (and hence lag) for a set of topic/consumerGroup pairs.
 *
 * @author Eric Bottard
 */
public class KafkaConsumerMonitor implements Closeable {

	private final String host;

	private final Map<String, Consumer<?, ?>> consumersByGroup = new HashMap<>();

	private final String monitoringConsumerGroupID = "monitor_" + UUID.randomUUID().toString();

	private volatile Consumer<?,?> logEndOffsetTrackingConsumer;

	private final Set<Subscription> tracked = new HashSet<>();

	public KafkaConsumerMonitor(String host) {
		this.host = host;
	}

	@Override
	public synchronized void close() {
		for (Subscription subscription : tracked) {
			stopTracking(subscription.group, subscription.topic);
		}
		if (logEndOffsetTrackingConsumer != null) {
			logEndOffsetTrackingConsumer.close();
		}
	}

	public synchronized void beginTracking(String group, String topic) {
		if (logEndOffsetTrackingConsumer == null) {
			logEndOffsetTrackingConsumer = createNewConsumer(monitoringConsumerGroupID);
		}
		consumersByGroup.computeIfAbsent(group, k -> createNewConsumer(group));
		tracked.add(new Subscription(topic, group));
		reassignPartitions();
	}

	public synchronized void stopTracking(String group, String topic) {
		tracked.remove(new Subscription(topic, group));
		if (tracked.stream().noneMatch(tg -> tg.group.equals(group))) {
			consumersByGroup.remove(group).close();
		}
		reassignPartitions();
	}

	private void reassignPartitions() {
		logEndOffsetTrackingConsumer.assign(
				tracked.stream()
						.flatMap(tg -> logEndOffsetTrackingConsumer.partitionsFor(tg.topic).stream()
								.map(pi -> new TopicPartition(tg.topic, pi.partition()))
						).collect(Collectors.toSet()));
	}

	public Map<Subscription, List<Offsets>> compute() {
		logEndOffsetTrackingConsumer.seekToEnd(Collections.emptySet()); // Means "all currently assigned"

		return tracked.stream()
				.collect(Collectors.toMap(
						Function.identity(),
						tg -> {
							List<PartitionInfo> partitionInfos = logEndOffsetTrackingConsumer.partitionsFor(tg.topic);
							return partitionInfos.stream()
									.map(pi -> new TopicPartition(tg.topic, pi.partition()))
									.map(
											tp -> {
												OffsetAndMetadata committed = consumersByGroup.get(tg.group).committed(tp);
												if (committed == null) {
													committed = new OffsetAndMetadata(0L);
												}
												long position = logEndOffsetTrackingConsumer.position(tp);
												return new Offsets(tp.partition(), position,
														committed.offset());
											})
									.collect(Collectors.toList());
						}));
	}

	static class Subscription {
		final String topic;
	
		final String group;

		Subscription(String topic, String group) {
			this.topic = topic;
			this.group = group;
		}

		@Override
		public boolean equals(Object o) {
			if (this == o) return true;
			if (o == null || getClass() != o.getClass()) return false;
			Subscription that = (Subscription) o;
			return Objects.equals(topic, that.topic) &&
					Objects.equals(group, that.group);
		}

		@Override
		public int hashCode() {
			return Objects.hash(topic, group);
		}

		@Override
		public String toString() {
			return "Subscription{" +
					"topic='" + topic + '\'' +
					", group='" + group + '\'' +
					'}';
		}
	}

	static class Offsets {
		final int partition;
		final long endOffset;
		final long currentOffset;

		Offsets(int partition, long endOffset, long currentOffset) {
			this.partition = partition;
			this.endOffset = endOffset;
			this.currentOffset = currentOffset;
		}

		long getLag() {
			return endOffset - currentOffset;
		}

		@Override
		public String toString() {
			return "Offsets{" +
					"partition=" + partition +
					", endOffset=" + endOffset +
					", currentOffset=" + currentOffset +
					'}';
		}
	}

	private KafkaConsumer<?, ?> createNewConsumer(String groupId) {
		Properties properties = new Properties();
		properties.put(ConsumerConfig.BOOTSTRAP_SERVERS_CONFIG, host);
		properties.put(ConsumerConfig.GROUP_ID_CONFIG, groupId);
		properties.put(ConsumerConfig.ENABLE_AUTO_COMMIT_CONFIG, "false");
		properties.put(ConsumerConfig.KEY_DESERIALIZER_CLASS_CONFIG, StringDeserializer.class);
		properties.put(ConsumerConfig.VALUE_DESERIALIZER_CLASS_CONFIG, StringDeserializer.class);
		return new KafkaConsumer<>(properties);
	}

}
