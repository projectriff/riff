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

package io.sk8s.core.resource;

import java.net.SocketTimeoutException;
import java.net.URL;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.atomic.AtomicBoolean;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

import org.springframework.context.SmartLifecycle;
import org.springframework.util.StringUtils;

import io.fabric8.kubernetes.client.DefaultKubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.utils.HttpClientUtils;

import com.fasterxml.jackson.databind.ObjectMapper;

import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;
import okio.BufferedSource;

/**
 * @author Mark Fisher
 */
public class ResourceWatcher<E extends ResourceEvent> implements SmartLifecycle {

	private static Log logger = LogFactory.getLog(ResourceWatcher.class);

	private final Class<E> resourceEventType;

	private final ResourceEventHandler handler;

	private final String url;

	private final KubernetesClient kubernetesClient;

	private final ObjectMapper objectMapper = new ObjectMapper();

	private final ExecutorService executor = Executors.newSingleThreadExecutor();

	private final AtomicBoolean running = new AtomicBoolean();

	private final int phase = 0;

	public ResourceWatcher(Class<E> resourceEventType, String pluralResourceName, ResourceEventHandler handler) {
		this.resourceEventType = resourceEventType;
		this.handler = handler;
		this.kubernetesClient = new DefaultKubernetesClient();
		URL masterUrl = kubernetesClient.getMasterUrl();
		// if "kubectl proxy" running on localhost can use localhost:8001
		this.url = masterUrl + "apis/extensions.sk8s.io/v1/" + pluralResourceName +"?watch=true";
		logger.info("URL: " + this.url);
	}

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
				try {
					while (this.running.get()) {
						// TODO: replace this code
						Response response = null;
						try {
							OkHttpClient httpClient = HttpClientUtils.createHttpClient(kubernetesClient.getConfiguration());
							Request.Builder requestBuilder = new Request.Builder().get().url(this.url);
							response = httpClient.newCall(requestBuilder.build()).execute();
							BufferedSource source = response.body().source();
							while (!source.exhausted()) {
								String line = source.readUtf8LineStrict();
								if (!StringUtils.hasText(line)) {
									break;
								}
								try {
									this.handleEvent(this.objectMapper.readValue(line, this.resourceEventType));
								}
								catch (Exception e) {
									e.printStackTrace();
									break;
								}
							}
						}
						catch (SocketTimeoutException e) {
							// reconnect
						}
						catch (Exception e) {
							e.printStackTrace();
						}
						finally {
							try {
								response.close();
							}
							catch (Exception e) {
								// ignore
							}
						}
					}
				}
				catch (Exception e) {
					e.printStackTrace();
				}
				return null;
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

	private void handleEvent(ResourceEvent event) {
		//logger.info("handling ResourceEvent: " + event);
		if ("ADDED".equalsIgnoreCase(event.getType())) {
			handler.resourceAdded(event.getResource());
		}
		else if ("DELETED".equalsIgnoreCase(event.getType())) {
			handler.resourceDeleted(event.getResource());
		}
		else {
			logger.debug("unhandled ResourceEvent type: " + event.getType());
		}
	}
}
