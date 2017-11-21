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

import java.net.URI;

import org.junit.After;
import org.junit.Before;
import org.junit.Rule;
import org.junit.Test;
import org.junit.rules.ExpectedException;

import org.springframework.beans.factory.BeanCreationException;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.test.web.client.TestRestTemplate;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.RequestEntity;
import org.springframework.http.ResponseEntity;
import org.springframework.util.SocketUtils;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * @author Dave Syer
 *
 */
public class IsolatedTests {

	private TestRestTemplate rest;
	private int port = SocketUtils.findAvailableTcpPort();
	private ApplicationRunner runner;

	@Rule
	public ExpectedException expected = ExpectedException.none();

	@Before
	public void init() {
		runner = new JavaFunctionInvokerApplication().runner();
		rest = new TestRestTemplate();
	}

	@After
	public void close() {
		if (runner != null) {
			runner.close();
		}
	}

	@Test
	public void fluxFunctionNotIsolated() throws Exception {
		expected.expect(BeanCreationException.class);
		SpringApplication.run(JavaFunctionInvokerApplication.class,
				"--server.port=" + port, "--function.uri=file:target/test-classes"
						+ "?handler=io.sk8s.invoker.java.function.FluxDoubler");
	}

	@Test
	public void fluxFunction() throws Exception {
		runner.run("--server.port=" + port, "--function.uri=file:target/test-classes"
				+ "?handler=io.sk8s.invoker.java.function.FluxDoubler");
		ResponseEntity<String> result = rest
				.exchange(
						RequestEntity.post(new URI("http://localhost:" + port + "/"))
								.contentType(MediaType.TEXT_PLAIN).body("5"),
						String.class);
		assertThat(result.getStatusCode()).isEqualTo(HttpStatus.OK);
		// Check single valued response in s-c-f
		assertThat(result.getBody()).isEqualTo("[10]");
	}

	@Test
	public void simpleFunction() throws Exception {
		runner.run("--server.port=" + port, "--function.uri=file:target/test-classes"
				+ "?handler=io.sk8s.invoker.java.function.Doubler");
		ResponseEntity<String> result = rest
				.exchange(
						RequestEntity.post(new URI("http://localhost:" + port + "/"))
								.contentType(MediaType.TEXT_PLAIN).body("5"),
						String.class);
		assertThat(result.getStatusCode()).isEqualTo(HttpStatus.OK);
		assertThat(result.getBody()).isEqualTo("10");
	}

	@Test
	public void appClassPath() throws Exception {
		runner.run("--server.port=" + port,
				"--function.uri=app:classpath?handler=io.sk8s.invoker.java.function.SpringDoubler");
		ResponseEntity<String> result = rest
				.exchange(
						RequestEntity.post(new URI("http://localhost:" + port + "/"))
								.contentType(MediaType.TEXT_PLAIN).body("5"),
						String.class);
		assertThat(result.getStatusCode()).isEqualTo(HttpStatus.OK);
		assertThat(result.getBody()).isEqualTo("10");
	}
}
