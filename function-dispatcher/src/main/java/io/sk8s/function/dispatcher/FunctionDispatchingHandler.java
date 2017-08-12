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
import java.util.HashMap;
import java.util.Map;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

import org.springframework.beans.BeansException;
import org.springframework.cloud.stream.binder.Binder;
import org.springframework.cloud.stream.binder.BinderFactory;
import org.springframework.cloud.stream.binder.Binding;
import org.springframework.cloud.stream.binder.ExtendedConsumerProperties;
import org.springframework.cloud.stream.binder.ExtendedProducerProperties;
import org.springframework.cloud.stream.binder.HeaderMode;
import org.springframework.cloud.stream.binder.kafka.properties.KafkaConsumerProperties;
import org.springframework.cloud.stream.binder.kafka.properties.KafkaProducerProperties;
import org.springframework.context.ApplicationContext;
import org.springframework.context.ApplicationContextAware;
import org.springframework.integration.channel.DirectChannel;
import org.springframework.integration.support.MessageBuilder;
import org.springframework.messaging.MessageChannel;
import org.springframework.messaging.MessageHandler;
import org.springframework.messaging.MessageHeaders;
import org.springframework.util.StringUtils;

import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.utils.HttpClientUtils;

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
public class FunctionDispatchingHandler implements FunctionResourceHandler, ApplicationContextAware {

	private static Log logger = LogFactory.getLog(FunctionDispatchingHandler.class);

	private final KubernetesClient kubernetesClient;

	private final Map<String, FunctionResource> resources = new HashMap<>();

	private final Map<String, Dispatcher> dispatchers = new HashMap<>();

	private final Map<String, Binding<?>> bindings = new HashMap<>();

	private final Binder<MessageChannel, ExtendedConsumerProperties<KafkaConsumerProperties>, ExtendedProducerProperties<KafkaProducerProperties>> binder;

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
			String dispatcherName = handlerResource.getSpec().getDispatcher();
			Dispatcher dispatcher = this.getDispatcher(dispatcherName);
			if (dispatcher != null) {
				dispatcher.dispatch(payload, m.getHeaders(), functionResource, handlerResource);
			}
			else {
				logger.info("unknown dispatcher: " + dispatcher);
			}
		}
		catch (UnsupportedEncodingException e) {
			e.printStackTrace();
		}
	};

	@SuppressWarnings("unchecked")
	public FunctionDispatchingHandler(KubernetesClient kubernetesClient, BinderFactory binderFactory) {
		this.kubernetesClient = kubernetesClient;
		this.binder = (Binder<MessageChannel, ExtendedConsumerProperties<KafkaConsumerProperties>, ExtendedProducerProperties<KafkaProducerProperties>>)
				binderFactory.getBinder("kafka", MessageChannel.class);
	}

	@Override
	public void setApplicationContext(ApplicationContext context) throws BeansException {
		for (Map.Entry<String, Dispatcher> entry : context.getBeansOfType(Dispatcher.class).entrySet()) {
			this.dispatchers.put(entry.getKey().toLowerCase(), entry.getValue());
		}
	}

	private Dispatcher getDispatcher(String name) {
		return this.dispatchers.get(name.toLowerCase());
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
}
