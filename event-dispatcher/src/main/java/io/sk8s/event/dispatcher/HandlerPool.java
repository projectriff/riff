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

import java.util.Collections;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

import org.springframework.cloud.stream.binder.Binder;
import org.springframework.cloud.stream.binder.BinderFactory;
import org.springframework.cloud.stream.binder.ExtendedConsumerProperties;
import org.springframework.cloud.stream.binder.ExtendedProducerProperties;
import org.springframework.cloud.stream.binder.HeaderMode;
import org.springframework.cloud.stream.binder.kafka.properties.KafkaConsumerProperties;
import org.springframework.cloud.stream.binder.kafka.properties.KafkaProducerProperties;
import org.springframework.context.SmartLifecycle;
import org.springframework.http.ResponseEntity;
import org.springframework.integration.channel.DirectChannel;
import org.springframework.integration.support.MessageBuilder;
import org.springframework.messaging.MessageChannel;
import org.springframework.messaging.MessageHeaders;
import org.springframework.web.client.RestTemplate;

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
import io.fabric8.kubernetes.client.KubernetesClientException;
import io.fabric8.kubernetes.client.Watch;
import io.fabric8.kubernetes.client.Watcher;

import io.sk8s.core.function.FunctionResource;
import io.sk8s.core.handler.HandlerResource;

/**
 * @author Mark Fisher
 */
public class HandlerPool implements Dispatcher, SmartLifecycle {

	private static Log logger = LogFactory.getLog(HandlerPool.class);

	private final KubernetesClient kubernetesClient;

	private final Binder binder;

	private final RestTemplate restTemplate = new RestTemplate();

	private final AtomicBoolean running = new AtomicBoolean();

	private final Map<String, Deployment> handlerDeployments = new HashMap<>();

	private final Map<String, Service> services = new HashMap<>();

	private final ConcurrentMap<String, CountDownLatch> serviceLatches = new ConcurrentHashMap<>();

	private final Map<String, Map<String, Pod>> handlerPods = new HashMap<>();

	private final Map<String, Pod> functionPods = new HashMap<>();

	private final Map<String, MessageChannel> outputChannels = new HashMap<>();

	private Watch poolWatch;

	private Watch serviceWatch;

	private Watch endpointsWatch;

	private Watch functionWatch;

