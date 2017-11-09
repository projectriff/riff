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

package io.sk8s.topic.gateway;

import java.util.Map;

import org.springframework.beans.BeansException;
import org.springframework.beans.DirectFieldAccessor;
import org.springframework.cloud.stream.binder.BinderFactory;
import org.springframework.cloud.stream.binder.ConsumerProperties;
import org.springframework.cloud.stream.binder.HeaderMode;
import org.springframework.cloud.stream.binder.ProducerProperties;
import org.springframework.cloud.stream.binding.BinderAwareChannelResolver;
import org.springframework.cloud.stream.binding.BindingService;
import org.springframework.cloud.stream.config.BinderProperties;
import org.springframework.cloud.stream.config.BindingProperties;
import org.springframework.cloud.stream.config.BindingServiceProperties;
import org.springframework.context.ApplicationContext;
import org.springframework.core.convert.ConversionService;
import org.springframework.messaging.MessageChannel;
import org.springframework.messaging.core.DestinationResolver;

/**
 * Extends {@link BinderAwareChannelResolver} to add the "embedded" header-mode property.
 *
 * @author Mark Fisher
 */
public class HeaderEmbeddingBinderAwareChannelResolver implements DestinationResolver<MessageChannel> {

	private final BinderAwareChannelResolver channelResolver;

	public HeaderEmbeddingBinderAwareChannelResolver(BinderFactory binderFactory, BinderAwareChannelResolver channelResolver) {
		this.channelResolver = channelResolver;
		DirectFieldAccessor accessor = new DirectFieldAccessor(channelResolver);
		BindingService bindingService = (BindingService) accessor.getPropertyValue("bindingService");
		accessor.setPropertyValue("bindingService",
				new HeaderEmbeddingBindingService(bindingService.getBindingServiceProperties(), binderFactory));
	}

	@Override
	public MessageChannel resolveDestination(String channelName) {
		return this.channelResolver.resolveDestination(channelName);
	}

	private static class HeaderEmbeddingBindingService extends BindingService {

		public HeaderEmbeddingBindingService(BindingServiceProperties bindingServiceProperties,
				BinderFactory binderFactory) {
			super(new HeaderEmbeddingBindingServiceProperties(bindingServiceProperties), binderFactory);
		}
	}

	private static class HeaderEmbeddingBindingServiceProperties extends BindingServiceProperties {

		private final BindingServiceProperties bindingServiceProperties;

		private HeaderEmbeddingBindingServiceProperties(BindingServiceProperties bindingServiceProperties) {
			this.bindingServiceProperties = bindingServiceProperties;
		}

		@Override
		public Map<String, BindingProperties> getBindings() {
			return this.bindingServiceProperties.getBindings();
		}

		@Override
		public void setBindings(Map<String, BindingProperties> bindings) {
			this.bindingServiceProperties.setBindings(bindings);
		}

		@Override
		public Map<String, BinderProperties> getBinders() {
			return this.bindingServiceProperties.getBinders();
		}

		@Override
		public void setBinders(Map<String, BinderProperties> binders) {
			this.bindingServiceProperties.setBinders(binders);
		}

		@Override
		public String getDefaultBinder() {
			return this.bindingServiceProperties.getDefaultBinder();
		}

		@Override
		public void setDefaultBinder(String defaultBinder) {
			this.bindingServiceProperties.setDefaultBinder(defaultBinder);
		}

		@Override
		public int getInstanceIndex() {
			return this.bindingServiceProperties.getInstanceIndex();
		}

		@Override
		public void setInstanceIndex(int instanceIndex) {
			this.bindingServiceProperties.setInstanceIndex(instanceIndex);
		}

		@Override
		public int getInstanceCount() {
			return this.bindingServiceProperties.getInstanceCount();
		}

		@Override
		public void setInstanceCount(int instanceCount) {
			this.bindingServiceProperties.setInstanceCount(instanceCount);
		}

		@Override
		public String[] getDynamicDestinations() {
			return this.bindingServiceProperties.getDynamicDestinations();
		}

		@Override
		public void setDynamicDestinations(String[] dynamicDestinations) {
			this.bindingServiceProperties.setDynamicDestinations(dynamicDestinations);
		}

		@Override
		public void setApplicationContext(ApplicationContext applicationContext) throws BeansException {
			this.bindingServiceProperties.setApplicationContext(applicationContext);
		}

		@Override
		public void setConversionService(ConversionService conversionService) {
			this.bindingServiceProperties.setConversionService(conversionService);
		}

		@Override
		public void afterPropertiesSet() throws Exception {
			this.bindingServiceProperties.afterPropertiesSet();
		}

		@Override
		public String getBinder(String bindingName) {
			return this.bindingServiceProperties.getBinder(bindingName);
		}

		@Override
		public Map<String, Object> asMapProperties() {
			return this.bindingServiceProperties.asMapProperties();
		}

		@Override
		public ConsumerProperties getConsumerProperties(String inputBindingName) {
			return this.bindingServiceProperties.getConsumerProperties(inputBindingName);
		}

		@Override
		public BindingProperties getBindingProperties(String bindingName) {
			return this.bindingServiceProperties.getBindingProperties(bindingName);
		}

		@Override
		public String getGroup(String bindingName) {
			return this.bindingServiceProperties.getGroup(bindingName);
		}

		@Override
		public String getBindingDestination(String bindingName) {
			return this.bindingServiceProperties.getBindingDestination(bindingName);
		}

		@Override
		public int hashCode() {
			return this.bindingServiceProperties.hashCode();
		}

		@Override
		public boolean equals(Object obj) {
			return this.bindingServiceProperties.equals(obj);
		}

		@Override
		public String toString() {
			return this.bindingServiceProperties.toString();
		}

		@Override
		public ProducerProperties getProducerProperties(String outputBindingName) {
			ProducerProperties producerProperties = this.bindingServiceProperties.getProducerProperties(outputBindingName);
			producerProperties.setHeaderMode(HeaderMode.embeddedHeaders);
			return producerProperties;
		}
	}
}
