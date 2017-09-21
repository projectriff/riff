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

import java.lang.reflect.Field;
import java.util.concurrent.atomic.AtomicReference;

import io.fabric8.kubernetes.client.DefaultKubernetesClient;
import io.fabric8.kubernetes.client.Watch;
import io.sk8s.core.resource.ResourceEventPublisher;
import io.sk8s.kubernetes.client.Sk8sClient;

import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.cloud.stream.annotation.EnableBinding;
import org.springframework.cloud.stream.binder.BinderFactory;
import org.springframework.cloud.stream.binder.ExtendedConsumerProperties;
import org.springframework.cloud.stream.binder.ExtendedProducerProperties;
import org.springframework.cloud.stream.binder.kafka.properties.KafkaConsumerProperties;
import org.springframework.cloud.stream.binder.kafka.properties.KafkaProducerProperties;
import org.springframework.cloud.stream.provisioning.ProvisioningProvider;
import org.springframework.context.ApplicationEventPublisher;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.messaging.MessageChannel;
import org.springframework.util.ReflectionUtils;
import org.springframework.util.ReflectionUtils.FieldCallback;

/**
 * @author Mark Fisher
 */
@Configuration
@EnableBinding
@EnableConfigurationProperties(TopicControllerProperties.class)
public class TopicControllerConfiguration {

	@Bean
	public Sk8sClient sk8sClient() {
		return new DefaultKubernetesClient().adapt(Sk8sClient.class);
	}

	@Bean
	public ResourceEventPublisher topicsEventPublisher(Sk8sClient client) {
		return new ResourceEventPublisher(client.topics());
	}

	@Bean
	public TopicCreatingHandler topicCreatingHandler(BinderFactory binderFactory) {
		return new TopicCreatingHandler(topicProvisioner(binderFactory));
	}

	@SuppressWarnings({ "unchecked", "rawtypes" })
	private ProvisioningProvider<ExtendedConsumerProperties<KafkaConsumerProperties>,
			ExtendedProducerProperties<KafkaProducerProperties>> topicProvisioner(BinderFactory binderFactory) {
		Object binder = binderFactory.getBinder("kafka", MessageChannel.class);
		final AtomicReference<Field> provisionerField = new AtomicReference<>();
		ReflectionUtils.doWithFields(binder.getClass(), new FieldCallback() {

			@Override
			public void doWith(Field field) throws IllegalArgumentException, IllegalAccessException {
				if (ProvisioningProvider.class.isAssignableFrom(field.getType())) {
					field.setAccessible(true);
					provisionerField.set(field);
				}
			}
		});
		return (ProvisioningProvider) ReflectionUtils.getField(provisionerField.get(), binder);
	}
}
