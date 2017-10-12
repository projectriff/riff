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
import java.util.List;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.util.ObjectUtils;

import io.fabric8.kubernetes.api.model.Container;
import io.fabric8.kubernetes.api.model.ContainerBuilder;
import io.fabric8.kubernetes.api.model.EmptyDirVolumeSourceBuilder;
import io.fabric8.kubernetes.api.model.EnvVar;
import io.fabric8.kubernetes.api.model.EnvVarBuilder;
import io.fabric8.kubernetes.api.model.JobSpecBuilder;
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
			.endMetadata()
			.withSpec(
				new JobSpecBuilder()
				.withNewTemplate()
					.withNewMetadata()
						.withLabels(Collections.singletonMap("function", functionName))
					.endMetadata()
					.withNewSpec()
						.withRestartPolicy("OnFailure")
						.withContainers(buildMainContainer(functionResource), buildSidecarContainer(functionResource))
						.withVolumes(new VolumeBuilder()
								.withName("pipes")
								.withEmptyDir(new EmptyDirVolumeSourceBuilder().build())
								.build(),
							new VolumeBuilder()
								.withName("resources")
								.withNewHostPath("/resources")
								.build()
						)
					.endSpec()
				.endTemplate()
				.build()
			)
			.done().toString();
		// @formatter:on
		System.out.println("JOB: " + job);
	}

	private Container buildMainContainer(XFunction function) {
		ContainerBuilder builder = new ContainerBuilder().withName("main")
				.withImage(function.getSpec().getImage())
				.withImagePullPolicy("IfNotPresent")
				.withVolumeMounts(buildSharedVolumeMounts());
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
		return new ContainerBuilder().withName("sidecar")
				.withImage("sk8s/function-sidecar:v0001")
				.withImagePullPolicy("IfNotPresent")
				.withEnv(buildSidecarEnvVars(function))
				.withVolumeMounts(buildSharedVolumeMounts())
				.build();
	}

	private VolumeMount[] buildSharedVolumeMounts() {
		return new VolumeMount[] {
				new VolumeMountBuilder().withMountPath("/pipes").withName("pipes").build(),
				new VolumeMountBuilder().withMountPath("/resources").withName("resources").build()
		};
	}

	private EnvVar[] buildSidecarEnvVars(XFunction function) {
		String json = "{\"spring.cloud.stream.bindings.input.destination\":\"" + function.getSpec().getInput() + "\","
				+ "\"spring.cloud.stream.bindings.output.destination\":\"" + function.getSpec().getOutput() + "\","
				+ "\"spring.cloud.stream.kafka.binder.brokers\":\"kafka:9092\","
				+ "\"spring.cloud.stream.kafka.binder.zkNodes\":\"zookeeper:2181\","
				+ "\"server.port\":\"7433\","
				+ "\"spring.profiles.active\":\"" + function.getSpec().getProtocol() + "\","
				+ "\"spring.application.name\":\"sidecar-" + function.getSpec().getInput() + "\"}";
		return new EnvVar[] {
				new EnvVarBuilder().withName("SPRING_APPLICATION_JSON").withValue(json).build(),
		};
	}
}
