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

package io.sk8s.function.dispatcher;

import java.io.UnsupportedEncodingException;
import java.util.Collections;
import java.util.HashMap;
import java.util.Map;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.cloud.stream.binder.Binder;
import org.springframework.cloud.stream.binder.BinderFactory;
import org.springframework.cloud.stream.binder.Binding;
import org.springframework.cloud.stream.binder.ExtendedConsumerProperties;
import org.springframework.cloud.stream.binder.ExtendedProducerProperties;
import org.springframework.cloud.stream.binder.HeaderMode;
import org.springframework.cloud.stream.binder.kafka.properties.KafkaConsumerProperties;
import org.springframework.cloud.stream.binder.kafka.properties.KafkaProducerProperties;
import org.springframework.integration.channel.DirectChannel;
import org.springframework.integration.support.MessageBuilder;
import org.springframework.messaging.MessageChannel;
import org.springframework.messaging.MessageHandler;
import org.springframework.messaging.MessageHeaders;

import io.fabric8.kubernetes.api.model.ContainerBuilder;
import io.fabric8.kubernetes.api.model.JobSpecBuilder;
import io.fabric8.kubernetes.api.model.VolumeBuilder;
import io.fabric8.kubernetes.api.model.VolumeMountBuilder;
import io.fabric8.kubernetes.client.DefaultKubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClient;

import io.sk8s.core.function.FunctionResource;
import io.sk8s.core.function.FunctionResourceHandler;

/**
 * @author Mark Fisher
 */
public class FunctionDispatchingHandler implements FunctionResourceHandler {

	private static Log logger = LogFactory.getLog(FunctionDispatchingHandler.class);

	@Autowired
	private FunctionDispatcherProperties properties;

	private final Map<String, Binding<?>> bindings = new HashMap<>();

	private final Binder<MessageChannel, ExtendedConsumerProperties<KafkaConsumerProperties>, ExtendedProducerProperties<KafkaProducerProperties>> binder;

	private final KubernetesClient kubernetesClient = new DefaultKubernetesClient();

	private final MessageHandler messageHandler = m -> {
		try {
			String payload = new String((byte[]) m.getPayload(), "UTF-8");
			launchJob(m.getHeaders().get("function", String.class), payload);
		}
		catch (UnsupportedEncodingException e) {
			e.printStackTrace();
		}
	};

	@SuppressWarnings("unchecked")
	public FunctionDispatchingHandler(BinderFactory binderFactory) {
		this.binder = (Binder<MessageChannel, ExtendedConsumerProperties<KafkaConsumerProperties>, ExtendedProducerProperties<KafkaProducerProperties>>)
				binderFactory.getBinder("kafka", MessageChannel.class);
	}

	@Override
	public void resourceAdded(FunctionResource resource) {
		String functionName = resource.getMetadata().get("name");
		addListener(resource);
		logger.info("function added: " + functionName);
	}

	private void addListener(FunctionResource resource) {
		String destination = resource.getSpec().getTopic();
		final String functionName = resource.getMetadata().get("name");
		DirectChannel channel = new DirectChannel();
		channel.subscribe(m -> {
			messageHandler.handleMessage(MessageBuilder.fromMessage(m)
					.setHeader("function", functionName)
					.setHeader(MessageHeaders.CONTENT_TYPE, "text/plain")
					.build());
		});
		Binding<?> binding = this.binder.bindConsumer(destination, functionName, channel, consumerProperties());
		this.bindings.put(functionName, binding);
	}

	private ExtendedConsumerProperties<KafkaConsumerProperties> consumerProperties() {
		KafkaConsumerProperties kafkaProps = new KafkaConsumerProperties();
		ExtendedConsumerProperties<KafkaConsumerProperties> extendedProps = new ExtendedConsumerProperties<>(kafkaProps);
		extendedProps.setHeaderMode(HeaderMode.raw);
		return extendedProps;
	}

	@Override
	public void resourceDeleted(FunctionResource resource) {
		String functionName = resource.getMetadata().get("name");
		Binding<?> binding = this.bindings.remove(functionName);
		if (binding != null) {
			binding.unbind();
		}
		logger.info("function deleted: " + functionName);
	}

	private void launchJob(String functionName, String payload) {
		String s = this.kubernetesClient.extensions().jobs().inNamespace(this.properties.getNamespace()).createNew()
			.withApiVersion("batch/v1")
			.withNewMetadata()
				.withName(functionName + "-" + System.currentTimeMillis())
			.endMetadata()
			.withSpec(new JobSpecBuilder()
				.withNewTemplate()
					.withNewMetadata()
						.withLabels(Collections.singletonMap("function", functionName))
					.endMetadata()
					.withNewSpec()
						.withRestartPolicy("OnFailure")
						.withActiveDeadlineSeconds(10L)
						.withContainers(new ContainerBuilder()
							.withName("main")
							.withImage("busybox")
							.withCommand("/bin/sh")
							.withArgs("-c", "echo " + payload + " >> /output/" + functionName + ".log")
							.withVolumeMounts(new VolumeMountBuilder()
								.withMountPath("/output")
								.withName("messages")
							.build())
						.build())
						.withVolumes(new VolumeBuilder()
							.withName("messages")
							.withNewHostPath("/messages")
						.build())
					.endSpec()
				.endTemplate()
			.build())
			.done().toString();
		System.out.println("JOB: " + s);
	}
}
