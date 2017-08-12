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

package io.sk8s.function.dispatcher;

import java.io.UnsupportedEncodingException;
import java.net.SocketTimeoutException;
import java.net.URL;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.cloud.stream.binder.Binder;
import org.springframework.cloud.stream.binder.BinderFactory;
import org.springframework.cloud.stream.binder.Binding;
import org.springframework.cloud.stream.binder.ExtendedConsumerProperties;
import org.springframework.cloud.stream.binder.ExtendedProducerProperties;
import org.springframework.cloud.stream.binder.HeaderMode;
import org.springframework.cloud.stream.binder.kafka.properties.KafkaConsumerProperties;
import org.springframework.cloud.stream.binder.kafka.properties.KafkaProducerProperties;
import org.springframework.http.ResponseEntity;
import org.springframework.integration.channel.DirectChannel;
import org.springframework.integration.support.MessageBuilder;
import org.springframework.messaging.MessageChannel;
import org.springframework.messaging.MessageHandler;
import org.springframework.messaging.MessageHeaders;
import org.springframework.util.Assert;
import org.springframework.util.StringUtils;
import org.springframework.web.client.RestTemplate;

import io.fabric8.kubernetes.api.model.ContainerBuilder;
import io.fabric8.kubernetes.api.model.EnvVar;
import io.fabric8.kubernetes.api.model.EnvVarBuilder;
import io.fabric8.kubernetes.api.model.JobSpecBuilder;
import io.fabric8.kubernetes.api.model.VolumeBuilder;
import io.fabric8.kubernetes.api.model.VolumeMountBuilder;
import io.fabric8.kubernetes.client.DefaultKubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.utils.HttpClientUtils;

import io.sk8s.core.function.Env;
import io.sk8s.core.function.FunctionResource;
import io.sk8s.core.function.FunctionResourceHandler;
import io.sk8s.core.handler.HandlerResource;

import com.fasterxml.jackson.databind.ObjectMapper;

import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;
import okio.BufferedSource;

/**
 * @author Mark Fisher
 */
public class FunctionDispatchingHandler implements FunctionResourceHandler {

	private static Log logger = LogFactory.getLog(FunctionDispatchingHandler.class);

	@Autowired
	private FunctionDispatcherProperties properties;

	private final Map<String, FunctionResource> resources = new HashMap<>();

	private final Map<String, Binding<?>> bindings = new HashMap<>();

	private final Binder<MessageChannel, ExtendedConsumerProperties<KafkaConsumerProperties>, ExtendedProducerProperties<KafkaProducerProperties>> binder;

	private final KubernetesClient kubernetesClient = new DefaultKubernetesClient();

	private final RestTemplate restTemplate = new RestTemplate();

	private final ObjectMapper objectMapper = new ObjectMapper();

	private final MessageHandler messageHandler = m -> {
		try {
			String payload = new String((byte[]) m.getPayload(), "UTF-8");
			String functionName = m.getHeaders().get("function", String.class);
			FunctionResource functionResource = this.resources.get(functionName);
			if (functionResource == null) {
				logger.error("unknown function: " + functionName);
			}
			String handlerName = functionResource.getSpec().getHandler();
			HandlerResource handlerResource = getHandler(handlerName);
			if (handlerResource == null) {
				logger.error("unknown handler: " + handlerName);
			}
			// TODO: extract Dispatcher as a pluggable strategy
			String dispatcher = handlerResource.getSpec().getDispatcher();
			if ("JobLauncher".equalsIgnoreCase(dispatcher)) {
				launchJob(functionResource, handlerResource, payload);
			}
			else if ("ServiceInvoker".equalsIgnoreCase(dispatcher)) {
				invokeService(functionResource, payload);
			}
			else {
				logger.info("unknown dispatcher: " + dispatcher);
			}
		}
		catch (UnsupportedEncodingException e) {
			e.printStackTrace();
		}
	};

	private List<String> resolvePlaceholders(List<String> original, FunctionResource functionResource) {
		// TODO: apply to entire resource, for now just args
		List<String> resolved = new ArrayList<>(original.size());
		for (int i = 0; i < original.size(); i++) {
			String s = original.get(i);
			// TODO: find the name with pattern, for now just "command"
			if (s.contains("$COMMAND")) {
				resolved.add(functionResource.getSpec().getParam("command"));
			}
			else {
				resolved.add(s);
			}
		}
		return resolved;
	}

	@SuppressWarnings("unchecked")
	public FunctionDispatchingHandler(BinderFactory binderFactory) {
		this.binder = (Binder<MessageChannel, ExtendedConsumerProperties<KafkaConsumerProperties>, ExtendedProducerProperties<KafkaProducerProperties>>)
				binderFactory.getBinder("kafka", MessageChannel.class);
	}

	@Override
	public void resourceAdded(FunctionResource resource) {
		String functionName = resource.getMetadata().get("name");
		this.resources.put(functionName, resource);
		addListener(resource);
		logger.info("function added: " + functionName);
	}

