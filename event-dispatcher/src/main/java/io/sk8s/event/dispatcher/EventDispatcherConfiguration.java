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
import io.fabric8.kubernetes.client.Watch;
import io.sk8s.core.resource.ResourceEventPublisher;
import io.sk8s.kubernetes.client.Sk8sClient;

import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.cloud.stream.annotation.EnableBinding;
import org.springframework.cloud.stream.binder.BinderFactory;
import org.springframework.context.ApplicationEventPublisher;
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
	public EventDispatchingHandler eventDispatchingHandler(BinderFactory binderFactory) {
		return new EventDispatchingHandler(binderFactory);
	}

	@Bean
	public ResourceEventPublisher functionsEventPublisher(Sk8sClient client) {
		return new ResourceEventPublisher(client.functions());
	}

	@Bean
	public ResourceEventPublisher handlersEventPublisher(Sk8sClient client) {
		return new ResourceEventPublisher(client.handlers());
	}

	@Bean
	public KubernetesClient kubernetesClient() {
		return new DefaultKubernetesClient();
	}

	@Bean
	public Sk8sClient sk8sClient(KubernetesClient kubernetesClient) {
		return kubernetesClient.adapt(Sk8sClient.class);
	}

	// Handler Pool
	@Bean
	public HandlerPool handlerPool(KubernetesClient kubernetesClient, BinderFactory binderFactory) {
		return new HandlerPool(kubernetesClient, binderFactory);
	}

	@Bean
	public ResourceEventPublisher podsEventPublisher(KubernetesClient client) {
		return new ResourceEventPublisher(client.pods());
	}

	@Bean
	public ResourceEventPublisher servicesEventPublisher(KubernetesClient client) {
		return new ResourceEventPublisher(client.services());
	}

	@Bean
	public ResourceEventPublisher endpointsEventPublisher(KubernetesClient client) {
		return new ResourceEventPublisher(client.endpoints());
	}

	@Bean
	public ResourceEventPublisher deploymentsEventPublisher(KubernetesClient client) {
		return new ResourceEventPublisher(client.extensions().deployments());
	}

	// Job Launcher
	@Bean
	public JobLauncher jobLauncher(KubernetesClient kubernetesClient) {
		return new JobLauncher(kubernetesClient);
	}

	// Service Invoker
	@Bean
	public ServiceInvoker serviceInvoker() {
		return new ServiceInvoker();
	}
}
