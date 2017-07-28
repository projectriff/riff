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

package io.sk8s.function.controller;

import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.LineNumberReader;
import java.util.Collections;
import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.atomic.AtomicBoolean;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.SmartLifecycle;
import org.springframework.http.HttpMethod;
import org.springframework.util.StringUtils;
import org.springframework.web.client.RestTemplate;

import com.fasterxml.jackson.databind.ObjectMapper;

import io.fabric8.kubernetes.api.model.ContainerBuilder;
import io.fabric8.kubernetes.api.model.EnvVarBuilder;
import io.fabric8.kubernetes.api.model.PodSpecBuilder;
import io.fabric8.kubernetes.api.model.PodTemplateSpecBuilder;
import io.fabric8.kubernetes.api.model.ServicePort;
import io.fabric8.kubernetes.api.model.ServiceSpecBuilder;
import io.fabric8.kubernetes.api.model.extensions.DeploymentSpecBuilder;
import io.fabric8.kubernetes.client.DefaultKubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClient;

/**
 * @author Mark Fisher
 */
public class FunctionResourceWatcher implements SmartLifecycle {

	private static Log logger = LogFactory.getLog(FunctionResourceWatcher.class);

	@Autowired
	private FunctionControllerProperties properties;

	// assumes "kubectl proxy" running on localhost
	private final String url = "http://localhost:8001/apis/extensions.sk8s.io/v1/functions?watch=true";

	private final RestTemplate restTemplate = new RestTemplate();

	private final ObjectMapper objectMapper = new ObjectMapper();

	private final KubernetesClient kubernetesClient = new DefaultKubernetesClient();

	private final ExecutorService executor = Executors.newSingleThreadExecutor();

	private final AtomicBoolean running = new AtomicBoolean();

	private final int phase = 0;

	@Override
	public boolean isRunning() {
		return this.running.get();
	}

	@Override
	public int getPhase() {
		return this.phase;
	}

	@Override
	public boolean isAutoStartup() {
		return true;
	}

	@Override
	public void start() {
		if (!this.running.get()) {
			this.running.set(true);
			this.executor.submit(() -> {
				this.restTemplate.execute(this.url, HttpMethod.GET, (request) -> {}, (response) -> {
					InputStream inputStream = response.getBody();
					LineNumberReader reader = null;
					try {
						reader = new LineNumberReader(new InputStreamReader(inputStream));
						while (this.running.get()) {
							String line = reader.readLine();
							if (!StringUtils.hasText(line)) {
								break;
							}
							try {
								this.handleEvent(this.objectMapper.readValue(line, FunctionResourceEvent.class));
							}
							catch (Exception e) {
								e.printStackTrace();
							}
						}
					}
					finally {
						if (reader != null) {
							reader.close();
						}
					}
					return null;
				});
			});
		}
	}

	@Override
	public void stop(Runnable callback) {
		this.stop();
		callback.run();
	}

	@Override
	public void stop() {
		if (this.running.get()) {
			this.executor.shutdownNow();
		}
		this.running.set(false);
	}

	private void handleEvent(FunctionResourceEvent event) {
		logger.info("handling FunctionResourceEvent: " + event);
		if ("ADDED".equalsIgnoreCase(event.getType())) {
			addFunction(event.getResource());
		}
		else if ("DELETED".equalsIgnoreCase(event.getType())) {
			deleteFunction(event.getResource());
		}
		else {
			logger.debug("unhandled FunctionResourceEvent type: " + event.getType());
		}
	}

	private void addFunction(FunctionResource functionResource) {
		String functionName = functionResource.getMetadata().get("name");
		Map<String, String> labels = Collections.singletonMap("function", functionName);
		createDeployment(functionName, labels, functionResource.getSpec());
		createService(functionName, labels, 8080);
	}

	private void deleteFunction(FunctionResource functionResource) {
		String functionName = functionResource.getMetadata().get("name");
		deleteService(functionName);
		deleteDeployment(functionName);
	}

	private void createDeployment(String name, Map<String, String> labels, FunctionResource.Spec spec) {
		this.kubernetesClient.extensions().deployments().inNamespace(this.properties.getNamespace()).createNew()
			.withNewMetadata()
				.withName(name)
				.withLabels(labels)
			.endMetadata()
			.withSpec(new DeploymentSpecBuilder()
				.withReplicas(1)
				.withTemplate(new PodTemplateSpecBuilder()
					.withNewMetadata()
						.withName(name + "-function")
						.withLabels(labels)
					.endMetadata()
					.withSpec(new PodSpecBuilder()
						.withContainers(new ContainerBuilder()
							.withName("function-runner")
							.withImage(this.properties.getImage())
							.withEnv(new EnvVarBuilder()
								.withName("FUNCTION_RUNNER_LAMBDA")
								.withValue(spec.getLambda())
								.build())
							.build())
						.build())
					.build())
				.build())
			.done();
	}

	private void deleteDeployment(String name) {
		this.kubernetesClient.extensions().deployments().inNamespace(this.properties.getNamespace()).withName(name).delete();
	}

	private void createService(String name, Map<String, String> labels, int port) {
		ServicePort servicePort = new ServicePort();
		servicePort.setPort(port);
		this.kubernetesClient.services().inNamespace(this.properties.getNamespace()).createNew()
			.withNewMetadata()
				.withName(name)
				.withLabels(labels)
			.endMetadata()
			.withSpec(new ServiceSpecBuilder()
				.withType("NodePort")
				.withSelector(labels)
				.addNewPortLike(servicePort).endPort().build())
		.done();
	}

	private void deleteService(String name) {
		this.kubernetesClient.services().inNamespace(this.properties.getNamespace()).withName(name).delete();
	}
}
