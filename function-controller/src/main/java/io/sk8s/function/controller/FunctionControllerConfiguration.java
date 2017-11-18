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

import java.util.HashMap;
import java.util.Map;

import org.apache.kafka.clients.producer.ProducerConfig;
import org.apache.kafka.common.serialization.ByteArraySerializer;

import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.cloud.stream.annotation.EnableBinding;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.kafka.core.DefaultKafkaProducerFactory;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.kafka.core.ProducerFactory;

import io.fabric8.kubernetes.client.DefaultKubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClient;

import io.sk8s.core.resource.ResourceEventPublisher;
import io.sk8s.kubernetes.client.Sk8sClient;

/**
 * @author Mark Fisher
 */
@Configuration
@EnableBinding
@EnableConfigurationProperties(SidecarProperties.class)
public class FunctionControllerConfiguration {

	@Bean
	public FunctionMonitor functionMonitor() {
		return new FunctionMonitor();
	}

	@Bean
	public ResourceEventPublisher functionEventPublisher(Sk8sClient client) {
		return new ResourceEventPublisher(client.functions());
	}

	@Bean
	public ResourceEventPublisher deploymentEventPublisher(KubernetesClient client) {
		return new ResourceEventPublisher(client.extensions().deployments());
	}

	@Bean
	public ResourceEventPublisher topicEventPublisher(Sk8sClient client) {
		return new ResourceEventPublisher(client.topics());
	}

	@Bean
	public KubernetesClient kubernetesClient() {
		return new DefaultKubernetesClient();
	}

	@Bean
	public Sk8sClient sk8sClient(KubernetesClient kubernetesClient) {
		return kubernetesClient.adapt(Sk8sClient.class);
	}

	@Bean
	public FunctionDeployer functionDeployer(KubernetesClient kubernetesClient) {
		return new FunctionDeployer(kubernetesClient);
	}

	@Bean
	public EventPublisher eventPublisher(KafkaTemplate<String, byte[]> kafkaTemplate) {
		return new EventPublisher(kafkaTemplate);
	}

	@Bean
	public KafkaTemplate<String, byte[]> kafkaTemplate(ProducerFactory<String, byte[]> producerFactory) {
		return new KafkaTemplate<>(producerFactory);
	}

	@Bean
	public ProducerFactory<String, byte[]> producerFactory() {
		return new DefaultKafkaProducerFactory<>(producerProps());
	}

	@Bean
	public Map<String, Object> producerProps() {
	    Map<String, Object> props = new HashMap<>();
	    props.put(ProducerConfig.BOOTSTRAP_SERVERS_CONFIG, System.getenv("SPRING_CLOUD_STREAM_KAFKA_BINDER_BROKERS"));
	    props.put(ProducerConfig.RETRIES_CONFIG, 0);
	    props.put(ProducerConfig.BATCH_SIZE_CONFIG, 16384);
	    props.put(ProducerConfig.LINGER_MS_CONFIG, 1);
	    props.put(ProducerConfig.BUFFER_MEMORY_CONFIG, 33554432);
	    props.put(ProducerConfig.KEY_SERIALIZER_CLASS_CONFIG, ByteArraySerializer.class);
	    props.put(ProducerConfig.VALUE_SERIALIZER_CLASS_CONFIG, ByteArraySerializer.class);
	    return props;
	}
}
