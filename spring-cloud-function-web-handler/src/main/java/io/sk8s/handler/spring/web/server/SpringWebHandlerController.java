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

package io.sk8s.handler.spring.web.server;

import java.util.concurrent.atomic.AtomicBoolean;
import java.util.function.Function;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.builder.SpringApplicationBuilder;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;

import io.sk8s.handler.spring.web.invoker.FunctionConfiguration;

import reactor.core.publisher.Flux;

/**
 * @author Mark Fisher
 */
@RestController
public class SpringWebHandlerController {

	private static Log logger = LogFactory.getLog(SpringWebHandlerController.class);

	@Autowired
	private ConfigurableApplicationContext context;

	private Function<Flux<String>, Flux<String>> function;

	private final AtomicBoolean initialized = new AtomicBoolean();

	public void setFunction(Function<Flux<String>, Flux<String>> function) {
		this.function = function;
	}

	@PostMapping("/init")
	@SuppressWarnings("unchecked")
	public void init(@RequestBody String uri, @RequestParam String classname) {
		if (initialized.compareAndSet(false, true)) {
			logger.info("init called with uri: " + uri);
			// TODO: use JSON with both uri and classname (eventually dependencies)
			ConfigurableApplicationContext functionContext = new SpringApplicationBuilder()
					.sources(FunctionConfiguration.class)
					.web(false)
					.parent(this.context)
					.run("--function.uri=" + uri,
							"--function.classname=" + classname);
			this.function = functionContext.getBean("function", Function.class);
		}
	}

	@PostMapping("/invoke")
	public String invoke(@RequestBody String body) {
		logger.info("invoke called with: " + body);
		if (this.function == null) {
			throw new FunctionNotInitializedException();
		}
		return this.function.apply(Flux.just(body)).blockFirst();
	}

	@ResponseStatus(HttpStatus.NOT_FOUND)
	public static class FunctionNotInitializedException extends RuntimeException {
	}
}
