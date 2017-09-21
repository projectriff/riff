
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

package io.sk8s.kubernetes.client;

import io.fabric8.kubernetes.client.Client;
import io.fabric8.kubernetes.client.dsl.MixedOperation;
import io.fabric8.kubernetes.client.dsl.Resource;
import io.sk8s.kubernetes.api.model.DoneableHandler;
import io.sk8s.kubernetes.api.model.DoneableTopic;
import io.sk8s.kubernetes.api.model.DoneableXFunction;
import io.sk8s.kubernetes.api.model.FunctionList;
import io.sk8s.kubernetes.api.model.Handler;
import io.sk8s.kubernetes.api.model.HandlerList;
import io.sk8s.kubernetes.api.model.Topic;
import io.sk8s.kubernetes.api.model.TopicList;
import io.sk8s.kubernetes.api.model.XFunction;

public interface Sk8sClient extends Client {

	MixedOperation<Topic, TopicList, DoneableTopic, Resource<Topic, DoneableTopic>> topics();

	MixedOperation<XFunction, FunctionList, DoneableXFunction, Resource<XFunction, DoneableXFunction>> functions();

	MixedOperation<Handler, HandlerList, DoneableHandler, Resource<Handler, DoneableHandler>> handlers();

}
