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

package io.sk8s.core.resource;

import io.fabric8.kubernetes.client.KubernetesClientException;
import io.fabric8.kubernetes.client.Watcher;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import org.springframework.context.ApplicationEventPublisher;

public class ResourceEventPublisher<T> implements Watcher<T> {

	private static final Logger log = LoggerFactory.getLogger(ResourceEventPublisher.class);

	private final ApplicationEventPublisher applicationEventPublisher;

	public ResourceEventPublisher(ApplicationEventPublisher applicationEventPublisher) {
		this.applicationEventPublisher = applicationEventPublisher;
	}

	@Override
	public void eventReceived(Action action, T resource) {
		switch (action) {
			case ADDED:
				applicationEventPublisher.publishEvent(new ResourceAddedEvent<>(resource));
				break;
			case DELETED:
				applicationEventPublisher.publishEvent(new ResourceDeletedEvent<>(resource));
				break;
			case MODIFIED:
				applicationEventPublisher.publishEvent(new ResourceModifiedEvent<>(resource));
			case ERROR:
				applicationEventPublisher.publishEvent(new ResourceErrorEvent<>(resource));
				break;
			default:
				log.warn("Unsupported event action {} received for resource {}", action, resource);
		}
	}

	@Override
	public void onClose(KubernetesClientException cause) {

	}
}
