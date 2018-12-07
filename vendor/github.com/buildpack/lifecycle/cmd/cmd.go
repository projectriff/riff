package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	DefaultLaunchDir      = "/workspace"
	DefaultAppDir         = "/workspace/app"
	DefaultCacheDir       = "/cache"
	DefaultBuildpacksDir  = "/buildpacks"
	DefaultPlatformDir    = "/platform"
	DefaultOrderPath      = "/buildpacks/order.toml"
	DefaultGroupPath      = `./group.toml`
	DefaultPlanPath       = "./plan.toml"
	DefaultUseDaemon      = false
	DefaultUseCredHelpers = false

	EnvRunImage = "PACK_RUN_IMAGE"
	EnvUID      = "PACK_USER_ID"
	EnvGID      = "PACK_GROUP_ID"
)

func FlagLaunchDir(dir *string) {
	flag.StringVar(dir, "launch", DefaultLaunchDir, "path to launch directory")
}

func FlagLaunchDirSrc(dir *string) {
	flag.StringVar(dir, "launch-src", DefaultLaunchDir, "path to source launch directory for export step")
}

func FlagDryRunDir(dir *string) {
	flag.StringVar(dir, "dry-run", "", "path to store first stage output in (Don't perform export)")
}

func FlagAppDir(dir *string) {
	flag.StringVar(dir, "app", DefaultAppDir, "path to app directory")
}

func FlagAppDirSrc(dir *string) {
	flag.StringVar(dir, "app-src", DefaultAppDir, "path to app directory for export step")
}

func FlagCacheDir(dir *string) {
	flag.StringVar(dir, "cache", DefaultCacheDir, "path to cache directory")
}

func FlagBuildpacksDir(dir *string) {
	flag.StringVar(dir, "buildpacks", DefaultBuildpacksDir, "path to buildpacks directory")
}

func FlagPlatformDir(dir *string) {
	flag.StringVar(dir, "platform", DefaultPlatformDir, "path to platform directory")
}

func FlagOrderPath(path *string) {
	flag.StringVar(path, "order", DefaultOrderPath, "path to order.toml")
}

func FlagGroupPath(path *string) {
	flag.StringVar(path, "group", DefaultGroupPath, "path to group.toml")
}

func FlagPlanPath(path *string) {
	flag.StringVar(path, "plan", DefaultPlanPath, "path to plan.toml")
}

func FlagRunImage(image *string) {
	flag.StringVar(image, "image", os.Getenv(EnvRunImage), "reference to run image")
}

func FlagMetadataPath(metadata *string) {
	flag.StringVar(metadata, "metadata", "", "path to json containing image metadata for previous image")
}

func FlagUseDaemon(use *bool) {
	flag.BoolVar(use, "daemon", DefaultUseDaemon, "export to docker daemon")
}

func FlagUseCredHelpers(use *bool) {
	flag.BoolVar(use, "helpers", DefaultUseCredHelpers, "use credential helpers")
}

func FlagUID(uid *int) {
	flag.IntVar(uid, "uid", intEnv(EnvUID), "UID of user in the stack's build and run images")
}

func FlagGID(gid *int) {
	flag.IntVar(gid, "gid", intEnv(EnvGID), "GID of user's group in the stack's build and run images")
}

const (
	CodeFailed      = 1
	CodeInvalidArgs = iota + 2
	CodeInvalidEnv
	CodeNotFound
	CodeFailedDetect
	CodeFailedBuild
	CodeFailedLaunch
	CodeFailedUpdate
)

type ErrorFail struct {
	Err    error
	Code   int
	Action []string
}

func (e *ErrorFail) Error() string {
	message := "failed to " + strings.Join(e.Action, " ")
	if e.Err == nil {
		return message
	}
	return fmt.Sprintf("%s: %s", message, e.Err)
}

func FailCode(code int, action ...string) error {
	return FailErrCode(nil, code, action...)
}

func FailErr(err error, action ...string) error {
	code := CodeFailed
	if err, ok := err.(*ErrorFail); ok {
		code = err.Code
	}
	return FailErrCode(err, code, action...)
}

func FailErrCode(err error, code int, action ...string) error {
	return &ErrorFail{Err: err, Code: code, Action: action}
}

func Exit(err error) {
	if err == nil {
		os.Exit(0)
	}
	log.Printf("Error: %s\n", err)
	if err, ok := err.(*ErrorFail); ok {
		os.Exit(err.Code)
	}
	os.Exit(CodeFailed)
}

func intEnv(k string) int {
	v := os.Getenv(k)
	d, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return d
}
