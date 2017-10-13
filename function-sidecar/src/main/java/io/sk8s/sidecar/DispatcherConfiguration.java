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

package io.sk8s.sidecar;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Profile;

/**
 * @author Mark Fisher
 */
@Configuration
public class DispatcherConfiguration {

	@Profile("http")
	static class HttpDispatcherConfig {

		@Bean
		public HttpDispatcher dispatcher() {
			return new HttpDispatcher();
		}
	}

	@Profile("stdio")
	static class StdioDispatcherConfig {

		@Bean
		public StdioDispatcher dispatcher() {
			return new StdioDispatcher();
		}
	}
}
