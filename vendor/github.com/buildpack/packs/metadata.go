package packs

const (
	BuildLabel     = "sh.packs.build"
	BuildpackLabel = "sh.packs.buildpacks"
)

type BuildMetadata struct {
	App        AppMetadata         `json:"app"`
	Config     ConfigMetadata      `json:"config"`
	Buildpacks []BuildpackMetadata `json:"buildpacks"`
	RunImage   RunImageMetadata    `json:"runimage"`
}

type BuildpackMetadata struct {
	Key     string                   `json:"key"`
	Name    string                   `json:"name"`
	Version string                   `json:"version,omitempty"`
	Layers  map[string]LayerMetadata `json:"layers,omitempty"`
}

type AppMetadata struct {
	Name string `json:"name"`
	SHA  string `json:"sha"`
}

type RunImageMetadata struct {
	Name string `json:"name"`
	SHA  string `json:"sha"`
}

type LayerMetadata struct {
	SHA  string      `json:"sha"`
	Data interface{} `json:"data"`
}

type ConfigMetadata struct {
	SHA string `json:"sha"`
}
