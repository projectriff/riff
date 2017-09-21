
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

import java.lang.reflect.Field;
import java.util.Map;

import io.fabric8.kubernetes.client.BaseClient;
import io.fabric8.kubernetes.client.Config;
import io.fabric8.kubernetes.client.dsl.MixedOperation;
import io.fabric8.kubernetes.client.dsl.Resource;
import io.fabric8.kubernetes.internal.KubernetesDeserializer;
import io.sk8s.kubernetes.api.model.DoneableHandler;
import io.sk8s.kubernetes.api.model.DoneableTopic;
import io.sk8s.kubernetes.api.model.DoneableXFunction;
import io.sk8s.kubernetes.api.model.FunctionList;
import io.sk8s.kubernetes.api.model.Handler;
import io.sk8s.kubernetes.api.model.HandlerList;
import io.sk8s.kubernetes.api.model.Topic;
import io.sk8s.kubernetes.api.model.TopicList;
import io.sk8s.kubernetes.api.model.XFunction;
import okhttp3.OkHttpClient;

public class DefaultSk8sClient extends BaseClient implements Sk8sClient {

	static {
		hackKubernetesDeserializer();
	}

	private static void hackKubernetesDeserializer() {
		try {
			Field mapField = KubernetesDeserializer.class.getDeclaredField("MAP");
			mapField.setAccessible(true);
			Map<String, Class<?>> map = (Map<String, Class<?>>) mapField.get(null);

			map.put("Topic", Topic.class);
			map.put("Handler", Handler.class);
			map.put("Function", XFunction.class);
		}
		catch (NoSuchFieldException | IllegalAccessException e) {
			throw new RuntimeException(e);
		}
	}

	public DefaultSk8sClient(OkHttpClient okHttpClient, Config configuration) {
		super(okHttpClient, configuration);
	}

	@Override
	public MixedOperation<Topic, TopicList, DoneableTopic, Resource<Topic, DoneableTopic>> topics() {
		return new TopicOperationsImpl(httpClient, getConfiguration(), getNamespace());
	}

	@Override
	public MixedOperation<Handler, HandlerList, DoneableHandler, Resource<Handler, DoneableHandler>> handlers() {
		return new HandlerOperationsImpl(httpClient, getConfiguration(), getNamespace());
	}

	@Override
	public MixedOperation<XFunction, FunctionList, DoneableXFunction, Resource<XFunction, DoneableXFunction>> functions() {
		return new FunctionOperationsImpl(httpClient, getConfiguration(), getNamespace());
	}
}
