package lifecycle

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func SetupKnativeLaunchDir(dir string) error {
	var ephemeralApp string
	if _, err := os.Stat(filepath.Join(dir, "app")); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		ephemeralApp = fmt.Sprintf("app-%d", time.Now().Unix())
		if err := os.Rename(
			filepath.Join(dir, "app"),
			filepath.Join(dir, ephemeralApp),
		); err != nil {
			return err
		}
	}
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return err
	}
	if err := os.Mkdir(filepath.Join(dir, "app"), 0755); err != nil {
		return err
	}
	for _, file := range files {
		if err := os.Rename(file, filepath.Join(dir, "app", filepath.Base(file))); err != nil {
			return err
		}
	}

	if ephemeralApp != "" {
		if err := os.Rename(
			filepath.Join(dir, "app", ephemeralApp),
			filepath.Join(dir, "app", "app"),
		); err != nil {
			return err
		}
	}
	return nil
}

func ChownDirs(launchDir, homeDir, cacheDir string, uid, gid int) error {
	fmt.Println("chowning docker config.json to pack:pack")
	err := os.Chown(filepath.Join(homeDir, ".docker", "config.json"), uid, gid)
	if err != nil {
		return err
	}

	fmt.Println("chowning /workspace to pack:pack")
	if err := filepath.Walk(launchDir, func(path string, info os.FileInfo, err error) error {
		return os.Chown(path, uid, gid)
	}); err != nil {
		return err
	}

	if _, err := os.Stat(cacheDir); err == nil {
		fmt.Println("chowning /cache to pack:pack")
		return filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
			return os.Chown(path, uid, gid)
		})
	}
	return nil
}
