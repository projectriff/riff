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

import java.util.HashMap;
import java.util.Map;

import org.apache.kafka.clients.producer.ProducerConfig;
import org.apache.kafka.common.serialization.ByteArraySerializer;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.kafka.core.DefaultKafkaProducerFactory;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.kafka.core.ProducerFactory;

/**
 * @author Mark Fisher
 */
@Configuration
@EnableConfigurationProperties(TopicGatewayProperties.class)
public class TopicGatewayConfiguration {

	@Autowired
	private TopicGatewayProperties properties;

	@Bean
	public MessagePublisher publisher(KafkaTemplate<String, byte[]> kafkaTemplate) {
		return new MessagePublisher(kafkaTemplate);
	}

	@Bean
	public KafkaTemplate<String, byte[]> kafkaTemplate(ProducerFactory<String, byte[]> producerFactory) {
		return new KafkaTemplate<>(producerFactory);
	}

	@Bean
	public ProducerFactory<String, byte[]> producerFactory() {
		return new DefaultKafkaProducerFactory<>(producerProps());
	}

	private Map<String, Object> producerProps() {
	    Map<String, Object> props = new HashMap<>();
	    props.put(ProducerConfig.BOOTSTRAP_SERVERS_CONFIG, this.properties.getBrokers());
	    props.put(ProducerConfig.RETRIES_CONFIG, 0);
	    props.put(ProducerConfig.BATCH_SIZE_CONFIG, 16384);
	    props.put(ProducerConfig.LINGER_MS_CONFIG, 1);
	    props.put(ProducerConfig.BUFFER_MEMORY_CONFIG, 33554432);
	    props.put(ProducerConfig.KEY_SERIALIZER_CLASS_CONFIG, ByteArraySerializer.class);
	    props.put(ProducerConfig.VALUE_SERIALIZER_CLASS_CONFIG, ByteArraySerializer.class);
	    return props;
	}
}
