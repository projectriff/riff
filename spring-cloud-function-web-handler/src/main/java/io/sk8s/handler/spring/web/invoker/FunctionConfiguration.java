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

package io.sk8s.handler.spring.web.invoker;

import java.io.IOException;
import java.io.UncheckedIOException;
import java.net.URL;
import java.net.URLClassLoader;
import java.util.Arrays;
import java.util.HashMap;
import java.util.Map;
import java.util.function.Consumer;
import java.util.function.Function;
import java.util.function.Supplier;

import org.springframework.beans.BeanUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.autoconfigure.condition.ConditionalOnMissingBean;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.cloud.deployer.resource.maven.MavenProperties;
import org.springframework.cloud.deployer.resource.maven.MavenResource;
import org.springframework.cloud.deployer.resource.maven.MavenResourceLoader;
import org.springframework.cloud.deployer.resource.support.DelegatingResourceLoader;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.core.io.ResourceLoader;
import org.springframework.util.ClassUtils;

/**
 * Sets up infrastructure capable of instantiating a "functional" bean (whether Supplier,
 * Function or Consumer) loaded dynamically according to {@link FunctionProperties}.
 *
 * <p>
 * In case of multiple class names, composes functions together.
 * </p>
 *
 * @author Eric Bottard
 * @author Mark Fisher
 */
@Configuration
@EnableConfigurationProperties(FunctionProperties.class)
public class FunctionConfiguration {

	@Autowired
	private ResourceLoader contextResourceLoader;

	@Autowired
	private DelegatingResourceLoader delegatingResourceLoader;

	@Autowired
	private FunctionProperties properties;

	@Bean
	@ConfigurationProperties("maven")
	public MavenProperties mavenProperties() {
		return new MavenProperties();
	}

	@Bean
	@ConditionalOnMissingBean(DelegatingResourceLoader.class)
	public DelegatingResourceLoader delegatingResourceLoader(MavenProperties mavenProperties) {
		Map<String, ResourceLoader> loaders = new HashMap<>();
		loaders.put(MavenResource.URI_SCHEME, new MavenResourceLoader(mavenProperties));
		return new DelegatingResourceLoader(loaders);
	}

	@Bean
	public Object function() throws Exception {
		URL[] urls = Arrays.stream(properties.getUri())
				.map(toResourceURL())
				.toArray(URL[]::new);

		try (URLClassLoader cl = new URLClassLoader(urls, /* explicit null parent */null)) {
			return Arrays.stream(properties.getClassName())
					.map(name -> (Object) BeanUtils.instantiateClass(ClassUtils.resolveClassName(name, cl)))
					.sequential()
					.reduce(this::compose).get();
		}
	}

	@SuppressWarnings("unchecked")
	private Object compose(Object a, Object b) {
		if (a instanceof Supplier && b instanceof Function) {
			Supplier supplier = (Supplier) a;
			Function function = (Function) b;
			return (Supplier) () -> function.apply(supplier.get());
		}
		else if (a instanceof Function && b instanceof Function) {
			Function function1 = (Function) a;
			Function function2 = (Function) b;
			return function1.andThen(function2);
		}
		else if (a instanceof Function && b instanceof Consumer) {
			Function function = (Function) a;
			Consumer consumer = (Consumer) b;
			return (Consumer) v -> consumer.accept(function.apply(v));
		}
		else {
			throw new IllegalArgumentException(
					String.format("Could not compose %s and %s", a.getClass(), b.getClass()));
		}
	}

	private Function<String, URL> toResourceURL() {
		return l -> {
			try {
				return delegatingResourceLoader.getResource(l).getFile().toURI().toURL();
			}
			catch (IOException e) {
				throw new UncheckedIOException(e);
			}
		};
	}

}
