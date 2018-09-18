package lifecycle

const (
	MetadataLabel = "io.buildpacks.lifecycle.metadata"
)

type AppImageMetadata struct {
	App        AppMetadata         `json:"app"`
	Config     ConfigMetadata      `json:"config"`
	Buildpacks []BuildpackMetadata `json:"buildpacks"`
	RunImage   RunImageMetadata    `json:"runimage"`
}

type AppMetadata struct {
	SHA string `json:"sha"`
}

type ConfigMetadata struct {
	SHA string `json:"sha"`
}

type BuildpackMetadata struct {
	ID      string                   `json:"key"`
	Version string                   `json:"version"`
	Layers  map[string]LayerMetadata `json:"layers"`
}

type LayerMetadata struct {
	SHA  string      `json:"sha"`
	Data interface{} `json:"data"`
}

type RunImageMetadata struct {
	Name string `json:"name"`
	SHA  string `json:"sha"`
}
