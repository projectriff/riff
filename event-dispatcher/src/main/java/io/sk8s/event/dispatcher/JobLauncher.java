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

import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import org.springframework.beans.factory.annotation.Autowired;

import io.fabric8.kubernetes.api.model.ContainerBuilder;
import io.fabric8.kubernetes.api.model.EnvVar;
import io.fabric8.kubernetes.api.model.EnvVarBuilder;
import io.fabric8.kubernetes.api.model.JobSpecBuilder;
import io.fabric8.kubernetes.api.model.VolumeBuilder;
import io.fabric8.kubernetes.api.model.VolumeMountBuilder;
import io.fabric8.kubernetes.client.KubernetesClient;

import io.sk8s.core.function.Env;
import io.sk8s.core.function.FunctionResource;
import io.sk8s.core.handler.HandlerResource;

/**
 * @author Mark Fisher
 */
public class JobLauncher implements Dispatcher {

	private final KubernetesClient kubernetesClient;

	@Autowired
	private EventDispatcherProperties properties;

	public JobLauncher(KubernetesClient kubernetesClient) {
		this.kubernetesClient = kubernetesClient;
	}

	@Override
	public void init(FunctionResource functionResource, HandlerResource handlerResource) {
	}

	@Override
	public void dispatch(String payload, Map<String, Object> headers, FunctionResource functionResource, HandlerResource handlerResource) {
		String functionName = functionResource.getMetadata().get("name");
		String job = this.kubernetesClient.extensions().jobs().inNamespace(this.properties.getNamespace()).createNew()
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
		System.out.println("JOB: " + job);
	}

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
