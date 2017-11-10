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

import java.util.stream.IntStream;

import javax.annotation.PostConstruct;
import javax.validation.constraints.NotNull;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.util.Assert;
import org.springframework.util.StringUtils;
import org.springframework.validation.annotation.Validated;

/**
 * @author Mark Fisher
 * @author Dave Syer
 */
@ConfigurationProperties("function")
@Validated
public class FunctionProperties {

	private static Log logger = LogFactory.getLog(FunctionProperties.class);

	@NotNull
	private String uri;
	private String[] jarLocation;
	private String[] className;

	private String functionName;

	public String getUri() {
		return uri;
	}

	public void setUri(String uri) {
		this.uri = uri;
	}

	public String[] getJarLocation() {
		return jarLocation;
	}

	public String[] getClassName() {
		return className;
	}

	@PostConstruct
	public void init() {
		if (uri != null) {
			logger.info("initializing with uri: " + uri);
			String[] tokens = uri.split("\\?");
			Assert.isTrue(tokens.length == 2,
					"expected format: [jarLocation]?[className]");
			String jarLocation = tokens[0];
			String className = tokens[1];
			this.jarLocation = StringUtils.commaDelimitedListToStringArray(jarLocation);
			this.className = StringUtils.commaDelimitedListToStringArray(className);
		}
		this.functionName = StringUtils
				.arrayToCommaDelimitedString(IntStream.range(0, this.className.length)
						.sequential().mapToObj(i -> "function" + i).toArray());
	}

	public String getFunctionName() {
		return this.functionName;
	}
}
