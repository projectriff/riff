/*
 * Copyright 2012-2015 the original author or authors.
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

import java.lang.reflect.Field;
import java.net.URL;
import java.util.Map;

import org.springframework.boot.builder.SpringApplicationBuilder;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.core.env.MapPropertySource;
import org.springframework.core.env.StandardEnvironment;
import org.springframework.util.ReflectionUtils;

/**
 * Utility class for starting a Spring Boot application in a separate thread. Best used
 * from an isolated class loader, e.g. through {@link ApplicationRunner}.
 * 
 * @author Dave Syer
 *
 */
public class ContextRunner {

	private ConfigurableApplicationContext context;
	private Thread runThread;
	private boolean running = false;
	private Throwable error;
	private long timeout = 120000;

	public void run(final String source, final Map<String, Object> properties,
			final String... args) {
		// Run in new thread to ensure that the context classloader is setup
		this.runThread = new Thread(new Runnable() {
			@Override
			public void run() {
				try {
					resetUrlHandler();
					StandardEnvironment environment = new StandardEnvironment();
					environment.getPropertySources().addAfter(
							StandardEnvironment.SYSTEM_ENVIRONMENT_PROPERTY_SOURCE_NAME,
							new MapPropertySource("appDeployer", properties));
					context = new SpringApplicationBuilder(source)
							.environment(environment).run(args);
				}
				catch (Throwable ex) {
					error = ex;
				}

			}
		});
		this.runThread.start();
		try {
			this.runThread.join(timeout);
			this.running = context != null && context.isRunning();
		}
		catch (InterruptedException e) {
			this.running = false;
			Thread.currentThread().interrupt();
		}

	}

	public void close() {
		if (this.context != null) {
			this.context.close();
			resetUrlHandler();
		}
		// TODO: JDBC leak protection?
		this.running = false;
		this.runThread = null;
	}

	public ConfigurableApplicationContext getContext() {
		return this.context;
	}

	private void resetUrlHandler() {
		// Tomcat always tries to set this, even if it was already set
		Field field = ReflectionUtils.findField(URL.class, "factory");
		ReflectionUtils.makeAccessible(field);
		ReflectionUtils.setField(field, null, null);
	}

	public boolean isRunning() {
		return running;
	}

	public Throwable getError() {
		return this.error;
	}

}