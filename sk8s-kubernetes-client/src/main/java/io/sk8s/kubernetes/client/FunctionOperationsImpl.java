
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

package io.sk8s.kubernetes.client;

import java.util.Map;
import java.util.TreeMap;

import io.fabric8.kubernetes.client.Config;
import io.fabric8.kubernetes.client.dsl.Resource;
import io.fabric8.kubernetes.client.dsl.base.HasMetadataOperation;
import io.sk8s.kubernetes.api.model.DoneableHandler;
import io.sk8s.kubernetes.api.model.DoneableXFunction;
import io.sk8s.kubernetes.api.model.FunctionList;
import io.sk8s.kubernetes.api.model.Handler;
import io.sk8s.kubernetes.api.model.HandlerList;
import io.sk8s.kubernetes.api.model.XFunction;
import okhttp3.OkHttpClient;

public class FunctionOperationsImpl
		extends HasMetadataOperation<XFunction, FunctionList, DoneableXFunction, Resource<XFunction, DoneableXFunction>> {

	public FunctionOperationsImpl(OkHttpClient client, Config config, String namespace) {
		this(client, config, "v1", namespace, null, true, null, null, false, -1, new TreeMap<String, String>(),
				new TreeMap<String, String>(), new TreeMap<String, String[]>(), new TreeMap<String, String[]>(),
				new TreeMap<String, String>());
	}

	public FunctionOperationsImpl(OkHttpClient client, Config config, String apiVersion, String namespace, String name,
	                              Boolean cascading, XFunction item, String resourceVersion, Boolean reloadingFromServer, long gracePeriodSeconds,
	                              Map<String, String> labels, Map<String, String> labelsNot, Map<String, String[]> labelsIn,
	                              Map<String, String[]> labelsNotIn, Map<String, String> fields) {
		super(client, config, "extensions.sk8s.io", apiVersion, "functions", namespace, name, cascading, item,
				resourceVersion, reloadingFromServer, gracePeriodSeconds, labels, labelsNot, labelsIn, labelsNotIn,
				fields);
	}
}
