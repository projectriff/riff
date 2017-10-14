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

import java.io.UnsupportedEncodingException;
import java.nio.charset.StandardCharsets;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.messaging.Message;
import org.springframework.messaging.MessageHeaders;
import org.springframework.messaging.support.MessageBuilder;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

/**
 * @author Mark Fisher
 */
@RestController
public class TopicGatewayController {

	@Autowired
	private MessagePublisher publisher;

	@PostMapping("/messages/{topic}")
	public String publishMessage(@PathVariable String topic, @RequestBody String payload) throws UnsupportedEncodingException {
		Message<byte[]> message = MessageBuilder.withPayload(payload.getBytes(StandardCharsets.UTF_8.name()))
				.setHeader("topic", topic)
				.setHeader(MessageHeaders.REPLY_CHANNEL, "foo")
				.build();
		this.publisher.publishMessage(topic, message);
		return "message published to topic: " + topic + "\n";
	}
}
