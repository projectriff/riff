/*
 * Copyright 2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cli

import (
	"github.com/ghodss/yaml"
	"github.com/projectriff/system/pkg/apis"
)

func PrintResourceStatus(c *Config, name string, condition *apis.Condition) {
	c.Printf("# %s: %s\n", name, FormatConditionStatus(condition))
	if condition != nil {
		s, _ := yaml.Marshal(condition)
		c.Printf("---\n")
		c.Printf("%s", string(s))
	}
}
