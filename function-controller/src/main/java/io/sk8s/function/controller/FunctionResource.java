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

import java.util.Map;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

/**
 * @author Mark Fisher
 */
// todo: add these to the model (can be empty)
@JsonIgnoreProperties({ "status", "message" })
public class FunctionResource {

	private String apiVersion;

	private String kind;

	private Map<String, String> metadata;

	private Spec spec;

	public String getApiVersion() {
		return apiVersion;
	}


	public void setApiVersion(String apiVersion) {
		this.apiVersion = apiVersion;
	}

	public String getKind() {
		return kind;
	}

	public void setKind(String kind) {
		this.kind = kind;
	}

	public Map<String, String> getMetadata() {
		return metadata;
	}

	public void setMetadata(Map<String, String> metadata) {
		this.metadata = metadata;
	}

	public Spec getSpec() {
		return spec;
	}

	public void setSpec(Spec spec) {
		this.spec = spec;
	}

	@Override
	public String toString() {
		return "FunctionResource [apiVersion=" + apiVersion + ", kind=" + kind + ", metadata=" + metadata + ", spec="
				+ spec + "]";
	}

	public class Spec {

		private String lambda;

		public String getLambda() {
			return lambda;
		}

		public void setLambda(String lambda) {
			this.lambda = lambda;
		}

		@Override
		public String toString() {
			return "Spec [lambda=" + lambda + "]";
		}
	}
}
