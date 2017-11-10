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

package io.sk8s.invoker.java.function;

import java.util.function.Consumer;
import java.util.function.Function;
import java.util.function.Supplier;

import io.sk8s.invoker.java.server.FunctionConfiguration;
import io.sk8s.invoker.java.server.FunctionProperties;

import org.junit.Rule;
import org.junit.Test;
import org.junit.runner.RunWith;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.rule.OutputCapture;
import org.springframework.test.context.TestPropertySource;
import org.springframework.test.context.junit4.SpringRunner;

import static org.hamcrest.Matchers.containsString;
import static org.hamcrest.Matchers.is;
import static org.junit.Assert.assertThat;

@RunWith(SpringRunner.class)
@SpringBootTest(classes = { FunctionConfiguration.class, FunctionProperties.class })
public abstract class FunctionConfigurationTests {

	@TestPropertySource(properties = { "function.uri=file:target/test-classes?io.sk8s.invoker.java.function.Doubler"
	})
	public static class SingleFunctionTests extends FunctionConfigurationTests {
	
		@Autowired
		@Qualifier("function0")
		private Function<Integer, Integer> function;
	
		@Test
		public void testDouble() {
			assertThat(function.apply(2), is(4));
		}
	}
	
	@TestPropertySource(properties = { "function.uri=file:target/test-classes?" +
					"io.sk8s.invoker.java.function.NumberEmitter"
	})
	public static class SingleSupplierTests extends FunctionConfigurationTests {

		@Autowired
		@Qualifier("function0")
		private Supplier<Integer> function;

		@Test
		public void testComposition() {
			assertThat(function.get(), is(1));
		}
	}

	@TestPropertySource(properties = { "function.uri=file:target/test-classes?" +
					"io.sk8s.invoker.java.function.Printer"
	})
	public static class SingleConsumerTests extends FunctionConfigurationTests {

		@Rule
		public OutputCapture capture = new OutputCapture();

		@Autowired
		@Qualifier("function0")
		private Consumer<Object> function;

		@Test
		public void testComposition() {
			function.accept("deux");
			capture.expect(containsString("Seen deux"));
		}
	}

}