	@SuppressWarnings("unchecked")
	public HandlerPool(KubernetesClient kubernetesClient, BinderFactory binderFactory) {
		this.kubernetesClient = kubernetesClient;
		this.binder = (Binder<MessageChannel,
				ExtendedConsumerProperties<KafkaConsumerProperties>,
				ExtendedProducerProperties<KafkaProducerProperties>>)
				binderFactory.getBinder("kafka", MessageChannel.class);
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

	@Override
	public void start() {
		this.poolWatch = kubernetesClient.extensions().deployments().watch(new Watcher<Deployment>() {

			@Override
			public void eventReceived(Watcher.Action action, Deployment deployment) {
				String name = deployment.getMetadata().getName();
				switch (action) {
				case DELETED:
					handlerDeployments.remove(name);
					break;
				case ADDED:
				case MODIFIED:
					handlerDeployments.put(name, deployment);
				default:
					break;
				}
			}

			@Override
			public void onClose(KubernetesClientException exception) {
			}
		});

		this.serviceWatch = kubernetesClient.services().watch(new Watcher<Service>() {

			@Override
			public void eventReceived(Action action, Service service) {
				String name = service.getMetadata().getName();
				switch (action) {
				case DELETED:
					services.remove(service);
					break;
				case ADDED:
				case MODIFIED:
					logger.info("SERVICE " + action.name() + ": " + service);
					services.put(name, service);
				default:
					break;
				}				
			}

			@Override
			public void onClose(KubernetesClientException cause) {
			}
			
		});

		this.endpointsWatch = kubernetesClient.endpoints().watch(new Watcher<Endpoints>() {

			@Override
			public void eventReceived(Action action, Endpoints endpoints) {
				switch (action) {
				case ADDED:
				case MODIFIED:
					String functionName = endpoints.getMetadata().getLabels().get("function");
					if (functionName != null && endpoints.getSubsets().size() > 0) {
						logger.info("service ready for " + functionName);
						serviceLatches.putIfAbsent(functionName, new CountDownLatch(1));
						serviceLatches.get(functionName).countDown();
					}
					break;
				default:
					break;
				}
			}

			@Override
			public void onClose(KubernetesClientException cause) {
			}
			
		});

		this.functionWatch = this.kubernetesClient.pods().watch(new Watcher<Pod>() {

			@Override
			public void eventReceived(Action action, Pod pod) {
				String handlerName = pod.getMetadata().getLabels().get("handler");
				String functionName = pod.getMetadata().getLabels().get("function");
				if (functionName != null) {
					switch (action) {
					case DELETED:
						functionPods.remove(functionName);
						break;
					case ADDED:
					case MODIFIED:
						functionPods.put(functionName, pod);
						logger.info("FUNCTION POD " + action + ": " + pod);
					default:
						break;
					}
				}
				else if (handlerName != null) {
					String ip = pod.getStatus().getPodIP();
					switch (action) {
					case DELETED:
						handlerPods.get(handlerName).remove(ip);
						break;
					case ADDED:
					case MODIFIED:
						if (!handlerPods.containsKey(handlerName)) {
							handlerPods.put(handlerName, new HashMap<>());
						}
						handlerPods.get(handlerName).put(ip, pod);
					default:
						break;
					}
				}
			}

			@Override
			public void onClose(KubernetesClientException exception) {
			}
		});
	}

	@Override
	public void stop(Runnable callback) {
		this.stop();
		callback.run();
	}

	@Override
	public void stop() {
		if (this.running.compareAndSet(true, false)) {
			this.poolWatch.close();
			this.serviceWatch.close();
			this.endpointsWatch.close();
			this.functionWatch.close();
			this.poolWatch = null;
			this.serviceWatch = null;
			this.endpointsWatch = null;
			this.functionWatch = null;
		}
	}

	@Override
	public void init(FunctionResource functionResource, HandlerResource handlerResource) {
		String functionName = functionResource.getMetadata().get("name");
		Map<String, String> functionLabels = Collections.singletonMap("function", functionName);
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
					.build())
			.endSpec()
		.done();

		String handlerName = handlerResource.getMetadata().get("name");
		int poolSize = handlerResource.getSpec().getReplicas();
		if (!this.handlerDeployments.containsKey(handlerName)) {
			Map<String, String> handlerLabels = Collections.singletonMap("handler", handlerName);
			Map<String, Quantity> resourceRequests = new HashMap<>();
			resourceRequests.put("cpu", new Quantity("500m"));
			resourceRequests.put("memory", new Quantity("512Mi"));
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
									.build())
									.withNewResources()
										.withRequests(resourceRequests)
									.endResources()
								.build())
								.withVolumes(new VolumeBuilder()
									.withName("functions")
									.withNewHostPath("/functions")
								.build())
							.endSpec()
						.endTemplate()
					.endSpec()
				.build());
		}
	}

	@Override
	public void dispatch(String payload, Map<String, Object> headers, FunctionResource functionResource, HandlerResource handlerResource) {
		String functionName = functionResource.getMetadata().get("name");
		Pod pod = this.functionPods.get(functionName);
		if (pod == null) {
			Map<String, String> functionLabels = Collections.singletonMap("function", functionName);
			String handlerName = handlerResource.getMetadata().get("name");
			Pod handlerPod = this.kubernetesClient.pods().withLabel("handler", handlerName).list().getItems().get(0);
			handlerPod.getMetadata().setLabels(functionLabels);
			pod = this.kubernetesClient.pods().createOrReplace(handlerPod);
			FunctionResource.FunctionSpec functionSpec = functionResource.getSpec();
			bindOutputChannel(functionName, functionSpec.getOutput());
			this.serviceLatches.putIfAbsent(functionName, new CountDownLatch(1));
			try {
				this.serviceLatches.get(functionName).await(10, TimeUnit.SECONDS);
			}
			catch (InterruptedException interrupted) {
				Thread.currentThread().interrupt();
				logger.error("timed out waiting for service for function: " + functionName);
				return;
			}
			Service service = this.services.get(functionName);
			InitPayload initPayload = new InitPayload(functionSpec.getParam("uri"), functionSpec.getParam("classname"));
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
			ResponseEntity<String> response = this.restTemplate.postForEntity(baseUrl + "/invoke", payload, String.class);
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
		ExtendedProducerProperties<KafkaProducerProperties> props = new ExtendedProducerProperties<>(new KafkaProducerProperties());
		props.setHeaderMode(HeaderMode.raw);
		this.binder.bindProducer(topic, channel, props);
	}

	private void sendResponse(String functionName, String payload) {
		this.outputChannels.get(functionName).send(MessageBuilder.withPayload(payload)
				.setHeader(MessageHeaders.CONTENT_TYPE, "text/plain").build());
	}

	/**
	 * The payload that will be POSTed as JSon to the {@literal init} endpoint of the web handler.
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
