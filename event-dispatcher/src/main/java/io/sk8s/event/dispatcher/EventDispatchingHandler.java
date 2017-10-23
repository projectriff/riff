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

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import javax.annotation.PostConstruct;
import javax.annotation.PreDestroy;

import io.sk8s.core.resource.ResourceAddedOrModifiedEvent;
import io.sk8s.core.resource.ResourceDeletedEvent;
import io.sk8s.kubernetes.api.model.XFunction;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.event.EventListener;
import org.springframework.expression.EvaluationContext;
import org.springframework.expression.Expression;
import org.springframework.expression.spel.standard.SpelExpressionParser;
import org.springframework.expression.spel.support.StandardEvaluationContext;

/**
 * Listens for function registration/un-registration and periodically scales deployment of function
 * pods based on the evaluation of a SpEL expression.
 *
 * @author Eric Bottard
 * @author Mark Fisher
 */
public class EventDispatchingHandler {

	private final Log logger = LogFactory.getLog(EventDispatchingHandler.class);

	// TODO: Change key to ObjectReference or similar
	private final Map<String, XFunction> functions = new HashMap<>();

	@Autowired
	private FunctionDeployer deployer;

	private volatile long scalingFrequencyMs = 0L; // Wait forever

	private final Thread scalingThread = new Thread(new ScalingMonitor());

	private final Object mutex = new Object();

	private volatile boolean running = false;

	private final KafkaConsumerMonitor kafkaConsumerMonitor = new KafkaConsumerMonitor(System.getenv("SPRING_CLOUD_STREAM_KAFKA_BINDER_BROKERS"));

	private final SpelExpressionParser spelExpressionParser = new SpelExpressionParser();

	@EventListener
	public void onFunctionRegistered(ResourceAddedOrModifiedEvent<XFunction> event) {
		XFunction functionResource = event.getResource();
		String functionName = functionResource.getMetadata().getName();
		this.functions.put(functionName, functionResource);
		this.kafkaConsumerMonitor.beginTracking(functionName, functionResource.getSpec().getInput());

		updateWakeUpInterval();
		if (!running) {
			running = true;
			scalingThread.start();
		}
		synchronized (mutex) {
			this.mutex.notify();
		}

		logger.info("function added: " + functionName);
	}


	@EventListener
	public void onFunctionUnregistered(ResourceDeletedEvent<XFunction> event) {
		XFunction functionResource = event.getResource();
		String functionName = functionResource.getMetadata().getName();
		this.functions.remove(functionName);
		deployer.undeploy(functionResource);

		this.kafkaConsumerMonitor.stopTracking(functionName, functionResource.getSpec().getInput());

		updateWakeUpInterval();
		synchronized (mutex) {
			this.mutex.notify();
		}

		logger.info("function deleted: " + functionName);
	}

	@PreDestroy
	public void tearDown() {
		running = false;
		synchronized (mutex) {
			mutex.notify();
		}
		kafkaConsumerMonitor.close();
	}

	/**
	 * Updates the wake up interval of the scaling thread to be the minimal desired value of all
	 * functions, or zero (sleep forever) if there is no function.
	 */
	private void updateWakeUpInterval() {
		this.scalingFrequencyMs = this.functions.values().stream()
				.mapToInt(f -> f.getSpec().getScalingStrategy() == null ? 1_000
						: f.getSpec().getScalingStrategy().getMaxUpdatePeriodMilliSeconds())
				.min().orElse(0);
	}

	private class ScalingMonitor implements Runnable {

		@Override
		public void run() {
			while (running) {

				Map<KafkaConsumerMonitor.TopicGroup, List<KafkaConsumerMonitor.Offsets>> offsets = kafkaConsumerMonitor.compute();
				logOffsets(offsets);
				Map<String, Long> lags = offsets.entrySet().stream()
						.collect(Collectors.toMap(
							e -> e.getKey().group,
							e -> e.getValue().stream()
									.mapToLong(KafkaConsumerMonitor.Offsets::getLag)
									.max().getAsLong()
						));
				functions.values().parallelStream().forEach(
						f -> {
							String expression = f.getSpec().getScalingStrategy() != null ? f.getSpec().getScalingStrategy().getReplicasExpression() : "1";
							Expression spel = spelExpressionParser.parseExpression(expression);
							EvaluationContext context = new StandardEvaluationContext();
							context.setVariable("lag", lags.get(f.getMetadata().getName()));
							int replicas = spel.getValue(context, Integer.class);
							// TODO no more replicas than topic.partitions
							deployer.deploy(f, replicas);
						}
				);
				
				synchronized (mutex) {
					try {
						mutex.wait(scalingFrequencyMs);
					}
					catch (InterruptedException e) {
						Thread.currentThread().interrupt();
					}
				}
			}
		}

		private void logOffsets(Map<KafkaConsumerMonitor.TopicGroup, List<KafkaConsumerMonitor.Offsets>> offsets) {
			for (Map.Entry<KafkaConsumerMonitor.TopicGroup, List<KafkaConsumerMonitor.Offsets>> entry : offsets.entrySet()) {
				System.out.println(entry.getKey());
				for (KafkaConsumerMonitor.Offsets values : entry.getValue()) {
					System.out.println("\t" + values + " Lag=" + values.getLag());
				}
			}
		}
	}
}
