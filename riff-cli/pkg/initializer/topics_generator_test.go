/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package initializer

import (
	"testing"

	projectriff_v1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1"
	"github.com/projectriff/riff/riff-cli/pkg/options"
	"github.com/stretchr/testify/assert"
)

func TestTopicsYaml(t *testing.T) {
	as := assert.New(t)

	topicTemplate := projectriff_v1.Topic{}
	opts := options.InitOptions{
		FunctionName: "myfunc",
		Input:        "in",
	}
	yaml, err := createTopicsYaml(topicTemplate, opts)

	t.Log(yaml)

	as.NoError(err)
	as.Equal(yaml, `---
apiVersion: projectriff.io/v1
kind: Topic
metadata:
  name: in
`)
}
func TestTopicsYaml_WithPartitions(t *testing.T) {
	as := assert.New(t)

	var partitions = int32(5)
	topicTemplate := projectriff_v1.Topic{
		Spec: projectriff_v1.TopicSpec{
			Partitions: &partitions,
		},
	}
	opts := options.InitOptions{
		FunctionName: "myfunc",
		Input:        "in",
	}
	yaml, err := createTopicsYaml(topicTemplate, opts)

	t.Log(yaml)

	as.NoError(err)
	as.Equal(yaml, `---
apiVersion: projectriff.io/v1
kind: Topic
metadata:
  name: in
spec:
  partitions: 5
`)
}

func TestTopicsYaml_WithOutput(t *testing.T) {
	as := assert.New(t)

	topicTemplate := projectriff_v1.Topic{}
	opts := options.InitOptions{
		FunctionName: "myfunc",
		Input:        "in",
		Output:       "out",
	}
	yaml, err := createTopicsYaml(topicTemplate, opts)

	t.Log(yaml)

	as.NoError(err)
	as.Equal(yaml, `---
apiVersion: projectriff.io/v1
kind: Topic
metadata:
  name: in

---
apiVersion: projectriff.io/v1
kind: Topic
metadata:
  name: out
`)
}
