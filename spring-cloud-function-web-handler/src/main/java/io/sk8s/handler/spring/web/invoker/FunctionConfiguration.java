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
import java.net.URL;
import java.net.URLClassLoader;
import java.util.ArrayList;
import java.util.List;
import java.util.function.Function;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.cloud.deployer.resource.maven.MavenProperties;
import org.springframework.cloud.deployer.resource.maven.MavenResource;
import org.springframework.cloud.deployer.resource.maven.MavenResourceLoader;
import org.springframework.cloud.function.support.FluxFunction;
import org.springframework.cloud.function.support.FunctionUtils;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.core.io.ResourceLoader;
import org.springframework.util.Assert;
import org.springframework.util.StringUtils;

import reactor.core.publisher.Flux;

/**
 * @author Mark Fisher
 */
@Configuration
@EnableConfigurationProperties(FunctionProperties.class)
public class FunctionConfiguration {

	@Autowired
	private ResourceLoader contextResourceLoader;

	@Autowired
	private FunctionProperties properties;

	@Bean
	@SuppressWarnings({ "unchecked", "rawtypes" })
	public Function<Flux<String>, Flux<String>> function() {
		URLClassLoader classLoader = null;
		List<URL> urls = new ArrayList<>();
		try {
			String[] locations = StringUtils.tokenizeToStringArray(this.properties.getUri(), ",");
			for (String location : locations) {
				if (location.startsWith("maven:")) {
					MavenResourceLoader mavenResourceLoader = new MavenResourceLoader(new MavenProperties());
					MavenResource resource = (MavenResource) mavenResourceLoader.getResource(location);
					urls.add(new URL("file:" + resource.getFile().getAbsolutePath()));
				}
				else {
					urls.add(this.contextResourceLoader.getResource(location).getURL());
				}
			}
			// TODO: add a 'dependencies' property for additional JARs
			classLoader = new URLClassLoader(urls.toArray(new URL[urls.size()]), this.getClass().getClassLoader());
			String[] classnames = StringUtils.tokenizeToStringArray(this.properties.getClassname(), ",");
			Assert.isTrue(classnames.length == locations.length, "locations and classnames must have the same number of values");
			Function function = null;
			for (int i = 0; i < classnames.length; i++) {
				Class<?> clazz = classLoader.loadClass(classnames[i]);
				Function currentFunction = (Function) clazz.newInstance();
				if (i == 0) {
					function = currentFunction;
				}
				else {
					function = function.andThen(currentFunction);
				}
			}
			return (!FunctionUtils.isFluxFunction(function))
					? new FluxFunction<String, String>((Function<String, String>) function)
					: (Function<Flux<String>, Flux<String>>) function;
		}
		catch (Exception e) {
			throw new IllegalArgumentException("failed to load function from resource", e);
		}
		finally {
			if (classLoader != null) {
				try {
					classLoader.close();
				}
				catch (IOException e) {
					// ignore
				}
			}
		}
	}
}
