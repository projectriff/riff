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

import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CountDownLatch;
import java.util.stream.Collectors;

import javax.annotation.PreDestroy;

import io.fabric8.kubernetes.api.model.extensions.Deployment;
import io.sk8s.core.resource.ResourceAddedOrModifiedEvent;
import io.sk8s.core.resource.ResourceDeletedEvent;
import io.sk8s.kubernetes.api.model.Topic;
import io.sk8s.kubernetes.api.model.TopicSpec;
import io.sk8s.kubernetes.api.model.XFunction;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.event.EventListener;

/**
 * Listens for function registration/un-registration and periodically scales deployment of
 * function pods based on the evaluation of a SpEL expression.
 *
 * @author Eric Bottard
 * @author Mark Fisher
 */
public class EventDispatchingHandler {

	private static final int DOWN_EDGE_LINGER_PERIOD = 10_000;

	private final Logger logger = LoggerFactory.getLogger(EventDispatchingHandler.class);

	// TODO: Change key to ObjectReference or similar for all these maps
	// Key is function name
	private final Map<String, XFunction> functions = new ConcurrentHashMap<>();

	private final Map<String, Topic> topics = new ConcurrentHashMap<>();

	private final Map<String, CountDownLatch> topicsReady = new ConcurrentHashMap<>();

	/** Keeps track of what the deployments ask. */
	private final Map<String, Integer> actualReplicaCount = new ConcurrentHashMap<>();

	/** Keeps track of the last time this decided to scale down (but not yet effective). */
	private final Map<String, Long> edgeInfos = new ConcurrentHashMap<>();

	@Autowired
	private FunctionDeployer deployer;

	private volatile long scalingMonitorIntervalMs = 0L; // Wait forever

	private final Thread scalingThread = new Thread(new ScalingMonitor());

	private volatile boolean running = false;

	private final KafkaConsumerMonitor kafkaConsumerMonitor = new KafkaConsumerMonitor(
			System.getenv("SPRING_CLOUD_STREAM_KAFKA_BINDER_BROKERS"));

	@EventListener
	public synchronized void onFunctionRegistered(ResourceAddedOrModifiedEvent<XFunction> event) {
		XFunction functionResource = event.getResource();
		String functionName = functionResource.getMetadata().getName();
		this.functions.put(functionName, functionResource);
		this.kafkaConsumerMonitor.beginTracking(functionName, functionResource.getSpec().getInput());

		updateMonitorInterval();
		if (!running) {
			running = true;
			scalingThread.start();
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

		deployer.undeploy(functionResource);

		this.kafkaConsumerMonitor.stopTracking(functionName, functionResource.getSpec().getInput());

		updateMonitorInterval();
		this.notify();

		logger.info("function deleted: " + functionName);
	}

	@EventListener
	public void onTopicRegistered(ResourceAddedOrModifiedEvent<Topic> event) {
		Topic topic = event.getResource();
		String name = topic.getMetadata().getName();
		topics.put(name, topic);
		topicsReady.computeIfAbsent(name, i -> new CountDownLatch(1)).countDown();
	}

	@EventListener
	public void onTopicUnregistered(ResourceDeletedEvent<Topic> event) {
		Topic topic = event.getResource();
		String name = topic.getMetadata().getName();
		topics.remove(name);
		topicsReady.remove(name);
	}

	@EventListener
	public void onDeploymentRegistered(ResourceAddedOrModifiedEvent<Deployment> event) {
		Deployment deployment = event.getResource();
		String functionName = deployment.getMetadata().getName();
		actualReplicaCount.put(functionName, deployment.getSpec().getReplicas());
	}

	@EventListener
	public void onDeploymentUnregistered(ResourceDeletedEvent<Deployment> event) {
		Deployment deployment = event.getResource();
		String functionName = deployment.getMetadata().getName();
		actualReplicaCount.remove(functionName);
	}

	@PreDestroy
	public synchronized void tearDown() {
		running = false;
		this.notify();
		kafkaConsumerMonitor.close();
	}

	private void updateMonitorInterval() {
		this.scalingMonitorIntervalMs = this.functions.isEmpty() ? 0 : 100;
	}

	private class ScalingMonitor implements Runnable {

		@Override
		public void run() {
			while (running) {
				synchronized (EventDispatchingHandler.this) {

					Map<KafkaConsumerMonitor.Subscription, List<KafkaConsumerMonitor.Offsets>> offsets = kafkaConsumerMonitor
							.compute();
					logOffsets(offsets);
					Map<String, Long> lags = offsets.entrySet().stream()
							.collect(Collectors.toMap(
									e -> e.getKey().group,
									e -> e.getValue().stream()
											.mapToLong(KafkaConsumerMonitor.Offsets::getLag)
											.max().getAsLong()));
					functions.values().stream().forEach(
							f -> {
								int desired = computeDesiredReplicaCount(lags, f);
								String name = f.getMetadata().getName();
								Integer currentDesired = actualReplicaCount.computeIfAbsent(name, k -> 0);
								logger.debug("Want {} for {} (Deployment currently set to {})", desired, name, currentDesired);
								if (desired < currentDesired) {
									if (edgeInfos.computeIfAbsent(name,
											n -> System.currentTimeMillis()) < System.currentTimeMillis()
													- DOWN_EDGE_LINGER_PERIOD) {
										deployer.deploy(f, desired);
										actualReplicaCount.put(name, desired);
										edgeInfos.remove(name);
									}
									else {
										logger.trace("Waiting another {}ms before scaling down to {} for {}",
												DOWN_EDGE_LINGER_PERIOD
														- (System.currentTimeMillis() - edgeInfos.get(name)),
												desired, name);
									}
								}
								else if (desired > currentDesired) {
									deployer.deploy(f, desired);
									actualReplicaCount.put(name, desired);
									edgeInfos.remove(name);
								}
								else {
									edgeInfos.remove(name);
								}
							});

					try {
						EventDispatchingHandler.this.wait(scalingMonitorIntervalMs);
					}
					catch (InterruptedException e) {
						Thread.currentThread().interrupt();
					}
				}
			}
		}

		/**
		 * Compute the desired nb of replicas for a function. This function leverages these 4 values (currently non configurable):
		 * <ul>
		 * <li>minReplicas (>= 0, default 0)</li>
		 * <li>maxReplicas (minReplicas <= maxReplicas <= nbPartitions, default nbPartitions)</li>
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
			int maxReplicas = nbPartitions(input);

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

		private int nbPartitions(String input) {
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

		private void logOffsets(Map<KafkaConsumerMonitor.Subscription, List<KafkaConsumerMonitor.Offsets>> offsets) {
			for (Map.Entry<KafkaConsumerMonitor.Subscription, List<KafkaConsumerMonitor.Offsets>> entry : offsets
					.entrySet()) {
				logger.debug(entry.getKey().toString());
				for (KafkaConsumerMonitor.Offsets values : entry.getValue()) {
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
