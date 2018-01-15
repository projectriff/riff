package osutils

import (
	"os"
	"path/filepath"
	"os/user"
	"fmt"
	"github.com/dturanski/riff-cli/pkg/ioutils"
)

func GetCWD() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return cwd
}

func GetCurrentBasePath() string {
	return filepath.Base(GetCWD())
}

func GetCurrentUsername() string {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	return user.Username
}

func FileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func IsDirectory(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		ioutils.Error(err)
		return false
	}
	return fi.Mode().IsDir()
}