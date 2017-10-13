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

package io.sk8s.invoker.java.server;

import java.util.function.Consumer;
import java.util.function.Function;
import java.util.function.Supplier;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

import org.springframework.beans.factory.InitializingBean;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.builder.SpringApplicationBuilder;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.http.HttpStatus;
import org.springframework.util.Assert;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;

import io.sk8s.invoker.java.function.FunctionConfiguration;

/**
 * @author Eric Bottard
 * @author Mark Fisher
 */
@RestController
public class JavaFunctionInvokerController implements InitializingBean {

	private static Log logger = LogFactory.getLog(JavaFunctionInvokerController.class);

	@Autowired
	private ConfigurableApplicationContext context;

	@Autowired
	private FunctionProperties properties;

	private volatile Object function;

	@Override
	public void afterPropertiesSet() throws Exception {
		String functionUri = properties.getUri();
		if (function == null) {
			logger.info("initializing with uri: " + functionUri);
			String[] tokens = functionUri.split("\\?");
			Assert.isTrue(tokens.length == 2, "expected format: [jarLocation]?[className]");
			String jarLocation = tokens[0];
			String className = tokens[1];
			ConfigurableApplicationContext functionContext = new SpringApplicationBuilder()
					.sources(FunctionConfiguration.class)
					.web(false)
					.parent(this.context)
					.run("--function.jarLocation=" + jarLocation,
							"--function.className=" + className);
			this.function = functionContext.getBean("function");
		}
	}

	@PostMapping("/")
	public Object invoke(@RequestBody Object body) {
		logger.info("invoke called with: " + body);
		if (this.function == null) {
			throw new FunctionNotInitializedException();
		}
		if (function instanceof Supplier) {
			return ((Supplier) function).get();
		}
		else if (function instanceof Function) {
			return ((Function) function).apply(body);
		}
		else if (function instanceof Consumer) {
			((Consumer) function).accept(body);
			return null;
		}
		else {
			throw new AssertionError();
		}
	}

	@ResponseStatus(HttpStatus.NOT_FOUND)
	public static class FunctionNotInitializedException extends RuntimeException {
	}

	/**
	 * The payload received to initialize a new function.
	 *
	 * @author Eric Bottard
	 */
	public static class InitPayload {
		private String uri;

		private String className;

		public void setUri(String uri) {
			this.uri = uri;
		}

		public void setClassName(String className) {
			this.className = className;
		}

		public String getUri() {
			return uri;
		}

		public String getClassName() {
			return className;
		}
	}
}
