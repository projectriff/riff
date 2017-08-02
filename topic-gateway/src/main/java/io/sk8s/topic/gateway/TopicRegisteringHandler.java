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

package io.sk8s.topic.gateway;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

import io.sk8s.core.resource.ResourceEventHandler;
import io.sk8s.core.topic.TopicResource;

/**
 * @author Mark Fisher
 */
public class TopicRegisteringHandler implements ResourceEventHandler<TopicResource> {

	private static Log logger = LogFactory.getLog(TopicRegisteringHandler.class);

	@Override
	public void resourceAdded(TopicResource resource) {
		TopicResource.TopicSpec spec = resource.getSpec();
		if (spec.isExposeRead()) {
			// TODO: add to whitelist for GET mapping
		}
		if (spec.isExposeWrite()) {
			// TODO: add to whitelist for POST mapping
		}
	}

	@Override
	public void resourceDeleted(TopicResource resource) {
		// TODO: remove if present in either whitelist
	}
}
