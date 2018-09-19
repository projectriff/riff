package packs

import (
	"flag"
	"os"
)

const (
	EnvAppDir = "PACK_APP_DIR"
	EnvAppZip = "PACK_APP_ZIP"

	EnvAppName = "PACK_APP_NAME"
	EnvAppURI  = "PACK_APP_URI"

	EnvAppDisk   = "PACK_APP_DISK"
	EnvAppMemory = "PACK_APP_MEM"
	EnvAppFds    = "PACK_APP_FDS"

	EnvDropletPath    = "PACK_DROPLET_PATH"
	EnvSlugPath       = "PACK_SLUG_PATH"
	EnvMetadataPath   = "PACK_METADATA_PATH"
	EnvBPPath         = "PACK_BP_PATH"
	EnvBPOrderPath    = "PACK_BP_ORDER_PATH"
	EnvBPGroupPath    = "PACK_BP_GROUP_PATH"
	EnvDetectInfoPath = "PACK_DETECT_INFO_PATH"

	EnvRunImage   = "PACK_RUN_IMAGE"
	EnvStackName  = "PACK_STACK_NAME"
	EnvUseDaemon  = "PACK_USE_DAEMON"
	EnvUseHelpers = "PACK_USE_HELPERS"
)

func InputDropletPath(path *string) {
	flag.StringVar(path, "droplet", os.Getenv(EnvDropletPath), "file containing droplet")
}

func InputSlugPath(path *string) {
	flag.StringVar(path, "slug", os.Getenv(EnvSlugPath), "file containing slug")
}

func InputMetadataPath(path *string) {
	flag.StringVar(path, "metadata", os.Getenv(EnvMetadataPath), "file containing artifact metadata")
}

func InputBPPath(path *string) {
	flag.StringVar(path, "buildpacks", os.Getenv(EnvBPPath), "directory containing buildpacks")
}

func InputBPOrderPath(path *string) {
	flag.StringVar(path, "order", os.Getenv(EnvBPOrderPath), "file containing detection order")
}

func InputBPGroupPath(path *string) {
	flag.StringVar(path, "group", os.Getenv(EnvBPGroupPath), "file containing a buildpack group")
}

func InputDetectInfoPath(path *string) {
	flag.StringVar(path, "info", os.Getenv(EnvDetectInfoPath), "file containing detection info")
}

func InputRunImage(image *string) {
	flag.StringVar(image, "run-image", os.Getenv(EnvRunImage), "image repository containing run image")
}

func InputStackName(image *string) {
	flag.StringVar(image, "stack", os.Getenv(EnvStackName), "image repository containing stack image")
}

func InputUseDaemon(use *bool) {
	flag.BoolVar(use, "daemon", BoolEnv(EnvUseDaemon), "export to docker daemon")
}

func InputUseHelpers(use *bool) {
	flag.BoolVar(use, "helpers", BoolEnv(EnvUseHelpers), "use credential helpers")
}
