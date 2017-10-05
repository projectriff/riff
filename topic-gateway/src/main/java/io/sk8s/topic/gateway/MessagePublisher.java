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

package io.sk8s.topic.gateway;

import org.springframework.cloud.stream.binder.EmbeddedHeaderUtils;
import org.springframework.cloud.stream.binder.MessageValues;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.messaging.Message;
import org.springframework.messaging.MessageHeaders;

/**
 * @author Mark Fisher
 */
public class MessagePublisher {

	private final KafkaTemplate<String, byte[]> kafkaTemplate;

	public MessagePublisher(KafkaTemplate<String, byte[]> kafkaTemplate) {
		this.kafkaTemplate = kafkaTemplate;
	}

	public void publishMessage(String topic, Message message) {
		try {
			byte[] bytes = EmbeddedHeaderUtils.embedHeaders(new MessageValues(message), MessageHeaders.REPLY_CHANNEL);
			this.kafkaTemplate.send(topic, bytes);
		}
		catch (Exception e) {
			throw new RuntimeException(e);
		}
	}
}
