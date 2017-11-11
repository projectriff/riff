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

import org.springframework.beans.BeanUtils;
import org.springframework.beans.BeansException;
import org.springframework.beans.factory.config.ConfigurableListableBeanFactory;
import org.springframework.beans.factory.support.BeanDefinitionRegistry;
import org.springframework.beans.factory.support.BeanDefinitionRegistryPostProcessor;
import org.springframework.boot.autoconfigure.condition.ConditionalOnMissingBean;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.cloud.deployer.resource.maven.MavenProperties;
import org.springframework.cloud.deployer.resource.maven.MavenResource;
import org.springframework.cloud.deployer.resource.maven.MavenResourceLoader;
import org.springframework.cloud.deployer.resource.support.DelegatingResourceLoader;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.core.annotation.Order;
import org.springframework.core.env.Environment;
import org.springframework.core.io.ResourceLoader;
import org.springframework.stereotype.Component;
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
public class FunctionConfiguration {

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
	@Component
	@Order(0)
	protected static class ContextFunctionPostProcessor
			implements BeanDefinitionRegistryPostProcessor {

		private Function<String, URL> toResourceURL(
				DelegatingResourceLoader resourceLoader) {
			return l -> {
				try {
					return resourceLoader.getResource(l).getFile().toURI().toURL();
				}
				catch (IOException e) {
					throw new UncheckedIOException(e);
				}
			};
		}

		@Override
		public void postProcessBeanFactory(ConfigurableListableBeanFactory beanFactory)
				throws BeansException {
			// In the BFPP so these things are available but not @Autowirable
			Environment environment = beanFactory.getBean("environment",
					Environment.class);
			FunctionProperties properties = beanFactory.getBean(FunctionProperties.class);
			// Spring Boot hasn't done the binding yet, but we're only interested in one
			// property. TODO: relaxed binding
			properties.setUri(environment.getProperty("function.uri"));
			properties.init();
			DelegatingResourceLoader delegatingResourceLoader = beanFactory
					.getBean(DelegatingResourceLoader.class);

			URL[] urls = Arrays.stream(properties.getJarLocation())
					.map(toResourceURL(delegatingResourceLoader)).toArray(URL[]::new);

			try (URLClassLoader cl = new URLClassLoader(urls,
					/* explicit null parent */null)) {
				AtomicInteger counter = new AtomicInteger(0);
				Arrays.stream(properties.getClassName())
						.map(name -> (Object) BeanUtils
								.instantiateClass(ClassUtils.resolveClassName(name, cl)))
						.sequential().forEach(bean -> beanFactory.registerSingleton(
								"function" + counter.getAndIncrement(), bean));
			}
			catch (IOException e) {

			}
		}

		@Override
		public void postProcessBeanDefinitionRegistry(BeanDefinitionRegistry registry)
				throws BeansException {
		}
	}

}
