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

package io.sk8s.function.controller;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import org.springframework.kafka.core.KafkaTemplate;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

/**
 * Publishes Events on a Topic.
 *
 * @author Mark Fisher
 */
public class EventPublisher {

	private final KafkaTemplate<String, String> kafkaTemplate;

	private final ObjectMapper mapper = new ObjectMapper();

	private final Logger logger = LoggerFactory.getLogger(EventPublisher.class);

	public EventPublisher(KafkaTemplate<String, String> kafkaTemplate) {
		this.kafkaTemplate = kafkaTemplate;
	}

	public <T> void publish(String topic, Object event) {
		try {
			this.kafkaTemplate.send(topic, this.mapper.writeValueAsString(event));
		}
		catch (JsonProcessingException e) {
			logger.warn("failed to publish event", e);;
		}
	}
}
