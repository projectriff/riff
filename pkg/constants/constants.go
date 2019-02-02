/*
 * Copyright 2019 The original author or authors
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

package constants

import (
	"fmt"
)

var (
	// TODO update to a release version before releasing riff
	BuilderVersion  = "0.2.0-snapshot-ci-63cd05079e1f"
	BuilderName         = fmt.Sprintf("projectriff/builder:%s", BuilderVersion)
	DefaultRunImage = "packs/run:v3alpha2"

)
