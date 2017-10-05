/*
 * Copyright 2017 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package io.sk8s.test;

import io.fabric8.kubernetes.client.DefaultKubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClient;
import org.springframework.cloud.stream.test.junit.AbstractExternalResourceTestSupport;

/**
 * @author David Turanski
 **/
public class KubernetesAvailableRule extends AbstractExternalResourceTestSupport {
	protected final KubernetesClient client;

	public KubernetesAvailableRule() {
		this("KUBERNETES");
	}

	protected KubernetesAvailableRule(String resourceDescription) {
		super(resourceDescription);
		this.client = new DefaultKubernetesClient();
	}

	@Override
	protected void cleanupResource() throws Exception {

	}

	@Override
	protected void obtainResource() throws Exception {
		try {
			this.client.services().list();
		}
		catch (Exception e) {
			throw new RuntimeException("Kubernetes is not available.");
		}
	}
}
