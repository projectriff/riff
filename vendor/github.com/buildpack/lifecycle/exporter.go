package lifecycle

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/pkg/errors"

	"github.com/buildpack/lifecycle/img"
)

type Exporter struct {
	Buildpacks []*Buildpack
	TmpDir     string
	In         []byte
	Out, Err   io.Writer
}

func (e *Exporter) Export(launchDir string, runImage, origImage v1.Image) (v1.Image, error) {
	runImageDigest, err := runImage.Digest()
	if err != nil {
		return nil, errors.Wrap(err, "find run image digest")
	}
	metadata := AppImageMetadata{
		RunImage: RunImageMetadata{
			SHA: runImageDigest.String(),
		},
	}

	repoImage, topLayerDigest, err := e.addDirAsLayer(runImage, filepath.Join(e.TmpDir, "app.tgz"), filepath.Join(launchDir, "app"), "workspace/app")
	if err != nil {
		return nil, errors.Wrap(err, "append app layer to run image")
	}
	metadata.App.SHA = topLayerDigest

	repoImage, topLayerDigest, err = e.addDirAsLayer(repoImage, filepath.Join(e.TmpDir, "config.tgz"), filepath.Join(launchDir, "config"), "workspace/config")
	if err != nil {
		return nil, errors.Wrap(err, "append config layer to run image")
	}
	metadata.Config.SHA = topLayerDigest

	var bpMetadata []BuildpackMetadata
	if origImage != nil {
		data, err := e.GetMetadata(origImage)
		if err != nil {
			return nil, errors.Wrap(err, "find metadata")
		}
		bpMetadata = data.Buildpacks
	}
	for _, buildpack := range e.Buildpacks {
		var origLayers map[string]LayerMetadata
		for _, md := range bpMetadata {
			if md.ID == buildpack.ID {
				origLayers = md.Layers
			}
		}
		bpMetadata := BuildpackMetadata{ID: buildpack.ID, Version: buildpack.Version}
		repoImage, bpMetadata.Layers, err = e.addBuildpackLayer(buildpack.ID, launchDir, repoImage, origImage, origLayers)
		if err != nil {
			return nil, errors.Wrap(err, "append layers")
		}
		metadata.Buildpacks = append(metadata.Buildpacks, bpMetadata)
	}

	buildJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, errors.Wrap(err, "get encoded metadata")
	}
	repoImage, err = img.Label(repoImage, MetadataLabel, string(buildJSON))
	if err != nil {
		return nil, errors.Wrap(err, "set metadata label")
	}

	return repoImage, nil
}

func (e *Exporter) addBuildpackLayer(id, launchDir string, repoImage, origImage v1.Image, origLayers map[string]LayerMetadata) (v1.Image, map[string]LayerMetadata, error) {
	metadata := map[string]LayerMetadata{}
	layers, err := filepath.Glob(filepath.Join(launchDir, id, "*.toml"))
	if err != nil {
		return nil, nil, err
	}
	for _, layer := range layers {
		if filepath.Base(layer) == "launch.toml" {
			continue
		}
		var layerDiffID string
		dir := strings.TrimSuffix(layer, ".toml")
		layerName := filepath.Base(dir)
		dirInfo, err := os.Stat(dir)
		if os.IsNotExist(err) {
			if origImage == nil || origLayers == nil || origLayers[layerName].SHA == "" {
				return nil, nil, errors.Errorf("layer TOML found, but no available contents for %s %s", id, layerName)
			}
			layerDiffID = origLayers[layerName].SHA
			hash, err := v1.NewHash(layerDiffID)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "parse hash: %s", origLayers[layerName].SHA)
			}
			topLayer, err := origImage.LayerByDiffID(hash)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "find previous layer %s/%s", id, layerName)
			}
			repoImage, err = mutate.AppendLayers(repoImage, topLayer)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "append layer %s/%s from previous image", id, layerName)
			}
		} else if err != nil {
			return nil, nil, err
		} else if !dirInfo.IsDir() {
			return nil, nil, errors.Errorf("expected %s to be a directory", dir)
		} else {
			tarFile := filepath.Join(e.TmpDir, fmt.Sprintf("layer.%s.%s.tgz", id, layerName))
			repoImage, layerDiffID, err = e.addDirAsLayer(repoImage, tarFile, dir, filepath.Join("workspace", id, layerName))
			if err != nil {
				return nil, nil, errors.Wrap(err, "append dir as layer")
			}
		}
		var tomlData map[string]interface{}
		if _, err := toml.DecodeFile(layer, &tomlData); err != nil {
			return nil, nil, errors.Wrap(err, "read layer TOML data")
		}
		metadata[layerName] = LayerMetadata{SHA: layerDiffID, Data: tomlData}
	}
	return repoImage, metadata, nil
}

func (e *Exporter) GetMetadata(image v1.Image) (AppImageMetadata, error) {
	var metadata AppImageMetadata
	cfg, err := image.ConfigFile()
	if err != nil {
		return metadata, err
	}
	label := cfg.Config.Labels[MetadataLabel]
	if err := json.Unmarshal([]byte(label), &metadata); err != nil {
		return metadata, err
	}
	return metadata, nil
}

func (e *Exporter) addDirAsLayer(image v1.Image, tarFile, fsDir, tarDir string) (v1.Image, string, error) {
	if err := e.createTarFile(tarFile, fsDir, tarDir); err != nil {
		return nil, "", errors.Wrapf(err, "tar %s to %s", fsDir, tarFile)
	}
	newImage, topLayer, err := img.Append(image, tarFile)
	if err != nil {
		return nil, "", errors.Wrap(err, "append layers to run image")
	}
	diffID, err := topLayer.DiffID()
	if err != nil {
		return nil, "", errors.Wrap(err, "calculate layer diff ID")
	}
	return newImage, diffID.String(), nil
}

func (e *Exporter) createTarFile(tarFile, fsDir, tarDir string) error {
	fh, err := os.Create(tarFile)
	if err != nil {
		return errors.Wrap(err, "create file for tar")
	}
	defer fh.Close()
	gzw := gzip.NewWriter(fh)
	defer gzw.Close()
	tw := tar.NewWriter(gzw)
	defer tw.Close()

	return filepath.Walk(fsDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.Mode().IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(fsDir, file)
		if err != nil {
			return err
		}

		var header *tar.Header
		if fi.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(file)
			if err != nil {
				return err
			}
			header, err = tar.FileInfoHeader(fi, target)
			if err != nil {
				return err
			}
		} else {
			header, err = tar.FileInfoHeader(fi, fi.Name())
			if err != nil {
				return err
			}
		}
		header.Name = filepath.Join(tarDir, relPath)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if fi.Mode().IsRegular() {
			f, err := os.Open(file)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
		}
		return nil
	})
}
