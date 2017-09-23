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

import io.fabric8.kubernetes.client.Watcher;

import org.springframework.core.ResolvableType;
import org.springframework.core.ResolvableTypeProvider;

public class WatcherEvent<T> implements ResolvableTypeProvider {

	private final T resource;

	private final Watcher.Action action;

	public WatcherEvent(T resource, Watcher.Action action) {
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
}
