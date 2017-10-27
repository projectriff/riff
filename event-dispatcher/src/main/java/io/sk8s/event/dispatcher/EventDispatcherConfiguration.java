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

package io.sk8s.event.dispatcher;

import io.fabric8.kubernetes.client.DefaultKubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.sk8s.core.resource.ResourceEventPublisher;
import io.sk8s.kubernetes.client.Sk8sClient;

import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.cloud.stream.annotation.EnableBinding;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/**
 * @author Mark Fisher
 */
@Configuration
@EnableBinding
@EnableConfigurationProperties(EventDispatcherProperties.class)
public class EventDispatcherConfiguration {

	@Bean
	public EventDispatchingHandler eventDispatchingHandler() {
		return new EventDispatchingHandler();
	}

	@Bean
	public ResourceEventPublisher functionsEventPublisher(Sk8sClient client) {
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
}
