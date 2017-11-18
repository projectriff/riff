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

import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.stream.Collectors;

import javax.annotation.PreDestroy;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.event.EventListener;

import io.fabric8.kubernetes.api.model.extensions.Deployment;

import io.sk8s.core.resource.ResourceAddedOrModifiedEvent;
import io.sk8s.core.resource.ResourceDeletedEvent;
import io.sk8s.kubernetes.api.model.Topic;
import io.sk8s.kubernetes.api.model.TopicSpec;
import io.sk8s.kubernetes.api.model.XFunction;

/**
 * Listens for function registration/un-registration and periodically scales deployment of
 * function pods based on the evaluation of a SpEL expression.
 *
 * @author Eric Bottard
 * @author Mark Fisher
 */
public class FunctionMonitor {

	private static final int SCALE_DOWN_LINGER_PERIOD = 10_000;

	private final Logger logger = LoggerFactory.getLogger(FunctionMonitor.class);

	// TODO: Change key to ObjectReference or similar for all these maps
	// Key is function name
	private final Map<String, XFunction> functions = new ConcurrentHashMap<>();

	private final Map<String, Topic> topics = new ConcurrentHashMap<>();

	private final Map<String, CountDownLatch> topicsReady = new ConcurrentHashMap<>();

	/** Keeps track of what the deployments ask. */
	private final Map<String, Integer> actualReplicaCount = new ConcurrentHashMap<>();

	private final Map<String, Integer> availableReplicaCount = new ConcurrentHashMap<>();

	/** Keeps track of the last time this decided to scale down (but not yet effective). */
	private final Map<String, Long> scaleDownStartTimes = new ConcurrentHashMap<>();

	@Autowired
	private FunctionDeployer deployer;

	@Autowired
	private EventPublisher publisher;

	private volatile long scalerIntervalMs = 0L; // Wait forever

	private final Thread scalerThread = new Thread(new Scaler());

	private final AtomicBoolean running = new AtomicBoolean();

	private final LagTracker lagTracker = new LagTracker(
			System.getenv("SPRING_CLOUD_STREAM_KAFKA_BINDER_BROKERS"));

	@EventListener
	public synchronized void onFunctionRegistered(ResourceAddedOrModifiedEvent<XFunction> event) {
		XFunction functionResource = event.getResource();
		String functionName = functionResource.getMetadata().getName();
		this.functions.put(functionName, functionResource);
		this.deployer.deploy(functionResource, 0);
		this.lagTracker.beginTracking(functionName, functionResource.getSpec().getInput());
		this.updateScalerInterval();
		if (this.running.compareAndSet(false, true)) {
			this.scalerThread.start();
		}
		this.notify();
		logger.info("function added: " + functionName);
	}

	@EventListener
	public synchronized void onFunctionUnregistered(ResourceDeletedEvent<XFunction> event) {
		XFunction functionResource = event.getResource();
		String functionName = functionResource.getMetadata().getName();
		this.functions.remove(functionName);
		this.actualReplicaCount.remove(functionName);
		this.deployer.undeploy(functionResource);
		this.lagTracker.stopTracking(functionName, functionResource.getSpec().getInput());
		this.updateScalerInterval();
		this.notify();
		logger.info("function deleted: " + functionName);
	}

	@EventListener
	public void onTopicRegistered(ResourceAddedOrModifiedEvent<Topic> event) {
		Topic topic = event.getResource();
		String name = topic.getMetadata().getName();
		this.topics.put(name, topic);
		this.topicsReady.computeIfAbsent(name, i -> new CountDownLatch(1)).countDown();
	}

	@EventListener
	public void onTopicUnregistered(ResourceDeletedEvent<Topic> event) {
		Topic topic = event.getResource();
		String name = topic.getMetadata().getName();
		this.topics.remove(name);
		this.topicsReady.remove(name);
	}

	@EventListener
	public void onDeploymentRegistered(ResourceAddedOrModifiedEvent<Deployment> event) {
		Deployment deployment = event.getResource();
		if (deployment.getMetadata().getLabels().containsKey("function")) {
			String functionName = deployment.getMetadata().getName();
			Integer replicas = deployment.getStatus().getReplicas();
			replicas = (replicas != null) ? replicas : 0;
			Integer availableReplicas = deployment.getStatus().getAvailableReplicas();
			availableReplicas = (availableReplicas != null) ? availableReplicas : 0;
			this.actualReplicaCount.put(functionName, replicas);
			Integer previous = this.availableReplicaCount.put(functionName, availableReplicas);
			if (previous != availableReplicas) {
				this.publisher.publish("riff_function_replicas",
						Collections.singletonMap(functionName, availableReplicas));
			}
		}
	}

	@EventListener
	public void onDeploymentUnregistered(ResourceDeletedEvent<Deployment> event) {
		Deployment deployment = event.getResource();
		if (deployment.getMetadata().getLabels().containsKey("function")) {
			String functionName = deployment.getMetadata().getName();
			this.actualReplicaCount.remove(functionName);
		}
	}

