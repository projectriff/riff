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
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.SynchronousQueue;
import java.util.concurrent.TimeUnit;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.cloud.stream.annotation.EnableBinding;
import org.springframework.cloud.stream.annotation.StreamListener;
import org.springframework.cloud.stream.messaging.Sink;
import org.springframework.messaging.Message;
import org.springframework.messaging.support.MessageBuilder;
import org.springframework.util.StringUtils;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

/**
 * @author Mark Fisher
 */
@RestController
@EnableBinding(Sink.class)
public class TopicGatewayController {

	private final Logger logger = LoggerFactory.getLogger(TopicGatewayController.class);

	@Autowired
	private MessagePublisher publisher;

	private final Map<String, SynchronousQueue<String>> replies = new HashMap<>();

	@PostMapping("/messages/{topic}")
	public String publishMessage(@PathVariable String topic, @RequestBody String payload) throws UnsupportedEncodingException {
		Message<byte[]> message = MessageBuilder.withPayload(payload.getBytes(StandardCharsets.UTF_8.name()))
				.setHeader("topic", topic)
				.build();
		this.publisher.publishMessage(topic, message);
		return "message published to topic: " + topic + "\n";
	}

	@PostMapping("/requests/{topic}")
	public String publishRequest(@PathVariable String topic, @RequestBody String payload) throws UnsupportedEncodingException {
		String correlationId = UUID.randomUUID().toString();
		Message<byte[]> message = MessageBuilder.withPayload(payload.getBytes(StandardCharsets.UTF_8.name()))
				.setHeader("topic", topic)
				.setHeader("correlationId", correlationId)
				.build();
		SynchronousQueue<String> queue = new SynchronousQueue<>();
		this.replies.put(correlationId, queue);
		this.publisher.publishMessage(topic, message);
		this.logger.debug("message published to '%s' with correlationId: %s", topic, correlationId);
		try {
			String reply = queue.poll(60, TimeUnit.SECONDS);
			return reply + "\n";
		}
		catch (InterruptedException e) {
			Thread.currentThread().interrupt();
		}
		finally {
			this.replies.remove(correlationId);
		}
		// TODO: throw exception 
		return "timed out waiting for reply for correlationId: " + correlationId;
	}

	@StreamListener(Sink.INPUT)
	public void handleReply(Message<byte[]> reply) {
		String correlationId = reply.getHeaders().get("correlationId", String.class);
		if (StringUtils.hasText(correlationId)) {
			try {
				this.replies.get(correlationId).put(new String(reply.getPayload(), StandardCharsets.UTF_8.name()));
			}
			catch (UnsupportedEncodingException e) {
				// ignore
			}
			catch (InterruptedException e) {
				Thread.currentThread().interrupt();
			}
		}
	}
}
