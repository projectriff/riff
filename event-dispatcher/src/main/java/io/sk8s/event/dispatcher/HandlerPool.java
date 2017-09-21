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

import static io.sk8s.kubernetes.api.model.FunctionUtils.param;

import java.util.Collections;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;

import io.fabric8.kubernetes.api.model.ContainerBuilder;
import io.fabric8.kubernetes.api.model.Endpoints;
import io.fabric8.kubernetes.api.model.Pod;
import io.fabric8.kubernetes.api.model.Quantity;
import io.fabric8.kubernetes.api.model.Service;
import io.fabric8.kubernetes.api.model.ServicePortBuilder;
import io.fabric8.kubernetes.api.model.VolumeBuilder;
import io.fabric8.kubernetes.api.model.VolumeMountBuilder;
import io.fabric8.kubernetes.api.model.extensions.Deployment;
import io.fabric8.kubernetes.api.model.extensions.DeploymentBuilder;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.Watch;
import io.fabric8.kubernetes.client.Watcher;
import io.sk8s.core.resource.ResourceAddedOrModifiedEvent;
import io.sk8s.core.resource.ResourceDeletedEvent;
import io.sk8s.core.resource.ResourceEvent;
import io.sk8s.kubernetes.api.model.FunctionSpec;
import io.sk8s.kubernetes.api.model.FunctionUtils;
import io.sk8s.kubernetes.api.model.Handler;
import io.sk8s.kubernetes.api.model.XFunction;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import org.springframework.cloud.stream.binder.Binder;
import org.springframework.cloud.stream.binder.BinderFactory;
import org.springframework.cloud.stream.binder.ExtendedConsumerProperties;
import org.springframework.cloud.stream.binder.ExtendedProducerProperties;
import org.springframework.cloud.stream.binder.HeaderMode;
import org.springframework.cloud.stream.binder.kafka.properties.KafkaConsumerProperties;
import org.springframework.cloud.stream.binder.kafka.properties.KafkaProducerProperties;
import org.springframework.context.SmartLifecycle;
import org.springframework.context.event.EventListener;
import org.springframework.http.ResponseEntity;
import org.springframework.integration.channel.DirectChannel;
import org.springframework.integration.support.MessageBuilder;
import org.springframework.messaging.MessageChannel;
import org.springframework.messaging.MessageHeaders;
import org.springframework.web.client.RestTemplate;

/**
 * A {@link Dispatcher} that maintains a pool of {@link Handler handlers} ready for
 * dispatching function invocation.
 *
 * @author Mark Fisher
 * @author Eric Bottard
 */
public class HandlerPool implements Dispatcher, SmartLifecycle {

	private static Logger logger = LoggerFactory.getLogger(HandlerPool.class);

	private final KubernetesClient kubernetesClient;

	private final Binder binder;

	private final RestTemplate restTemplate = new RestTemplate();

	private final AtomicBoolean running = new AtomicBoolean();

	// TODO Use ObjectReference as key
	private final Map<String, Deployment> handlerDeployments = new HashMap<>();

	// TODO Use ObjectReference as key
	private final Map<String, Service> services = new HashMap<>();

	private final ConcurrentMap<String, CountDownLatch> serviceLatches = new ConcurrentHashMap<>();

	// TODO Use ObjectReference as key
	private final Map<String, Pod> functionPods = new HashMap<>();

	private final Map<String, MessageChannel> outputChannels = new HashMap<>();

	@SuppressWarnings("unchecked")
	public HandlerPool(KubernetesClient kubernetesClient, BinderFactory binderFactory) {
		this.kubernetesClient = kubernetesClient;
		this.binder = (Binder<MessageChannel, ExtendedConsumerProperties<KafkaConsumerProperties>, ExtendedProducerProperties<KafkaProducerProperties>>) binderFactory
				.getBinder("kafka", MessageChannel.class);
	}

	@Override
	public boolean isRunning() {
		return this.running.get();
	}

	@Override
	public int getPhase() {
		return 0;
	}

	@Override
	public boolean isAutoStartup() {
		return true;
	}

	@EventListener
	public void onDeploymentAdded(ResourceAddedOrModifiedEvent<Deployment> event) {
		if (isRunning()) {
			String name = event.getResource().getMetadata().getName();
			handlerDeployments.put(name, event.getResource());
		}
	}

