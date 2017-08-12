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

import org.springframework.http.ResponseEntity;
import org.springframework.util.Assert;
import org.springframework.web.client.RestTemplate;

import io.sk8s.core.function.FunctionResource;
import io.sk8s.core.handler.HandlerResource;

/**
 * @author Mark Fisher
 */
public class ServiceInvoker implements Dispatcher {

	private final RestTemplate restTemplate = new RestTemplate();

	@Override
	public void init(FunctionResource functionResource, HandlerResource handlerResource) {
	}

	@Override
	public void dispatch(String payload, Map<String, Object> headers, FunctionResource functionResource, HandlerResource handlerResource) {
		// TODO: create the Handler instance by passing params
		String url = functionResource.getSpec().getParam("url");
		Assert.hasText(url, "no url provided for function " + functionResource.getMetadata().get("name"));
		ResponseEntity<?> response = this.restTemplate.postForEntity(url, payload, String.class);
		System.out.println("response: "+ response);
	}
}
