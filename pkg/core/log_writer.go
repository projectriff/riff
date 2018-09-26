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

package core

import (
	"fmt"
	"io"

	"github.com/boz/kail"
)

type Writer interface {
	Print(event kail.Event) error
	Fprint(w io.Writer, event kail.Event) error
}

func NewWriter(out io.Writer) Writer {
	return &writer{out}
}

type writer struct {
	out io.Writer
}

func (w *writer) Print(ev kail.Event) error {
	return w.Fprint(w.out, ev)
}

func (w *writer) Fprint(out io.Writer, ev kail.Event) error {
	if _, err := out.Write(prefix(ev)); err != nil {
		return err
	}

	log := ev.Log()

	if _, err := out.Write(log); err != nil {
		return err
	}

	if sz := len(log); sz == 0 || log[sz-1] != byte('\n') {
		if _, err := out.Write([]byte("\n")); err != nil {
			return err
		}
	}
	return nil
}

func prefix(ev kail.Event) []byte {
	return []byte(fmt.Sprintf("%v/%v[%v]: ",
		ev.Source().Namespace(),
		ev.Source().Name(),
		ev.Source().Container()))
}
