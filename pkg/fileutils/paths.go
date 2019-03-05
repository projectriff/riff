package fileutils

import (
	"fmt"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

func ResolveTilde(path string) (string, error) {
	if !StartsWithCurrentUserDirectoryAsTilde(path, runtime.GOOS) {
		return path, nil
	}
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	homeDirectory := currentUser.HomeDir
	if homeDirectory == "" {
		return "", fmt.Errorf("current user %s has no resolvable home directory", currentUser.Name)
	}
	return filepath.Join(homeDirectory, path[2:]), nil
}

func StartsWithCurrentUserDirectoryAsTilde(path string, os string) bool {
	if strings.HasPrefix(path, "~/") {
		return true
	}
	if os != "windows" {
		return false
	}
	return strings.HasPrefix(path, `~\`)
}
