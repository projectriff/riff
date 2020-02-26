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

package pack

import (
	"context"
	"io"

	"github.com/buildpacks/pack"
	"github.com/buildpacks/pack/logging"
)

// Client exposes methods on pack.Client. Its only purpose is to decouple riff from the
// pack.Client struct at test time.
type Client interface {
	Build(ctx context.Context, opts pack.BuildOptions) error
}

func NewClient(stdout io.Writer) (Client, error) {
	logger := logging.New(stdout)
	return pack.NewClient(pack.WithLogger(logger))
}
