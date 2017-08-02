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

package io.sk8s.core.function;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

import io.sk8s.core.function.FunctionResource.FunctionSpec;
import io.sk8s.core.resource.Resource;

/**
 * @author Mark Fisher
 */
// todo: add these to the model (can be empty)
@JsonIgnoreProperties({ "status", "message" })
public class FunctionResource extends Resource<FunctionSpec> {

	public class FunctionSpec {

		private String topic;

		private String lambda;

		public String getTopic() {
			return topic;
		}

		public void setTopic(String topic) {
			this.topic = topic;
		}

		public String getLambda() {
			return lambda;
		}

		public void setLambda(String lambda) {
			this.lambda = lambda;
		}

		@Override
		public String toString() {
			return "Spec [topic=" + topic + ", lambda=" + lambda + "]";
		}
	}
}
