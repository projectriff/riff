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

import io.grpc.Channel;
import io.grpc.ManagedChannelBuilder;
import io.sk8s.sidecar.grpc.function.StringFunctionGrpc;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.autoconfigure.condition.ConditionalOnMissingBean;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Conditional;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Profile;

import java.util.concurrent.TimeUnit;

/**
 * @author Mark Fisher
 * @author David Turanski
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

	@Profile("grpc")
	@EnableConfigurationProperties(GrpcProperties.class)
	static class GrpcDispatcherConfig {

		@Profile("grpc")
		@ConditionalOnProperty(value = "grpc.stub", havingValue = "blocking", matchIfMissing = true)
		static class BlockingStubConfiguration {
			@Bean
			public StringFunctionGrpc.StringFunctionBlockingStub functionStub(Channel grpcChannel) {
				return StringFunctionGrpc.newBlockingStub(grpcChannel);
			}

			@Bean
			public Dispatcher dispatcher(StringFunctionGrpc.StringFunctionBlockingStub stub) {
				return new GrpcBlockingDispatcher(stub);
			}
		}

		@Profile("grpc")
		@ConditionalOnProperty(value = "grpc.stub", havingValue = "async")
		static class AsyncStubConfiguration {
			@Bean
			public StringFunctionGrpc.StringFunctionStub functionStub(Channel grpcChannel) {
				return StringFunctionGrpc.newStub(grpcChannel);
			}

			@Bean
			public Dispatcher dispatcher(StringFunctionGrpc.StringFunctionStub stub) {
				return new GrpcAsyncDispatcher(stub);
			}
		}

		@Bean
		@ConditionalOnMissingBean(Channel.class)
		public Channel grpcChannel(GrpcProperties properties) {
			ManagedChannelBuilder<?> managedChannelBuilder = ManagedChannelBuilder
				.forAddress("localhost", properties.getPort()).usePlaintext(properties.isPlainText()).directExecutor();
			if (properties.getIdleTimeout() > 0) {
				managedChannelBuilder = managedChannelBuilder
					.idleTimeout(properties.getIdleTimeout(), TimeUnit.SECONDS);
			}
			if (properties.getMaxMessageSize() > 0) {
				managedChannelBuilder = managedChannelBuilder.maxInboundMessageSize(properties.getMaxMessageSize());
			}
			return managedChannelBuilder.build();
		}
	}
}