	private void addListener(FunctionResource resource) {
		String destination = resource.getSpec().getTopic();
		final String functionName = resource.getMetadata().get("name");
		DirectChannel channel = new DirectChannel();
		channel.subscribe(m -> {
			messageHandler.handleMessage(MessageBuilder.fromMessage(m)
					.setHeader("function", functionName)
					.setHeader(MessageHeaders.CONTENT_TYPE, "text/plain")
					.build());
		});
		Binding<?> binding = this.binder.bindConsumer(destination, functionName, channel, consumerProperties());
		this.bindings.put(functionName, binding);
	}

	private ExtendedConsumerProperties<KafkaConsumerProperties> consumerProperties() {
		KafkaConsumerProperties kafkaProps = new KafkaConsumerProperties();
		ExtendedConsumerProperties<KafkaConsumerProperties> extendedProps = new ExtendedConsumerProperties<>(kafkaProps);
		extendedProps.setHeaderMode(HeaderMode.raw);
		return extendedProps;
	}

	@Override
	public void resourceDeleted(FunctionResource resource) {
		String functionName = resource.getMetadata().get("name");
		Binding<?> binding = this.bindings.remove(functionName);
		if (binding != null) {
			binding.unbind();
		}
		this.resources.remove(functionName);
		logger.info("function deleted: " + functionName);
	}

	private HandlerResource getHandler(String name) {
		// TODO: use this.handlers (local cache from a watch)
		OkHttpClient httpClient = HttpClientUtils.createHttpClient(kubernetesClient.getConfiguration());
		Response response = null;
		try {
			// TODO: replace this code
			URL url = new URL(kubernetesClient.getMasterUrl() + "apis/extensions.sk8s.io/v1/namespaces/default/handlers/" + name);
			Request.Builder requestBuilder = new Request.Builder().get().url(url);
			response = httpClient.newCall(requestBuilder.build()).execute();
			BufferedSource source = response.body().source();
			while (!source.exhausted()) {
				String line = source.readUtf8LineStrict();
				if (!StringUtils.hasText(line)) {
					break;
				}
				try {
					return this.objectMapper.readValue(line, HandlerResource.class);
				}
				catch (Exception e) {
					e.printStackTrace();
					break;
				}
			}
		}
		catch (SocketTimeoutException e) {
			// reconnect
		}
		catch (Exception e) {
			e.printStackTrace();
		}
		finally {
			try {
				response.close();
			}
			catch (Exception e) {
				// ignore
			}
		}
		return null;
	}

	private void invokeService(FunctionResource functionResource, String payload) {
		// TODO: create the Handler instance by passing params
		String url = functionResource.getSpec().getParam("url");
		Assert.hasText(url, "no url provided for function " + functionResource.getMetadata().get("name"));
		ResponseEntity<?> response = this.restTemplate.postForEntity(url, payload, String.class);
		System.out.println("response: "+ response);
	}

	private void launchJob(FunctionResource functionResource, HandlerResource handlerResource, String payload) {
		String functionName = functionResource.getMetadata().get("name");
		String s = this.kubernetesClient.extensions().jobs().inNamespace(this.properties.getNamespace()).createNew()
			.withApiVersion("batch/v1")
			.withNewMetadata()
				.withName(functionName + "-" + System.currentTimeMillis())
			.endMetadata()
			.withSpec(new JobSpecBuilder()
				.withNewTemplate()
					.withNewMetadata()
						.withLabels(Collections.singletonMap("function", functionName))
					.endMetadata()
					.withNewSpec()
						.withRestartPolicy("OnFailure")
						.withActiveDeadlineSeconds(10L)
						.withContainers(new ContainerBuilder()
							.withName("main")
							.withImage(handlerResource.getSpec().getImage())
							.withCommand(handlerResource.getSpec().getCommand())
							.withArgs(this.resolvePlaceholders(handlerResource.getSpec().getArgs(), functionResource))
							.withEnv(buildEnvVars(functionResource.getSpec().getEnv(), payload))
							.withVolumeMounts(new VolumeMountBuilder()
								.withMountPath("/output")
								.withName("messages")
							.build())
						.build())
						.withVolumes(new VolumeBuilder()
							.withName("messages")
							.withNewHostPath("/messages")
						.build())
					.endSpec()
				.endTemplate()
			.build())
			.done().toString();
		System.out.println("JOB: " + s);
	}

	private EnvVar[] buildEnvVars(List<Env> envList, String payload) {
		Map<String, String> map = new HashMap<>();
		if (envList != null) {
			for (Env env : envList) {
				String value = ("payload".equalsIgnoreCase(env.getValueFrom())) ? payload : "";
				map.put(env.getName(), value);
			}
		}
		map.put("MESSAGE", payload);
		EnvVar[] vars = new EnvVar[map.size()];
		int i = 0;
		for (Map.Entry<String, String> entry : map.entrySet()) {
			vars[i++] = new EnvVarBuilder()
					.withName(entry.getKey())
					.withValue(entry.getValue())
					.build();
		}
		return vars;
	}
}
