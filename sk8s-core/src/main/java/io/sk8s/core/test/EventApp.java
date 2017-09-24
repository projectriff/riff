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

package io.sk8s.core.test;

import io.fabric8.kubernetes.client.DefaultKubernetesClient;
import io.fabric8.kubernetes.client.Watch;
import io.sk8s.core.resource.ResourceEventPublisher;
import io.sk8s.core.resource.ResourceEvent;
import io.sk8s.kubernetes.api.model.Topic;
import io.sk8s.kubernetes.client.Sk8sClient;

import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.builder.SpringApplicationBuilder;
import org.springframework.context.ApplicationEventPublisher;
import org.springframework.context.annotation.Bean;
import org.springframework.context.event.EventListener;
import org.springframework.stereotype.Component;

@SpringBootApplication
public class EventApp {

	public static void main(String[] args) {
		new SpringApplicationBuilder(EventApp.class).web(false).build().run(args);
	}

	@Bean
	public Sk8sClient sk8sClient() {
		return new DefaultKubernetesClient().adapt(Sk8sClient.class);
	}

	@Bean
	public ResourceEventPublisher toplicsEventPublisher(Sk8sClient client) {
		return new ResourceEventPublisher<>(client.topics());
	}

	@Component
	public static class Client {

		@EventListener
		public void onTopicAdded(ResourceEvent<Topic> event) {
			System.out.println("Received " + event);
		}

	}
}
