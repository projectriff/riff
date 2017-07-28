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

package io.sk8s.function.runner;

import java.util.function.Function;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.cloud.function.compiler.FunctionCompiler;
import org.springframework.cloud.function.compiler.proxy.LambdaCompilingFunction;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.core.io.ByteArrayResource;
import org.springframework.core.io.Resource;

/**
 * @author Mark Fisher
 */
@Configuration
@EnableConfigurationProperties(FunctionRunnerProperties.class)
public class FunctionRunnerConfiguration {

	@Autowired
	private FunctionRunnerProperties properties;

	@Bean
	FunctionCompiler<String, String> functionCompiler() {
		return new FunctionCompiler<>();
	}

	@Bean
	Function<String, String> compiledFunction(FunctionCompiler<String, String> functionCompiler) throws Exception {
		Resource resource = new ByteArrayResource(properties.getLambda().getBytes());
		LambdaCompilingFunction<String, String> function = new LambdaCompilingFunction<>(resource, functionCompiler);
		function.setTypeParameterizations("String, String");
		return function;
	}
}