	@EventListener
	public void onDeploymentDeleted(ResourceDeletedEvent<Deployment> event) {
		if (isRunning()) {
			String name = event.getResource().getMetadata().getName();
			handlerDeployments.remove(name);
		}
	}

	@EventListener
	public void onServiceAdded(ResourceAddedOrModifiedEvent<Service> event) {
		if (isRunning()) {
			String name = event.getResource().getMetadata().getName();
			logger.info("SERVICE {}: {}", event.getAction(), name);
			services.put(name, event.getResource());
		}
	}

	@EventListener
	public void onServiceDeleted(ResourceDeletedEvent<Service> event) {
		if (isRunning()) {
			String name = event.getResource().getMetadata().getName();
			services.remove(name);
			logger.info("SERVICE {}: {}", event.getAction(), name);
		}
	}

	@EventListener(condition = "#event.resource.metadata.labels?.get('function') != null " +
			"&& !#event.resource.subsets.empty")
	public void onEndpointAdded(ResourceAddedOrModifiedEvent<Endpoints> event) {
		if (isRunning()) {
			String functionName = event.getResource().getMetadata().getLabels().get("function");
			serviceLatches.putIfAbsent(functionName, new CountDownLatch(1));
			serviceLatches.get(functionName).countDown();
		}
	}

	@EventListener(condition = "#event.resource.metadata.labels?.get('function') != null ")
	public void onFunctionPodAdded(ResourceAddedOrModifiedEvent<Pod> event) {
		if (isRunning()) {
			Pod pod = event.getResource();
			String functionName = pod.getMetadata().getLabels().get("function");
			functionPods.put(functionName, pod);
			logger.info("FUNCTION POD {}: {}", event.getAction(), pod.getMetadata().getName());
		}
	}

	@EventListener(condition = "#event.resource.metadata.labels?.get('function') != null ")
	public void onFunctionPodRemoved(ResourceDeletedEvent<Pod> event) {
		if (isRunning()) {
			Pod pod = event.getResource();
			String functionName = pod.getMetadata().getLabels().get("function");
			functionPods.remove(functionName);
		}
	}

	@Override
	public void start() {
		this.running.compareAndSet(false, true);
	}

	@Override
	public void stop(Runnable callback) {
		this.stop();
		callback.run();
	}

	@Override
	public void stop() {
		this.running.compareAndSet(true, false);
	}

	@Override
	public void init(XFunction functionResource, Handler handlerResource) {
		String functionName = functionResource.getMetadata().getName();
		Map<String, String> functionLabels = Collections.singletonMap("function", functionName);
		//@formatter:off
		this.kubernetesClient.services().createNew()
			.withNewMetadata()
				.withName(functionName)
				.withLabels(functionLabels)
			.endMetadata()
			.withNewSpec()
				.withSelector(functionLabels)
				.withPorts(new ServicePortBuilder()
					.withPort(80)
					.withNewTargetPort(8080)
					.build()
				)
			.endSpec()
			.done();
		//@formatter:on

		String handlerName = handlerResource.getMetadata().getName();
		Integer poolSize = handlerResource.getSpec().getReplicas();
		if (poolSize == null) {
			poolSize = 1;
		}
		if (!this.handlerDeployments.containsKey(handlerName)) {
			Map<String, String> handlerLabels = Collections.singletonMap("handler", handlerName);
			Map<String, Quantity> resourceRequests = new HashMap<>();
			resourceRequests.put("cpu", new Quantity("500m"));
			resourceRequests.put("memory", new Quantity("512Mi"));
			//@formatter:off
			this.kubernetesClient.extensions().deployments().create(new DeploymentBuilder()
				.withNewMetadata()
					.withName(handlerName)
					.withLabels(handlerLabels)
				.endMetadata()
				.withNewSpec()
					.withReplicas(poolSize)
					.withNewSelector()
						.withMatchLabels(handlerLabels)
					.endSelector()
					.withNewTemplate()
						.withNewMetadata()
							.withName(handlerName)
							.withLabels(handlerLabels)
						.endMetadata()
						.withNewSpec()
							.withContainers(new ContainerBuilder()
								.withName("main")
								.withImage(handlerResource.getSpec().getImage())
								//.withCommand(handlerResource.getSpec().getCommand())
								//.withArgs(handlerResource.getSpec().getArgs())
								.withVolumeMounts(new VolumeMountBuilder()
									.withMountPath("/functions")
									.withName("functions")
									.build()
								)
								.withNewResources()
									.withRequests(resourceRequests)
								.endResources()
								.build()
							)
							.withVolumes(new VolumeBuilder()
								.withName("functions")
								.withNewHostPath("/functions")
								.build()
							)
						.endSpec()
					.endTemplate()
				.endSpec()
			.build()
			);
			//@formatter:on
		}
	}

