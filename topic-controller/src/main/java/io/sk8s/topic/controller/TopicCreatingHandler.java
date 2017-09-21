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

package io.sk8s.topic.controller;

import io.sk8s.core.resource.ResourceAddedEvent;
import io.sk8s.core.resource.ResourceDeletedEvent;
import io.sk8s.kubernetes.api.model.Topic;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

import org.springframework.cloud.stream.binder.ExtendedConsumerProperties;
import org.springframework.cloud.stream.binder.ExtendedProducerProperties;
import org.springframework.cloud.stream.binder.kafka.properties.KafkaConsumerProperties;
import org.springframework.cloud.stream.binder.kafka.properties.KafkaProducerProperties;
import org.springframework.cloud.stream.provisioning.ProvisioningProvider;
import org.springframework.context.event.EventListener;

/**
 * @author Mark Fisher
 * @author Eric Bottard
 */
public class TopicCreatingHandler {

	private static Log logger = LogFactory.getLog(TopicCreatingHandler.class);

	private final ProvisioningProvider<ExtendedConsumerProperties<KafkaConsumerProperties>, ExtendedProducerProperties<KafkaProducerProperties>> provisioner; 

	public TopicCreatingHandler(ProvisioningProvider<ExtendedConsumerProperties<KafkaConsumerProperties>, ExtendedProducerProperties<KafkaProducerProperties>> provisioner) {
		this.provisioner = provisioner;
	}

	@EventListener
	public void onTopicAdded(ResourceAddedEvent<Topic> event) {
		Topic resource = event.getResource();
		String topicName = resource.getMetadata().getName();
		logger.info("adding topic: " + topicName);
		this.createTopic(topicName, 1);
	}

	@EventListener
	public void onTopicDeleted(ResourceDeletedEvent<Topic> event) {
		logger.info("topic deletion not yet supported");
	}

	private void createTopic(String topic, int partitionCount) {
		ExtendedProducerProperties<KafkaProducerProperties> producerProperties =
				new ExtendedProducerProperties<KafkaProducerProperties>(new KafkaProducerProperties());
		producerProperties.setPartitionCount(partitionCount);
		this.provisioner.provisionProducerDestination(topic, producerProperties);
	}
}
