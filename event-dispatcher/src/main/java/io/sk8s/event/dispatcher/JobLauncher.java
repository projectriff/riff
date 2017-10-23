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
import java.util.List;
import java.util.Map;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.util.ObjectUtils;

import io.fabric8.kubernetes.api.model.Container;
import io.fabric8.kubernetes.api.model.ContainerBuilder;
import io.fabric8.kubernetes.api.model.EmptyDirVolumeSourceBuilder;
import io.fabric8.kubernetes.api.model.EnvVar;
import io.fabric8.kubernetes.api.model.EnvVarBuilder;
import io.fabric8.kubernetes.api.model.JobSpecBuilder;
import io.fabric8.kubernetes.api.model.PodSpec;
import io.fabric8.kubernetes.api.model.PodSpecBuilder;
import io.fabric8.kubernetes.api.model.VolumeBuilder;
import io.fabric8.kubernetes.api.model.VolumeMount;
import io.fabric8.kubernetes.api.model.VolumeMountBuilder;
import io.fabric8.kubernetes.client.KubernetesClient;

import io.sk8s.kubernetes.api.model.FunctionEnvVar;
import io.sk8s.kubernetes.api.model.XFunction;

/**
 * @author Mark Fisher
 */
public class JobLauncher {

	private final KubernetesClient kubernetesClient;

	private final ObjectMapper objectMapper = new ObjectMapper();

	@Autowired
	private EventDispatcherProperties properties;

	public JobLauncher(KubernetesClient kubernetesClient) {
		this.kubernetesClient = kubernetesClient;
	}

	public void dispatch(XFunction functionResource) {
		String functionName = functionResource.getMetadata().getName();
		// @formatter:off
		String job = this.kubernetesClient.extensions().jobs().inNamespace(this.properties.getNamespace()).createNew()
			.withApiVersion("batch/v1")
			.withNewMetadata()
				.withName(functionName + "-" + System.currentTimeMillis())
				.addNewOwnerReference()
					.withApiVersion(functionResource.getApiVersion())
					.withKind(functionResource.getKind())
					.withName(functionName)
					.withUid(functionResource.getMetadata().getUid())
					.withController(true)
				.endOwnerReference()
			.endMetadata()
			.withSpec(
				new JobSpecBuilder()
					.withNewTemplate()
						.withNewMetadata()
							.withLabels(Collections.singletonMap("function", functionName))
						.endMetadata()
						.withSpec(buildPodSpec(functionResource))
					.endTemplate()
				.build()
			)
			.done().toString();
		// @formatter:on
		System.out.println("JOB: " + job);
	}

	private PodSpec buildPodSpec(XFunction function) {
		PodSpecBuilder builder = new PodSpecBuilder()
				.withRestartPolicy("OnFailure")
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
				.withImage("sk8s/function-sidecar:0.0.1-SNAPSHOT")
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
		Map<String, String> springApplicationConfig = new HashMap<>();

		springApplicationConfig.put("spring.cloud.stream.bindings.input.destination", function.getSpec().getInput());
		springApplicationConfig.put("spring.cloud.stream.bindings.input.group", function.getMetadata().getName());
		springApplicationConfig.put("spring.cloud.stream.bindings.output.destination", function.getSpec().getOutput());
		springApplicationConfig.put("spring.cloud.stream.kafka.binder.brokers",
				System.getenv("SPRING_CLOUD_STREAM_KAFKA_BINDER_BROKERS"));
		springApplicationConfig.put("spring.cloud.stream.kafka.binder.zkNodes",
				System.getenv("SPRING_CLOUD_STREAM_KAFKA_BINDER_ZK_NODES"));
		springApplicationConfig.put("spring.application.name","sidecar-" + function.getMetadata().getName());
		springApplicationConfig.put("server.port", "7433");
		springApplicationConfig.put("spring.profiles.active",  function.getSpec().getProtocol());

		String json;
		try {
			json = objectMapper.writeValueAsString(springApplicationConfig);
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
