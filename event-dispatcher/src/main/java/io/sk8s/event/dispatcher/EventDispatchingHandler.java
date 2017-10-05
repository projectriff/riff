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

import java.util.HashMap;
import java.util.Map;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.event.EventListener;

import io.sk8s.core.resource.ResourceAddedEvent;
import io.sk8s.core.resource.ResourceDeletedEvent;
import io.sk8s.kubernetes.api.model.XFunction;

/**
 * @author Mark Fisher
 */
public class EventDispatchingHandler {

	private final Log logger = LogFactory.getLog(EventDispatchingHandler.class);

	// TODO: Change key to ObjectReference or similar
	private final Map<String, XFunction> functions = new HashMap<>();

	@Autowired // TODO: merge in here?
	private JobLauncher launcher;

	@EventListener
	public void onFunctionRegistered(ResourceAddedEvent<XFunction> event) {
		XFunction functionResource = event.getResource();
		String functionName = functionResource.getMetadata().getName();
		this.functions.put(functionName, functionResource);
		this.launcher.dispatch(functionResource);
		logger.info("function added: " + functionName);
	}

	@EventListener
	public void onFunctionUnregistered(ResourceDeletedEvent<XFunction> event) {
		XFunction functionResource = event.getResource();
		String functionName = functionResource.getMetadata().getName();
		// TODO: delete any Job(s)
		this.functions.remove(functionName);
		logger.info("function deleted: " + functionName);
	}
}
