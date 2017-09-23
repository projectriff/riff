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

import java.util.Arrays;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.http.MediaType;
import org.springframework.http.converter.HttpMessageConverter;
import org.springframework.http.converter.StringHttpMessageConverter;

/**
 * Configures the server related beans of the application.
 *
 * @author Eric Bottard
 */
@Configuration
public class SpringWebHandlerConfiguration {

	/**
	 * Installs a message converter that supports {@literal text/plain}
	 * if the target type is assignable from {@literal String}.
	 *
	 * <p>
	 *     This departs from the standard converter that is automatically registered, which
	 *     requires that the target type is exactly String (but accepts all content types).
	 *     This is necessary here because the signature of {@link SpringWebHandlerController#invoke(Object)}
	 *     accepts {@literal Object}.
	 * </p>
	 */
	@Bean
	public HttpMessageConverter<?> textPlainMessageConverter() {
		StringHttpMessageConverter converter = new StringHttpMessageConverter() {
			@Override
			public boolean supports(Class<?> clazz) {
				return clazz.isAssignableFrom(String.class);
			}
		};
		converter.setSupportedMediaTypes(Arrays.asList(MediaType.TEXT_PLAIN));
		return converter;
	}
}
