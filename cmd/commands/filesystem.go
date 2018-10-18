package commands

import "os"

type Filesystem interface {
	Mkdir(name string, perm os.FileMode) error
	Stat(name string) (os.FileInfo, error)
}

type LocalFs struct{}

func (LocalFs) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}
func (LocalFs) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
