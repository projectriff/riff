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

package io.sk8s.sidecar;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.cloud.stream.annotation.EnableBinding;
import org.springframework.cloud.stream.annotation.StreamListener;
import org.springframework.cloud.stream.messaging.Processor;
import org.springframework.integration.support.MessageBuilder;
import org.springframework.messaging.Message;
import org.springframework.messaging.handler.annotation.SendTo;

/**
 * @author Mark Fisher
 */
@EnableBinding(Processor.class)
public class EventHandler {

	@Autowired
	private Dispatcher dispatcher;

	@StreamListener(Processor.INPUT)
	@SendTo(Processor.OUTPUT)
	public Message<byte[]> invoke(Message<byte[]> input) {
		try {
			String payload = new String(input.getPayload(), "UTF-8");
			System.out.println("SIDECAR request: " + payload);
			String output = this.dispatcher.dispatch(payload);
			System.out.println("SIDECAR response: " + output);
			return MessageBuilder.withPayload(output.getBytes("UTF-8")).copyHeadersIfAbsent(input.getHeaders()).build();
		}
		catch (Exception e) {
			throw new RuntimeException("failed to dispatch event to function invoker", e);
		}
	}
}
