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
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.fabric8.kubernetes.api.model.Container;
import io.fabric8.kubernetes.api.model.ContainerBuilder;
import io.fabric8.kubernetes.api.model.EmptyDirVolumeSourceBuilder;
import io.fabric8.kubernetes.api.model.EnvVar;
import io.fabric8.kubernetes.api.model.EnvVarBuilder;
import io.fabric8.kubernetes.api.model.PodSpec;
import io.fabric8.kubernetes.api.model.PodSpecBuilder;
import io.fabric8.kubernetes.api.model.VolumeBuilder;
import io.fabric8.kubernetes.api.model.VolumeMount;
import io.fabric8.kubernetes.api.model.VolumeMountBuilder;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.sk8s.kubernetes.api.model.FunctionEnvVar;
import io.sk8s.kubernetes.api.model.XFunction;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.util.ObjectUtils;

/**
 * Makes sure a certain function is running on Kubernetes.
 *
 * <p>This implementation uses k8s deployments with a desired replicas count (possibly 0).</p>
 *
 * @author Eric Bottard
 * @author Mark Fisher
 */
public class FunctionDeployer {

	private final static Logger logger = LoggerFactory.getLogger(FunctionDeployer.class);

	private final KubernetesClient kubernetesClient;

	private final ObjectMapper objectMapper = new ObjectMapper();

	@Autowired
	private EventDispatcherProperties properties;

	public FunctionDeployer(KubernetesClient kubernetesClient) {
		this.kubernetesClient = kubernetesClient;
	}

	/**
	 * Requests that the given function be deployed with N replicas.
	 */
	public void deploy(XFunction functionResource, int replicas) {
		String functionName = functionResource.getMetadata().getName();
		logger.debug("Setting {} replicas for {}", replicas, functionName);
		// @formatter:off
		this.kubernetesClient.extensions().deployments()
				.inNamespace(functionResource.getMetadata().getNamespace())
				.createOrReplaceWithNew()
					.withApiVersion("extensions/v1beta1")
					.withNewMetadata()
						.withName(functionName)
						.addNewOwnerReference()
							.withApiVersion(functionResource.getApiVersion())
							.withKind(functionResource.getKind())
							.withName(functionResource.getMetadata().getName())
							.withUid(functionResource.getMetadata().getUid())
							.withController(true)
						.endOwnerReference()
					.endMetadata()
					.withNewSpec()
						.withReplicas(replicas)
						.withNewTemplate()
							.withNewMetadata()
								.withName(functionName)
								.withLabels(Collections.singletonMap("function", functionName))
							.endMetadata()
							.withSpec(buildPodSpec(functionResource))
						.endTemplate()
					.endSpec()
				.done();
		// @formatter:on
	}

	/**
	 * Returns the system to a clean slate regarding the deployment of the given function.
	 */
	public void undeploy(XFunction function) {
		String functionName = function.getMetadata().getName();
		this.kubernetesClient.extensions().deployments()
				.inNamespace(function.getMetadata().getNamespace())
				.withName(functionName)
				.delete();
	}

	private PodSpec buildPodSpec(XFunction function) {
		PodSpecBuilder builder = new PodSpecBuilder()
				.withContainers(buildMainContainer(function), buildSidecarContainer(function));
		if ("stdio".equals(function.getSpec().getProtocol())) {
			builder.withVolumes(new VolumeBuilder()
					.withName("pipes")
					.withEmptyDir(new EmptyDirVolumeSourceBuilder().build())
					.build());
		}
		return builder.build();
	}

	private Container buildMainContainer(XFunction function) {
		ContainerBuilder builder = new ContainerBuilder().withName("main")
				.withImage(function.getSpec().getImage())
				.withImagePullPolicy("IfNotPresent");
		if ("stdio".equals(function.getSpec().getProtocol())) {
			builder.withVolumeMounts(buildNamedPipesMount());
		}
		List<String> command = function.getSpec().getCommand();
		if (!ObjectUtils.isEmpty(command)) {
			builder.addAllToCommand(command);
		}
		List<String> args = function.getSpec().getArgs();
		if (!ObjectUtils.isEmpty(args)) {
			builder.addAllToArgs(args);
		}
		List<FunctionEnvVar> envVars = function.getSpec().getEnv();
		if (!ObjectUtils.isEmpty(envVars)) {
			for (FunctionEnvVar envVar : envVars) {
				builder.addToEnv(new EnvVarBuilder()
						.withName(envVar.getName())
						.withValue(envVar.getValue())
						.build());
			}
		}
		return builder.build();
	}

	private Container buildSidecarContainer(XFunction function) {
		ContainerBuilder builder = new ContainerBuilder().withName("sidecar")
				.withImage(properties.getSidecarImageName())
				.withImagePullPolicy("IfNotPresent")
				.withEnv(buildSidecarEnvVars(function));
		if ("stdio".equals(function.getSpec().getProtocol())) {
			builder.withVolumeMounts(buildNamedPipesMount());
		}
		return builder.build();
	}

	private VolumeMount buildNamedPipesMount() {
		return new VolumeMountBuilder().withMountPath("/pipes").withName("pipes").build();
	}

	private EnvVar[] buildSidecarEnvVars(XFunction function) {
		Map<String, Object> config = new LinkedHashMap<>();
		config.put("spring.cloud.stream.bindings.input.destination", function.getSpec().getInput());
		config.put("spring.cloud.stream.bindings.output.destination", function.getSpec().getOutput());
		config.put("spring.cloud.stream.bindings.input.group", function.getMetadata().getName());
		config.put("spring.profiles.active", function.getSpec().getProtocol());
		config.put("spring.application.name", "sidecar-" + function.getSpec().getInput());

		config.put("spring.cloud.stream.kafka.binder.brokers", System.getenv("SPRING_CLOUD_STREAM_KAFKA_BINDER_BROKERS"));
		config.put("spring.cloud.stream.kafka.binder.zkNodes", System.getenv("SPRING_CLOUD_STREAM_KAFKA_BINDER_ZK_NODES"));
		config.put("server.port", "-1");

		String json;
		try {
			json = objectMapper.writeValueAsString(config);
		}
		catch (JsonProcessingException e) {
			throw new RuntimeException(e);
		}

		return new EnvVar[] {
				new EnvVarBuilder().withName("JAVA_TOOL_OPTIONS").withValue("-Xmx512m").build(),
				new EnvVarBuilder().withName("SPRING_APPLICATION_JSON").withValue(json).build()

		};
	}
}
