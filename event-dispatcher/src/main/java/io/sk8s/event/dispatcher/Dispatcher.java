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

import io.sk8s.core.function.FunctionResource;
import io.sk8s.core.handler.HandlerResource;

/**
 * @author Mark Fisher
 */
public interface Dispatcher {

	void init(FunctionResource functionResource, HandlerResource handlerResource);

	void dispatch(String payload, Map<String, Object> headers, FunctionResource functionResource, HandlerResource handlerResource);

}
