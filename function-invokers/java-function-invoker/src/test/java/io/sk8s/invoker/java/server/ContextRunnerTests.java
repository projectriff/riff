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

import java.util.Collections;

import io.sk8s.invoker.java.function.Doubler;

import org.junit.Test;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * @author Dave Syer
 *
 */
public class ContextRunnerTests {

	@Test
	public void startEvaluateAndStop() {
		ContextRunner runner = new ContextRunner();
		runner.run(Doubler.class.getName(), Collections.emptyMap(), "--spring.main.webEnvironment=false");
		assertThat(runner.getContext()).isNotNull();
		runner.close();
	}

}
