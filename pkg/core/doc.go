/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// package core contains a high level client for interacting with knative primitives used to realise, amongst other
// things, riff functions. The client typically returns structs for kubernetes resources, as they would have been
// written "by hand" to be applied by kubectl. It is not the role of the client to deal with printing, formatting or
// command line parsing.
package core
