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

package io.sk8s.core.handler;

import java.util.List;

import io.sk8s.core.handler.HandlerResource.HandlerSpec;
import io.sk8s.core.resource.Resource;

/**
 * @author Mark Fisher
 */
public class HandlerResource extends Resource<HandlerSpec> {

	public class HandlerSpec {

		private String dispatcher;

		private int replicas = 1;

		private String image;

		private String command;

		private List<String> args;

		public String getDispatcher() {
			return dispatcher;
		}

		public void setDispatcher(String dispatcher) {
			this.dispatcher = dispatcher;
		}

		public int getReplicas() {
			return replicas;
		}

		public void setReplicas(int replicas) {
			this.replicas = replicas;
		}

		public String getImage() {
			return image;
		}

		public void setImage(String image) {
			this.image = image;
		}

		public String getCommand() {
			return command;
		}

		public void setCommand(String command) {
			this.command = command;
		}

		public List<String> getArgs() {
			return args;
		}

		public void setArgs(List<String> args) {
			this.args = args;
		}
	}
}
