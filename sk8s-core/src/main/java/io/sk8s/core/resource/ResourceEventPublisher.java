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

import javax.annotation.PreDestroy;

import io.fabric8.kubernetes.api.model.HasMetadata;
import io.fabric8.kubernetes.api.model.KubernetesResourceList;
import io.fabric8.kubernetes.client.KubernetesClientException;
import io.fabric8.kubernetes.client.Watch;
import io.fabric8.kubernetes.client.Watcher;
import io.fabric8.kubernetes.client.dsl.MixedOperation;
import io.fabric8.kubernetes.client.dsl.Resource;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.ApplicationEventPublisher;
import org.springframework.context.event.ContextRefreshedEvent;
import org.springframework.context.event.EventListener;

/**
 * Will publish variations of {@link ResourceEvent} to an {@link ApplicationEventPublisher}
 * by adapting a {@link Watcher} to the Spring event model.
 *
 * <p>After creation, a catchup mechanism will fire {@link io.fabric8.kubernetes.client.Watcher.Action#ADDED}
 * events for all items that were already there.</p>
 *
 * @author Eric Bottard
 */
public class ResourceEventPublisher<T extends HasMetadata, L extends KubernetesResourceList<T>, D, R extends Resource<T, D>> {

	private static final Logger log = LoggerFactory.getLogger(ResourceEventPublisher.class);

	private ApplicationEventPublisher applicationEventPublisher;

	private Watch watch;

	private final MixedOperation<T, L, D, R> watchListDeletable;

	public ResourceEventPublisher(MixedOperation<T, L, D, R> watchListDeletable) {
		this.watchListDeletable = watchListDeletable;
	}

	@Autowired
	public void setApplicationEventPublisher(ApplicationEventPublisher applicationEventPublisher) {
		this.applicationEventPublisher = applicationEventPublisher;
	}

	/**
	 * Create a Watcher that will get notified on new events on a watchable, after having listed all items of said
	 * watchable and artificially fired an {@link io.fabric8.kubernetes.client.Watcher.Action#ADDED} event for all of them. This allows listeners to
	 * catch up with all items that were already there before they registered.
	 */
	@EventListener(ContextRefreshedEvent.class)
	public void lateInit() {
		ApplicationEventPublishingWatcher<T> publisher = new ApplicationEventPublishingWatcher<>();
		L list = watchListDeletable.list();
		String from = list.getMetadata().getResourceVersion();
		list.getItems().forEach(i -> publisher.eventReceived(Watcher.Action.ADDED, (T)i));
		watch = watchListDeletable.withResourceVersion(from).watch(publisher);
	};

	@PreDestroy
	public void destroy() {
		watch.close();
	}

	private class ApplicationEventPublishingWatcher<T> implements Watcher<T> {

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
}
