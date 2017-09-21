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

import java.util.function.Consumer;
import java.util.function.Function;
import java.util.function.Supplier;

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
@SpringBootTest(classes = FunctionConfiguration.class)
public abstract class FunctionConfigurationTests {

	@TestPropertySource(properties = { "function.uri=file:target/test-classes",
			"function.className=" +
					"io.sk8s.handler.spring.web.invoker.FunctionConfigurationTests.NumberEmitter," +
					"io.sk8s.handler.spring.web.invoker.FunctionConfigurationTests.Frenchizer"
	})
	public static class SupplierCompositionTests extends FunctionConfigurationTests {

		@Autowired
		@Qualifier("function")
		private Supplier<String> function;

		@Test
		public void testComposition() {
			assertThat(function.get(), is("un"));
		}
	}

	@TestPropertySource(properties = { "function.uri=file:target/test-classes",
			"function.className=" +
					"io.sk8s.handler.spring.web.invoker.FunctionConfigurationTests.Doubler," +
					"io.sk8s.handler.spring.web.invoker.FunctionConfigurationTests.Frenchizer"
	})
	public static class FunctionCompositionTests extends FunctionConfigurationTests {

		@Autowired
		@Qualifier("function")
		private Function<Integer, String> function;

		@Test
		public void testComposition() {
			assertThat(function.apply(2), is("quatre"));
		}
	}

	@TestPropertySource(properties = { "function.uri=file:target/test-classes",
			"function.className=" +
					"io.sk8s.handler.spring.web.invoker.FunctionConfigurationTests.Frenchizer," +
					"io.sk8s.handler.spring.web.invoker.FunctionConfigurationTests.StaticFieldSetter"
	})
	public static class ConsumerCompositionTests extends FunctionConfigurationTests {

		@Rule
		public OutputCapture capture = new OutputCapture();

		@Autowired
		@Qualifier("function")
		private Consumer<Integer> function;

		@Test
		public void testComposition() {
			function.accept(2);
			capture.expect(containsString("Seen deux"));
		}
	}

	public static class Frenchizer implements Function<Integer, String> {

		@Override
		public String apply(Integer integer) {
			switch (integer) {
			case 1:
				return "un";
			case 2:
				return "deux";
			case 3:
				return "trois";
			case 4:
				return "quatre";
			default:
				throw new RuntimeException();
			}
		}
	}

	public static class Doubler implements Function<Integer, Integer> {
		@Override
		public Integer apply(Integer integer) {
			return 2 * integer;
		}
	}

	public static class NumberEmitter implements Supplier<Integer> {
		@Override
		public Integer get() {
			return 1;
		}
	}

	public static class StaticFieldSetter implements Consumer<Object> {

		@Override
		public void accept(Object o) {
			System.err.println("Seen " + o);
		}
	}

}
