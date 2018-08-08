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
