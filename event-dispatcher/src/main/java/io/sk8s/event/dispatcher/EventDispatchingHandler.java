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
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

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

	public static final int DOWN_EDGE_LINGER_PERIOD = 10_000;

	private final Log logger = LogFactory.getLog(EventDispatchingHandler.class);

	// TODO: Change key to ObjectReference or similar
	private final Map<String, XFunction> functions = new ConcurrentHashMap<>();

	// TODO: Change key to ObjectReference or similar
	private final Map<String, Topic> topics = new ConcurrentHashMap<>();

	private final Map<String, CountDownLatch> topicsReady = new ConcurrentHashMap<>();

	private final Map<String, Integer> actualNbReplicas = new ConcurrentHashMap<>();

	private final Map<String, Long> edgeInfos = new ConcurrentHashMap<>();

	@Autowired
	private FunctionDeployer deployer;

	private volatile long scalingFrequencyMs = 0L; // Wait forever

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

		updateWakeUpInterval();
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
		this.actualNbReplicas.remove(functionName);

		deployer.undeploy(functionResource);

		this.kafkaConsumerMonitor.stopTracking(functionName, functionResource.getSpec().getInput());

		updateWakeUpInterval();
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
		actualNbReplicas.put(functionName, deployment.getSpec().getReplicas());
	}

	@EventListener
	public void onDeploymentUnregistered(ResourceDeletedEvent<Deployment> event) {
		Deployment deployment = event.getResource();
		String functionName = deployment.getMetadata().getName();
		actualNbReplicas.remove(functionName);
	}

	@PreDestroy
	public synchronized void tearDown() {
		running = false;
		this.notify();
		kafkaConsumerMonitor.close();
	}

	private void updateWakeUpInterval() {
		this.scalingFrequencyMs = this.functions.isEmpty() ? 0 : 100;
	}

	private class ScalingMonitor implements Runnable {

		@Override
		public void run() {
			while (running) {
				synchronized (EventDispatchingHandler.this) {

					Map<KafkaConsumerMonitor.TopicGroup, List<KafkaConsumerMonitor.Offsets>> offsets = kafkaConsumerMonitor
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
								int desired = computeDesiredNbReplicas(lags, f);
								String name = f.getMetadata().getName();
								Integer currentDesired = actualNbReplicas.computeIfAbsent(name, k -> desired);
								System.out.println("Want " + desired + " for " + name);
								if (desired < currentDesired) {
									if (edgeInfos.computeIfAbsent(name,
											n -> System.currentTimeMillis()) < System.currentTimeMillis()
													- DOWN_EDGE_LINGER_PERIOD) {
										deployer.deploy(f, desired);
										actualNbReplicas.put(name, desired);
										edgeInfos.remove(name);
									}
									else {
										System.out.printf("Waiting another %dms before scaling down to %d for %s%n",
												DOWN_EDGE_LINGER_PERIOD
														- (System.currentTimeMillis() - edgeInfos.get(name)),
												desired, name);
									}
								}
								else if (desired > currentDesired) {
									deployer.deploy(f, desired);
									actualNbReplicas.put(name, desired);
									edgeInfos.remove(name);
								}
								else {
									edgeInfos.remove(name);
								}
							});

					try {
						EventDispatchingHandler.this.wait(scalingFrequencyMs);
					}
					catch (InterruptedException e) {
						Thread.currentThread().interrupt();
					}
				}
			}
		}

		/**
		 * Compute the desired nb of replicas for a function. Each function defines 4 values:
		 * <ul>
		 * <li>minReplicas (>= 0, default 0)</li>
		 * <li>maxReplicas (minReplicas <= maxReplicas <= nbPartitions, default nbPartitions)</li>
		 * <li>l, the amount of lag required to trigger the first pod to appear, default 1</li>
		 * <li>L, the amount of lag required to trigger all (maxReplicas) pods to appear. Default
		 * 10.</li>
		 * </ul>
		 * This method linearly interpolates based on witnessed lag and clamps the result between
		 * min/maxReplicas.
		 */
		private int computeDesiredNbReplicas(Map<String, Long> lags, XFunction f) {
			String fnName = f.getMetadata().getName();
			long lag = lags.get(fnName);

			// TODO: those 3 numbers part of Function spec?
			int bigL = 10;
			int littleL = 1;
			int min = 0;

			String input = f.getSpec().getInput();
			double max = nbPartitions(input);
			double slope = (max - 1) / (bigL - littleL);
			int computedReplicas;
			if (slope > 0d) {
				// max>1
				computedReplicas = (int) (1 + (lag - littleL) * slope);
			}
			else {
				computedReplicas = lag >= littleL ? 1 : 0;
			}
			return clamp(computedReplicas, min, (int) max);
		}

		private int nbPartitions(String input) {
			waitForTopicEntry(input);
			TopicSpec spec = topics.get(input).getSpec();
			if (spec == null || spec.getPartitions() == null) {
				return 1;
			}
			else {
				return spec.getPartitions();
			}
		}

		private void waitForTopicEntry(String input) {
			try {
				topicsReady.computeIfAbsent(input, i -> new CountDownLatch(1)).await();
			}
			catch (InterruptedException e) {
				Thread.currentThread().interrupt();
			}
		}

		private void logOffsets(Map<KafkaConsumerMonitor.TopicGroup, List<KafkaConsumerMonitor.Offsets>> offsets) {
			for (Map.Entry<KafkaConsumerMonitor.TopicGroup, List<KafkaConsumerMonitor.Offsets>> entry : offsets
					.entrySet()) {
				System.out.println(entry.getKey());
				for (KafkaConsumerMonitor.Offsets values : entry.getValue()) {
					System.out.println("\t" + values + " Lag=" + values.getLag());
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
