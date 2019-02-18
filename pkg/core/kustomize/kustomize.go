package kustomize

import (
	"bytes"
	"fmt"
	"github.com/projectriff/riff/pkg/fileutils"
	"io/ioutil"
	"net/url"
	"sigs.k8s.io/kustomize/k8sdeps"
	"sigs.k8s.io/kustomize/pkg/commands/build"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sort"
	"strings"
	"time"
)

type Kustomizer interface {
	// Applies the provided labels to the provided remote or local resource definition
	// Returns the customized resource contents
	// Returns an error if
	// - the URL scheme is not supported (only file, http and https are)
	// - retrieving the content fails
	// - applying the customization fails
	// As of the current implementation, it is not safe to call this function concurrently
	ApplyLabels(resourceUri *url.URL, labels map[string]string) ([]byte, error)
}

type kustomizer struct {
	fakeDir     string
	fs          fs.FileSystem
	httpTimeout time.Duration
}

func MakeKustomizer(timeout time.Duration) Kustomizer {
	return &kustomizer{
		fs:          fs.MakeFakeFS(), // keep contents in-memory
		fakeDir:     "/",
		httpTimeout: timeout,
	}
}

func (kust *kustomizer) ApplyLabels(resourceUri *url.URL, labels map[string]string) ([]byte, error) {
	resourcePath, err := kust.writeResourceFile(resourceUri)
	if err != nil {
		return nil, err
	}
	err = kust.writeKustomizationFile(resourcePath, labels)
	if err != nil {
		return nil, err
	}
	return kust.runBuild()
}

func (kust *kustomizer) writeResourceFile(resourceUri *url.URL) (string, error) {
	resourceContents, err := fileutils.ReadUrl(resourceUri, kust.httpTimeout)
	if err != nil {
		return "", err
	}
	resourcePath := "resource.yaml"
	err = kust.fs.WriteFile(kust.fakeDir+resourcePath, []byte(resourceContents))
	if err != nil {
		return "", err
	}
	return resourcePath, nil
}

func (kust *kustomizer) writeKustomizationFile(resourcePath string, labels map[string]string) error {
	err := kust.fs.WriteFile(kust.fakeDir+"kustomization.yaml", []byte(fmt.Sprintf(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
commonLabels:%s
resources:
  - %s
`, formatLabels(labels), resourcePath)))
	if err != nil {
		return err
	}
	return nil
}

func (kust *kustomizer) runBuild() ([]byte, error) {
	var out bytes.Buffer
	kustomizeFactory := k8sdeps.NewFactory()
	kustomizeBuildCommand := build.NewCmdBuild(&out, kust.fs, kustomizeFactory.ResmapF, kustomizeFactory.TransformerF)
	kustomizeBuildCommand.SetArgs([]string{kust.fakeDir})
	kustomizeBuildCommand.SetOutput(ioutil.Discard)
	_, err := kustomizeBuildCommand.ExecuteC()
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func formatLabels(labels map[string]string) string {
	builder := strings.Builder{}
	keys := keysOf(labels)
	sort.Strings(keys)
	for _, key := range keys {
		builder.WriteString(fmt.Sprintf("\n  %s: %s", key, labels[key]))
	}
	return builder.String()
}

func keysOf(dict map[string]string) []string {
	result := make([]string, len(dict))
	i := 0
	for k := range dict {
		result[i] = k
		i++
	}
	return result
}
