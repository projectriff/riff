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

import java.util.Map;

import io.sk8s.kubernetes.api.model.Handler;
import io.sk8s.kubernetes.api.model.XFunction;

/**
 * Strategy interface for deciding how to handle function application using some handler.
 *
 * <p>Implementations may decide to spawn a one off container, or maintain a pool, <em>etc.</em></p>
 *
 * @author Mark Fisher
 * @author Eric Bottard
 */
public interface Dispatcher {

	/**
	 * Called when a new function is registered, to let the dispatcher set up infrastructure if needed.
	 *
	 * <p>Default implementation does nothing.</p>
	 *
	 * @param functionResource the new function
	 * @param handlerResource the handler used by the function
	 */
	default void init(XFunction functionResource, Handler handlerResource) {}

	/**
	 * Called when a function is un-registered, to let the dispatcher tear down infrastructure if needed.
	 *
	 * <p>Default implementation does nothing.</p>
	 *
	 * @param functionResource the function that is going away
	 * @param handlerResource the handler used by the function
	 */
	default void destroy(XFunction functionResource, Handler handlerResource) {}

	/**
	 * Perform actual invocation of the function.
	 *
	 * @param payload the input to the function
	 * @param headers message headers received as part of invocation request
	 * @param functionResource the function to apply
	 * @param handlerResource the handler used by the function
	 */
	void dispatch(String payload, Map<String, Object> headers, XFunction functionResource, Handler handlerResource);

}
