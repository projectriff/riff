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

import java.util.List;

import io.sk8s.core.function.FunctionResource.FunctionSpec;
import io.sk8s.core.resource.Resource;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

/**
 * @author Mark Fisher
 */
// todo: add these to the model (can be empty)
@JsonIgnoreProperties({ "status", "message" })
public class FunctionResource extends Resource<FunctionSpec> {

	public class FunctionSpec {

		private String input;

		private String output;

		private String handler;

		private List<Param> params;

		private List<Env> env;

		public String getInput() {
			return input;
		}

		public void setInput(String topic) {
			this.input = topic;
		}

		public String getOutput() {
			return output;
		}

		public void setOutput(String output) {
			this.output = output;
		}

		public String getHandler() {
			return handler;
		}

		public void setHandler(String handler) {
			this.handler = handler;
		}

		public String getParam(String name) {
			for (Param param : params) {
				if (param.getName().equals(name)) {
					return param.getValue();
				}
			}
			return "";
		}

		public List<Param> getParams() {
			return params;
		}

		public void setParams(List<Param> params) {
			this.params = params;
		}

		public List<Env> getEnv() {
			return env;
		}

		public void setEnv(List<Env> env) {
			this.env = env;
		}

		@Override
		public String toString() {
			return "Spec [topic=" + input + ", handler=" + handler + "]";
		}
	}
}