	@PreDestroy
	public synchronized void tearDown() {
		this.running.set(false);
		this.notify();
		this.lagTracker.close();
	}

	private void updateScalerInterval() {
		this.scalerIntervalMs = this.functions.isEmpty() ? 0 : 100;
	}

	private class Scaler implements Runnable {

		@Override
		public void run() {
			while (running.get()) {
				synchronized (FunctionMonitor.this) {
					Map<LagTracker.Subscription, List<LagTracker.Offsets>> offsets = lagTracker.compute();
					logOffsets(offsets);
					Map<String, Long> lags = offsets.entrySet().stream()
							.collect(Collectors.toMap(
									e -> e.getKey().group,
									e -> e.getValue().stream()
											.mapToLong(LagTracker.Offsets::getLag)
											.max().getAsLong()));
					functions.values().stream().forEach(
							f -> {
								int desired = computeDesiredReplicaCount(lags, f);
								String name = f.getMetadata().getName();
								Integer current = actualReplicaCount.computeIfAbsent(name, k -> 0);
								logger.debug("Want {} for {} (Deployment currently set to {})", desired, name, current);
								boolean resetScaleDownStartTime = true;
								if (desired < current) {
									if (scaleDownStartTimes.computeIfAbsent(name,
											n -> System.currentTimeMillis()) < System.currentTimeMillis()
													- SCALE_DOWN_LINGER_PERIOD) {
										deployer.deploy(f, desired);
										actualReplicaCount.put(name, desired);
									}
									else {
										resetScaleDownStartTime = false;
										logger.trace("Waiting another {}ms before scaling down to {} for {}",
												SCALE_DOWN_LINGER_PERIOD
														- (System.currentTimeMillis() - scaleDownStartTimes.get(name)),
												desired, name);
									}
								}
								else if (desired > current) {
									deployer.deploy(f, desired);
									actualReplicaCount.put(name, desired);
								}
								if (resetScaleDownStartTime) {
									scaleDownStartTimes.remove(name);
								}
							});
					try {
						FunctionMonitor.this.wait(scalerIntervalMs);
					}
					catch (InterruptedException e) {
						Thread.currentThread().interrupt();
					}
				}
			}
		}

		/**
		 * Compute the desired replica count for a function. This function leverages these 4 values (currently non configurable):
		 * <ul>
		 * <li>minReplicas (>= 0, default 0)</li>
		 * <li>maxReplicas (minReplicas <= maxReplicas <= partitionCount, default partitionCount)</li>
		 * <li>lagRequiredForOne, the amount of lag required to trigger the first pod to appear, default 1</li>
		 * <li>lagRequiredForMax, the amount of lag required to trigger all (maxReplicas) pods to appear. Default
		 * 10.</li>
		 * </ul>
		 * This method linearly interpolates based on witnessed lag and clamps the result between
		 * min/maxReplicas.
		 */
		private int computeDesiredReplicaCount(Map<String, Long> lags, XFunction f) {
			String fnName = f.getMetadata().getName();
			long lag = lags.get(fnName);
			String input = f.getSpec().getInput();

			// TODO: those 4 numbers part of Function spec?
			int lagRequiredForMax = 10;
			int lagRequiredForOne = 1;
			int minReplicas = 0;
			int maxReplicas = partitionCount(input);

			double slope = ((double) maxReplicas - 1) / (lagRequiredForMax - lagRequiredForOne);
			int computedReplicas;
			if (slope > 0d) {
				// max>1
				computedReplicas = (int) (1 + (lag - lagRequiredForOne) * slope);
			}
			else {
				computedReplicas = lag >= lagRequiredForOne ? 1 : 0;
			}
			return clamp(computedReplicas, minReplicas, maxReplicas);
		}

		private int partitionCount(String input) {
			waitForTopic(input);
			TopicSpec spec = topics.get(input).getSpec();
			if (spec == null || spec.getPartitions() == null) {
				return 1;
			}
			else {
				return spec.getPartitions();
			}
		}

		private void waitForTopic(String input) {
			try {
				topicsReady.computeIfAbsent(input, i -> new CountDownLatch(1)).await();
			}
			catch (InterruptedException e) {
				Thread.currentThread().interrupt();
			}
		}

		private void logOffsets(Map<LagTracker.Subscription, List<LagTracker.Offsets>> offsets) {
			for (Map.Entry<LagTracker.Subscription, List<LagTracker.Offsets>> entry : offsets
					.entrySet()) {
				logger.debug(entry.getKey().toString());
				for (LagTracker.Offsets values : entry.getValue()) {
					logger.debug("\t" + values + " Lag=" + values.getLag());
				}
			}
		}
	}

	private static int clamp(int value, int min, int max) {
		value = Math.min(value, max);
		value = Math.max(value, min);
		return value;
	}
}
