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

import java.io.IOException;
import java.io.UncheckedIOException;
import java.net.URL;
import java.net.URLClassLoader;
import java.util.Arrays;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.function.Function;

import javax.annotation.PostConstruct;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.autoconfigure.condition.ConditionalOnMissingBean;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.cloud.deployer.resource.maven.MavenProperties;
import org.springframework.cloud.deployer.resource.maven.MavenResource;
import org.springframework.cloud.deployer.resource.maven.MavenResourceLoader;
import org.springframework.cloud.deployer.resource.support.DelegatingResourceLoader;
import org.springframework.cloud.function.context.FunctionRegistration;
import org.springframework.cloud.function.context.FunctionRegistry;
import org.springframework.context.ConfigurableApplicationContext;
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
 * @author Dave Syer
 */
@Configuration
@EnableConfigurationProperties
public class FunctionConfiguration {

	@Autowired
	private FunctionRegistry registry;

	@Autowired
	private FunctionProperties properties;

	@Autowired
	private DelegatingResourceLoader delegatingResourceLoader;

	@Autowired
	private ConfigurableApplicationContext context;

	@Bean
	@ConfigurationProperties("maven")
	public MavenProperties mavenProperties() {
		return new MavenProperties();
	}

	@Bean
	@ConditionalOnMissingBean(DelegatingResourceLoader.class)
	public DelegatingResourceLoader delegatingResourceLoader(
			MavenProperties mavenProperties) {
		Map<String, ResourceLoader> loaders = new HashMap<>();
		loaders.put(MavenResource.URI_SCHEME, new MavenResourceLoader(mavenProperties));
		return new DelegatingResourceLoader(loaders);
	}

	/**
	 * Registers a singleton bean for each of the function classes passed into the
	 * {@link FunctionProperties}. They are named sequentially "function0", "function1",
	 * etc. The instances are created in an isolated class loader, so the jar they are
	 * packed in has to define all the dependencies (except core JDK).
	 * 
	 * TODO: Add reactor and maybe Message (not spring-messaging) to class loader in a way
	 * that they can be shared with the main application context.
	 */
	@PostConstruct
	public void init() {
		URL[] urls = Arrays.stream(properties.getJarLocation())
				.map(toResourceURL(delegatingResourceLoader)).toArray(URL[]::new);

		try {
			BeanCreator creator = new BeanCreator(urls);
			Arrays.stream(properties.getClassName()).map(creator::create).sequential()
					.forEach(creator::register);
		}
		catch (Exception e) {
			throw new IllegalStateException("Cannot create functions", e);
		}
	}

	private Function<String, URL> toResourceURL(DelegatingResourceLoader resourceLoader) {
		return l -> {
			try {
				return resourceLoader.getResource(l).getFile().toURI().toURL();
			}
			catch (IOException e) {
				throw new UncheckedIOException(e);
			}
		};
	}

	private class BeanCreator {

		private URLClassLoader cl;
		private AtomicInteger counter = new AtomicInteger(0);

		public BeanCreator(URL[] urls) {
			this.cl = new URLClassLoader(urls, /* explicit null parent */null);
		}

		public Object create(String type) {
			return context.getAutowireCapableBeanFactory()
					.createBean(ClassUtils.resolveClassName(type, cl));
		}

		public void register(Object bean) {
			registry.register(new FunctionRegistration<Object>(bean)
					.names("function" + counter.getAndIncrement()));
		}

	}

}
