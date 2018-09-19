package lifecycle

import (
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type BuildpackMap map[string]*Buildpack

type buildpackTOML struct {
	Buildpack struct {
		ID      string `toml:"id"`
		Version string `toml:"version"`
		Name    string `toml:"name"`
	} `toml:"buildpack"`
}

func NewBuildpackMap(dir string) (BuildpackMap, error) {
	buildpacks := BuildpackMap{}
	glob := filepath.Join(dir, "*", "*", "buildpack.toml")
	files, err := filepath.Glob(glob)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		buildpackDir := filepath.Dir(file)
		base, version := filepath.Split(buildpackDir)
		_, id := filepath.Split(filepath.Clean(base))
		var bpTOML buildpackTOML
		if _, err := toml.DecodeFile(file, &bpTOML); err != nil {
			return nil, err
		}
		buildpacks[id+"@"+version] = &Buildpack{
			ID:      bpTOML.Buildpack.ID,
			Version: bpTOML.Buildpack.Version,
			Name:    bpTOML.Buildpack.Name,
			Dir:     buildpackDir,
		}
	}
	return buildpacks, nil
}

func (m BuildpackMap) lookup(l []*Buildpack) []*Buildpack {
	out := make([]*Buildpack, 0, len(l))
	for _, b := range l {
		ref := b.ID + "@" + b.Version
		if b.Version == "" {
			ref += "latest"
		}
		if bp, ok := m[ref]; ok {
			bp := *bp
			bp.Optional = b.Optional
			out = append(out, &bp)
		}
	}
	return out
}

func (m BuildpackMap) ReadOrder(orderPath string) (BuildpackOrder, error) {
	var order struct {
		Groups BuildpackOrder `toml:"groups"`
	}
	if _, err := toml.DecodeFile(orderPath, &order); err != nil {
		return nil, err
	}

	var groups BuildpackOrder
	for _, g := range order.Groups {
		groups = append(groups, BuildpackGroup{
			Buildpacks: m.lookup(g.Buildpacks),
		})
	}
	return groups, nil
}

func (g *BuildpackGroup) Write(path string) error {
	data := struct {
		Buildpacks []*Buildpack `toml:"buildpacks"`
	}{
		Buildpacks: g.Buildpacks,
	}
	return WriteTOML(path, data)
}

func (m BuildpackMap) ReadGroup(path string) (*BuildpackGroup, error) {
	var group BuildpackGroup
	if _, err := toml.DecodeFile(path, &group); err != nil {
		return nil, err
	}
	group.Buildpacks = m.lookup(group.Buildpacks)
	return &group, nil
}
