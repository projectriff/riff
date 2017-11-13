/*
 * Copyright 2016-2017 the original author or authors.
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

import java.lang.reflect.Method;
import java.util.Collections;
import java.util.Map;
import java.util.UUID;

import javax.annotation.PreDestroy;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

import org.springframework.context.support.LiveBeansView;
import org.springframework.expression.Expression;
import org.springframework.expression.spel.standard.SpelExpressionParser;
import org.springframework.expression.spel.support.StandardEvaluationContext;
import org.springframework.util.ClassUtils;
import org.springframework.util.ReflectionUtils;

/**
 * Driver class for running a Spring Boot application via an isolated classpath.
 * Initialize an instance of this class with the class loader to be used and the name of
 * the main class (usually a <code>@SpringBootApplication</code>), and then
 * {@link #run(String...)} it, cleaning up with a call to {@link #close()}.
 * 
 * @author Dave Syer
 *
 */
public class ApplicationRunner {

	private static Log logger = LogFactory.getLog(ApplicationRunner.class);

	private final ClassLoader classLoader;

	private final String source;

	private Object app;

	public ApplicationRunner(ClassLoader classLoader, String source) {
		this.classLoader = classLoader;
		this.source = source;
	}

	public void run(String... args) {
		ClassLoader contextLoader = Thread.currentThread().getContextClassLoader();
		try {
			ClassUtils.overrideThreadContextClassLoader(this.classLoader);
			Class<?> cls = this.classLoader.loadClass(ContextRunner.class.getName());
			this.app = cls.newInstance();
			runContext(this.source,
					Collections.singletonMap(LiveBeansView.MBEAN_DOMAIN_PROPERTY_NAME,
							"function-invoker-" + UUID.randomUUID()),
					args);
		}
		catch (Exception e) {
			logger.error("Cannot deploy", e);
		}
		finally {
			ClassUtils.overrideThreadContextClassLoader(contextLoader);
		}
		RuntimeException e = getError();
		if (e != null) {
			throw e;
		}
	}

	public Object getBean(String name) {
		if (this.app != null) {
			if (containsBeanByName(name)) {
				return getBeanByName(name);
			}
			return getBeanByType(name);
		}
		return null;
	}

	private boolean containsBeanByName(String name) {
		Expression parsed = new SpelExpressionParser()
				.parseExpression("containsBean(\"" + name + "\")");
		Method method = ReflectionUtils.findMethod(this.app.getClass(), "getContext");
		return (Boolean) parsed.getValue(new StandardEvaluationContext(
				ReflectionUtils.invokeMethod(method, this.app)));
	}

	private Object getBeanByName(String name) {
		Expression parsed = new SpelExpressionParser()
				.parseExpression("getBean(\"" + name + "\")");
		Method method = ReflectionUtils.findMethod(this.app.getClass(), "getContext");
		return parsed.getValue(new StandardEvaluationContext(
				ReflectionUtils.invokeMethod(method, this.app)));
	}

	private Object getBeanByType(String name) {
		Expression parsed = new SpelExpressionParser()
				.parseExpression("getBean(T(" + name + "))");
		Method method = ReflectionUtils.findMethod(this.app.getClass(), "getContext");
		return parsed.getValue(new StandardEvaluationContext(
				ReflectionUtils.invokeMethod(method, this.app)));
	}

	public boolean containsBean(String name) {
		if (this.app != null) {
			if (containsBeanByName(name)) {
				return true;
			}
			Expression parsed = new SpelExpressionParser()
					.parseExpression("getBeansOfType(T(" + name + "))");
			Method method = ReflectionUtils.findMethod(this.app.getClass(), "getContext");
			@SuppressWarnings("unchecked")
			Map<String, Object> beans = (Map<String, Object>) parsed
					.getValue(new StandardEvaluationContext(
							ReflectionUtils.invokeMethod(method, this.app)));
			return !beans.isEmpty();
		}
		return false;
	}

	@PreDestroy
	public void close() {
		closeContext();
	}

	private RuntimeException getError() {
		if (this.app == null) {
			return null;
		}
		Method method = ReflectionUtils.findMethod(this.app.getClass(), "getError");
		Throwable e;
		e = (Throwable) ReflectionUtils.invokeMethod(method, this.app);
		if (e == null) {
			return null;
		}
		if (e instanceof RuntimeException) {
			return (RuntimeException) e;
		}
		return new IllegalStateException("Cannot launch", e);
	}

	private void runContext(String mainClass, Map<String, String> properties,
			String... args) {
		Method method = ReflectionUtils.findMethod(this.app.getClass(), "run",
				String.class, Map.class, String[].class);
		ReflectionUtils.invokeMethod(method, this.app, mainClass, properties, args);
	}

	private void closeContext() {
		if (this.app != null) {
			Method method = ReflectionUtils.findMethod(this.app.getClass(), "close");
			ReflectionUtils.invokeMethod(method, this.app);
		}
	}

}