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

package io.sk8s.event.dispatcher;

import java.io.UnsupportedEncodingException;
import java.util.HashMap;
import java.util.Map;

import io.sk8s.core.resource.ResourceAddedEvent;
import io.sk8s.core.resource.ResourceAddedOrModifiedEvent;
import io.sk8s.core.resource.ResourceDeletedEvent;
import io.sk8s.kubernetes.api.model.FunctionUtils;
import io.sk8s.kubernetes.api.model.Handler;
import io.sk8s.kubernetes.api.model.HandlerReference;
import io.sk8s.kubernetes.api.model.HandlerReferenceBuilder;
import io.sk8s.kubernetes.api.model.HandlerUtils;
import io.sk8s.kubernetes.api.model.XFunction;
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
import org.springframework.context.event.EventListener;
import org.springframework.integration.channel.DirectChannel;
import org.springframework.integration.support.MessageBuilder;
import org.springframework.messaging.MessageChannel;
import org.springframework.messaging.MessageHandler;
import org.springframework.messaging.MessageHeaders;

/**
 * @author Mark Fisher
 */
public class EventDispatchingHandler implements ApplicationContextAware {

	private final Log logger = LogFactory.getLog(EventDispatchingHandler.class);

	// TODO: Change key to ObjectReference or similar
	private final Map<String, XFunction> functions = new HashMap<>();

	private final Map<HandlerReference, Handler> handlers = new HashMap<>();

	private final Map<String, Dispatcher> dispatchers = new HashMap<>();

	// TODO: Change key to ObjectReference or similar
	private final Map<String, Binding<?>> bindings = new HashMap<>();

	private final Binder<MessageChannel, ExtendedConsumerProperties<KafkaConsumerProperties>, ExtendedProducerProperties<KafkaProducerProperties>> binder;

	private final MessageHandler messageHandler = m -> {
		try {
			String payload = new String((byte[]) m.getPayload(), "UTF-8");
			String functionName = m.getHeaders().get("function", String.class);
			XFunction functionResource = this.functions.get(functionName);
			if (functionResource == null) {
				logger.error("unknown function: " + functionName);
				return; // TODO is that the best thing to do?
			}
			HandlerReference handlerRef = FunctionUtils.adjustHandlerReference(functionResource);
			Handler handlerResource = handlers.get(handlerRef);
			if (handlerResource == null) {
				logger.error("unknown handler: " + handlerRef);
			}
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
	public EventDispatchingHandler(BinderFactory binderFactory) {
		this.binder = (Binder<MessageChannel,
				ExtendedConsumerProperties<KafkaConsumerProperties>,
				ExtendedProducerProperties<KafkaProducerProperties>>)
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


	@EventListener
	public void onHandlerRegistered(ResourceAddedOrModifiedEvent<Handler> event) {
		HandlerReference reference = HandlerUtils.refOf(event.getResource());
		handlers.put(reference, event.getResource());
	}

	@EventListener
	public void onHandlerUnregistered(ResourceDeletedEvent<Handler> event) {
		HandlerReference reference = HandlerUtils.refOf(event.getResource());
		handlers.remove(reference);
	}

	@EventListener
	public void onFunctionRegistered(ResourceAddedEvent<XFunction> event) {
		XFunction functionResource = event.getResource();
		String functionName = functionResource.getMetadata().getName();
		this.functions.put(functionName, functionResource);
		HandlerReference handlerReference = FunctionUtils.adjustHandlerReference(functionResource);
		Handler handlerResource = handlers.get(handlerReference);
		Dispatcher dispatcher = this.getDispatcher(handlerResource.getSpec().getDispatcher());
		dispatcher.init(functionResource, handlerResource);
		addListener(functionResource);
		logger.info("function added: " + functionName);
	}

	@EventListener
	public void onFunctionUnregistered(ResourceDeletedEvent<XFunction> event) {
		XFunction functionResource = event.getResource();
		String functionName = functionResource.getMetadata().getName();
		removeListener(functionResource);
		HandlerReference handlerReference = FunctionUtils.adjustHandlerReference(functionResource);
		Handler handlerResource = handlers.get(handlerReference);
		Dispatcher dispatcher = this.getDispatcher(handlerResource.getSpec().getDispatcher());
		dispatcher.destroy(functionResource, handlerResource);
		this.functions.remove(functionName);
		logger.info("function deleted: " + functionName);
	}

	private void addListener(XFunction function) {
		String topic = function.getSpec().getInput();
		final String functionName = function.getMetadata().getName();
		DirectChannel channel = new DirectChannel();
		channel.subscribe(m -> {
			messageHandler.handleMessage(MessageBuilder.fromMessage(m)
					.setHeader("function", functionName)
					.setHeader(MessageHeaders.CONTENT_TYPE, "text/plain")
					.build());
		});
		Binding<?> binding = this.binder.bindConsumer(topic, functionName, channel, consumerProperties());
		this.bindings.put(functionName, binding);
	}

	private void removeListener(XFunction resource) {
		String functionName = resource.getMetadata().getName();
		Binding<?> binding = this.bindings.remove(functionName);
		if (binding != null) {
			binding.unbind();
		}
	}

	private ExtendedConsumerProperties<KafkaConsumerProperties> consumerProperties() {
		KafkaConsumerProperties kafkaProps = new KafkaConsumerProperties();
		ExtendedConsumerProperties<KafkaConsumerProperties> extendedProps = new ExtendedConsumerProperties<>(kafkaProps);
		extendedProps.setHeaderMode(HeaderMode.raw);
		return extendedProps;
	}
}
