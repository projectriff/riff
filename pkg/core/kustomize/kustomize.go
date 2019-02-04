package kustomize

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sigs.k8s.io/kustomize/k8sdeps"
	"sigs.k8s.io/kustomize/pkg/commands/build"
	"sigs.k8s.io/kustomize/pkg/fs"
	"time"
)

type Kustomizer interface {
	// Applies the provided label to the provided remote or local resource definition
	// Returns the customized resource contents
	// Returns an error if
	// - the URL scheme is not supported (only file, http and https are)
	// - retrieving the content fails
	// - applying the customization fails
	ApplyLabel(resourceUri *url.URL, label *Label) ([]byte, error)
}

type Label struct {
	Name  string
	Value string
}

func (l *Label) AsMap() map[string]string {
	return map[string]string{l.Name: l.Value}
}

type kustomizer struct {
	fakeDir    string
	fs         fs.FileSystem
	httpClient *http.Client
}

func MakeKustomizer(timeout time.Duration) Kustomizer {
	return &kustomizer{
		fs:      fs.MakeFakeFS(),
		fakeDir: "/",
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (kust *kustomizer) ApplyLabel(resourceUri *url.URL, label *Label) ([]byte, error) {
	resourcePath, err := kust.writeResourceFile(resourceUri)
	if err != nil {
		return nil, err
	}
	err = kust.writeKustomizationFile(resourcePath, label)
	if err != nil {
		return nil, err
	}
	return kust.runBuild()
}

func (kust *kustomizer) writeResourceFile(resourceUri *url.URL) (string, error) {
	resourceContents, err := kust.fetch(resourceUri)
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

func (kust *kustomizer) writeKustomizationFile(resourcePath string, label *Label) error {
	err := kust.fs.WriteFile(kust.fakeDir+"kustomization.yaml", []byte(fmt.Sprintf(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
commonLabels:
  %s: %s
resources:
  - %s
`, label.Name, label.Value, resourcePath)))
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

func (kust *kustomizer) fetch(url *url.URL) (string, error) {
	switch url.Scheme {
	case "http":
		fallthrough
	case "https":
		return kust.fetchUri(url)
	case "file":
		return read(url)
	default:
		return "", fmt.Errorf("unsupported scheme %s", url.Scheme)
	}

}

func read(path *url.URL) (string, error) {
	contents, err := ioutil.ReadFile(path.Path)
	if err != nil {
		return "", nil
	}
	return string(contents), nil
}

func (kust *kustomizer) fetchUri(url *url.URL) (string, error) {
	response, err := kust.httpClient.Get(url.String())
	if err != nil {
		return "", err
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			panic(err)
		}
	}()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}
