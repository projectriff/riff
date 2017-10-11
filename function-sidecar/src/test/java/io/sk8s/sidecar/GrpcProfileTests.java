/*
 * Copyright 2017 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package io.sk8s.sidecar;

import static org.assertj.core.api.Java6Assertions.assertThat;

import io.grpc.Channel;
import io.grpc.ManagedChannel;
import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.inprocess.InProcessChannelBuilder;
import io.grpc.inprocess.InProcessServerBuilder;
import io.sk8s.sidecar.grpc.fntypes.FunctionTypeProtos;
import io.sk8s.sidecar.grpc.function.StringFunctionGrpc;
import org.junit.AfterClass;
import org.junit.BeforeClass;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Import;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.test.context.junit4.SpringRunner;

import java.io.IOException;
import java.util.UUID;

/**
 * @author David Turanski
 **/
@SpringBootTest(classes = GrpcProfileTests.TestConfiguration.class, webEnvironment = SpringBootTest.WebEnvironment.NONE)
@RunWith(SpringRunner.class)
@ActiveProfiles("grpc")
public class GrpcProfileTests {
	private static FunctionServer server;
	private static ManagedChannel inProcessChannel;
	private static String serverName = UUID.randomUUID().toString();

	@BeforeClass
	public static void setUp() throws Exception {

		server = new FunctionServer(InProcessServerBuilder.forName(serverName).directExecutor());
		server.start();
		inProcessChannel = InProcessChannelBuilder.forName(serverName).directExecutor().build();
	}

	@AfterClass
	public static void tearDown() throws Exception {
		inProcessChannel.shutdownNow();
		server.stop();
	}

	@Autowired
	GrpcBlockingDispatcher dispatcher;

	@Test
	public void contextLoads() {

	}

	@Test
	public void grpcToUpper() {
		try {
			assertThat(dispatcher.dispatch("hello")).isEqualTo("HELLO");
		}
		catch (Exception e) {
			e.printStackTrace();
		}

	}

	@Configuration
	@Import(DispatcherConfiguration.class)
	static class TestConfiguration {

		@Bean
		public Channel grpcChannel() {
			return inProcessChannel;
		}
	}

	static class FunctionServer {

		private final Server server;

		public FunctionServer(ServerBuilder<?> serverBuilder) {

			server = serverBuilder.addService(new FunctionService()).build();
		}

		/**
		 * Start serving requests.
		 */
		public void start() throws IOException {
			server.start();

			Runtime.getRuntime().addShutdownHook(new Thread() {
				@Override
				public void run() {
					// Use stderr here since the logger may has been reset by its JVM shutdown hook.
					System.err.println("*** shutting down gRPC server since JVM is shutting down");
					FunctionServer.this.stop();
					System.err.println("*** server shut down");
				}
			});
		}

		/**
		 * Stop serving requests and shutdown resources.
		 */
		public void stop() {
			if (server != null) {
				server.shutdown();
			}
		}

		static class FunctionService extends StringFunctionGrpc.StringFunctionImplBase {
			@Override
			public void call(io.sk8s.sidecar.grpc.fntypes.FunctionTypeProtos.Request request,
				io.grpc.stub.StreamObserver<io.sk8s.sidecar.grpc.fntypes.FunctionTypeProtos.Reply> observer) {
				FunctionTypeProtos.Reply reply = FunctionTypeProtos.Reply.newBuilder()
					.setBody(request.getBody().toUpperCase()).build();
				observer.onNext(reply);
				observer.onCompleted();
			}
		}

	}
}
