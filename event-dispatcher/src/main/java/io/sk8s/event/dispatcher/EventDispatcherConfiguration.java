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

	@Bean(destroyMethod = "close" /*explicit to make sure this stays as a bean*/)
	public Watch functionEventPublisher(Sk8sClient client, ApplicationEventPublisher eventPublisher) {
		return client.functions().watch(new ResourceEventPublisher<>(eventPublisher));
	}

	@Bean
	public KubernetesClient kubernetesClient() {
		return new DefaultKubernetesClient();
	}

	@Bean
	public HandlerPool handlerPool(KubernetesClient kubernetesClient, BinderFactory binderFactory) {
		return new HandlerPool(kubernetesClient, binderFactory);
	}

	@Bean
	public JobLauncher jobLauncher(KubernetesClient kubernetesClient) {
		return new JobLauncher(kubernetesClient);
	}

	@Bean
	public ServiceInvoker serviceInvoker() {
		return new ServiceInvoker();
	}
}
