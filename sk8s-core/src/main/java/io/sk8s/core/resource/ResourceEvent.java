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

import io.fabric8.kubernetes.api.model.HasMetadata;
import io.fabric8.kubernetes.client.Watcher;

import org.springframework.core.ResolvableType;
import org.springframework.core.ResolvableTypeProvider;
import org.springframework.util.Assert;

/**
 * Fired by {@link ResourceEventPublisher} when a change happens on a watched resource.
 * Contains references to the {@link io.fabric8.kubernetes.client.Watcher.Action} and the resource.
 *
 * @param <T> the type of resource that is the subject of change.
 * @author Eric Bottard
 */
public class ResourceEvent<T> implements ResolvableTypeProvider {

	private final T resource;

	private final Watcher.Action action;

	ResourceEvent(T resource, Watcher.Action action) {
		Assert.notNull(resource, "Resource cannot be null");
		Assert.notNull(action, "Action cannot be null");
		this.resource = resource;
		this.action = action;
	}

	@Override
	public ResolvableType getResolvableType() {
		return ResolvableType.forClassWithGenerics(getClass(),
				ResolvableType.forInstance(getResource()));
	}

	public Watcher.Action getAction() {
		return action;
	}

	public T getResource() {
		return resource;
	}

	@Override
	public String toString() {
		if (resource instanceof HasMetadata) {
			HasMetadata hasMetadata = (HasMetadata) resource;
			return String.format("%s<%s>(%s)", getClass().getSimpleName(), resource.getClass().getSimpleName(), hasMetadata.getMetadata().getName());
		} else {
			return String.format("%s<%s>(%s)", getClass().getSimpleName(), resource.getClass().getSimpleName(), resource);
		}
	}
}