	@Override
	public void destroy(XFunction functionResource, Handler handlerResource) {
		String functionName = functionResource.getMetadata().getName();
		Map<String, String> functionLabels = Collections.singletonMap("function", functionName);
		this.kubernetesClient.services().withLabels(functionLabels).delete();

		// TODO: "reference count" handlers?
	}

	@Override
	public void dispatch(String payload, Map<String, Object> headers, XFunction functionResource,
			Handler handlerResource) {
		String functionName = functionResource.getMetadata().getName();
		Pod pod = this.functionPods.get(functionName);
		if (pod == null) {
			Map<String, String> functionLabels = Collections.singletonMap("function", functionName);
			String handlerName = handlerResource.getMetadata().getName();
			Pod handlerPod = this.kubernetesClient.pods().withLabel("handler", handlerName).list().getItems().get(0);
			handlerPod.getMetadata().setLabels(functionLabels);
			pod = this.kubernetesClient.pods().createOrReplace(handlerPod);
			FunctionSpec functionSpec = functionResource.getSpec();
			bindOutputChannel(functionName, functionSpec.getOutput());
			try {
				logger.info("Waiting for service for function {}", functionName);
				this.serviceLatches.putIfAbsent(functionName, new CountDownLatch(1));
				this.serviceLatches.get(functionName).await(10, TimeUnit.SECONDS);
				logger.info("Done waiting for {} service", functionName);
			}
			catch (InterruptedException interrupted) {
				Thread.currentThread().interrupt();
				logger.error("timed out waiting for service for function: " + functionName);
				return;
			}
			Service service = this.services.get(functionName);
			InitPayload initPayload = new InitPayload(param("uri", functionResource), param("classname", functionResource));
			String url = "http://" + service.getSpec().getClusterIP() + "/init";
			logger.info("POST to /init for function '" + functionName + "' with params: " + initPayload);
			ResponseEntity<String> initResponse = this.restTemplate.postForEntity(
					url, initPayload, String.class);
			logger.info("Response: " + initResponse);
		}
		Service service = this.services.get(functionName);
		if (service != null) {
			String baseUrl = "http://" + service.getSpec().getClusterIP();
			logger.info("POST to " + baseUrl + "/invoke with message: " + payload);
			ResponseEntity<String> response = this.restTemplate.postForEntity(baseUrl + "/invoke", payload,
					String.class);
			logger.info("Response: " + response);
			sendResponse(functionName, response.getBody());
		}
		else {
			logger.info("failed to retrieve service for function: " + functionName);
		}
	}

	private void bindOutputChannel(String functionName, String topic) {
		DirectChannel channel = new DirectChannel();
		this.outputChannels.put(functionName, channel);
		ExtendedProducerProperties<KafkaProducerProperties> props = new ExtendedProducerProperties<>(
				new KafkaProducerProperties());
		props.setHeaderMode(HeaderMode.raw);
		this.binder.bindProducer(topic, channel, props);
	}

	private void sendResponse(String functionName, String payload) {
		this.outputChannels.get(functionName).send(MessageBuilder.withPayload(payload)
				.setHeader(MessageHeaders.CONTENT_TYPE, "text/plain").build());
	}

	/**
	 * The payload that will be POSTed as JSon to the {@literal init} endpoint of the web
	 * handler.
	 *
	 * @author Eric Bottard
	 */
	private static class InitPayload {
		private final String uri;

		private final String className;

		private InitPayload(String uri, String className) {
			this.uri = uri;
			this.className = className;
		}

		public String getUri() {
			return uri;
		}

		public String getClassName() {
			return className;
		}

		@Override
		public String toString() {
			return "{" +
					"uri='" + uri + '\'' +
					", className='" + className + '\'' +
					'}';
		}
	}
}
