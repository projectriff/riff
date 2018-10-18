package commands

import "os"

type Filesystem interface {
	MkdirAll(name string, perm os.FileMode) error
}

type LocalFs struct{}

func (LocalFs) MkdirAll(name string, perm os.FileMode) error {
	return os.MkdirAll(name, perm)
}
