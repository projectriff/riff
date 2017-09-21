
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

import io.fabric8.kubernetes.client.DefaultKubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.sk8s.kubernetes.client.Sk8sClient;

/**
 * Hello world!
 *
 */
public class App {
	public static void main(String[] args) {
		KubernetesClient kubernetesClient = new DefaultKubernetesClient();

		Sk8sClient sk8sClient = kubernetesClient.adapt(Sk8sClient.class);
		// sk8sClient.topics().inNamespace("default").createNew()
		// .withNewMetadata()
		// .withName("greetings")
		// .endMetadata()
		// .done();

		sk8sClient.topics().list().getItems().forEach(t -> System.out.println(t.getMetadata().getName()));

		System.out.println("ResourceVersion:" + sk8sClient.topics().list().getMetadata().getResourceVersion());

	}
}
