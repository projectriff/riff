package kustomize_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/pkg/core/kustomize"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

var _ = Describe("Kustomize wrapper", func() {

	var (
		ownerLabel              *kustomize.Label
		initialResourceContent  string
		expectedResourceContent string
		kustomizer              kustomize.Kustomizer
		timeout                 time.Duration
	)

	BeforeEach(func() {
		ownerLabel = &kustomize.Label{Name: "created-by", Value: "riff"}
		initialResourceContent = `kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: riff-cnb-cache
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 8Gi`
		expectedResourceContent = fmt.Sprintf(`apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  labels:
    %s: %s
  name: riff-cnb-cache
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 8Gi
`, ownerLabel.Name, ownerLabel.Value)
		timeout = 500 * time.Millisecond
		kustomizer = kustomize.MakeKustomizer(timeout)
	})

	It("customizes remote resources with provided label", func() {
		resourceListener, _ := net.Listen("tcp", "127.0.0.1:0")
		go serve(resourceListener, initialResourceContent)
		resourceUrl := unsafeParseUrl(fmt.Sprintf("http://%s/%s", resourceListener.Addr().String(), "pvc.yaml"))

		result, err := kustomizer.ApplyLabel(resourceUrl, ownerLabel)

		Expect(err).To(Not(HaveOccurred()))
		Expect(string(result)).To(Equal(expectedResourceContent))
	})

	It("customizes local resources with provided label", func() {
		file, fileUri := localFile("pvc.yaml", initialResourceContent)
		defer unsafeClose(file)

		resourceUrl := unsafeParseUrl(fileUri)

		result, err := kustomizer.ApplyLabel(resourceUrl, ownerLabel)

		Expect(err).To(Not(HaveOccurred()))
		Expect(string(result)).To(Equal(expectedResourceContent))
	})

	It("fails on unsupported scheme", func() {
		resourceUrl := unsafeParseUrl("ftp://127.0.0.1/goodluck.yaml")

		_, err := kustomizer.ApplyLabel(resourceUrl, ownerLabel)

		Expect(err).To(MatchError("unsupported scheme ftp"))
	})

	It("fails if the resource is not reachable", func() {
		_, err := kustomizer.ApplyLabel(unsafeParseUrl("http://localhost:12345/nope.yaml"), ownerLabel)

		Expect(err).To(SatisfyAll(
			Not(BeNil()),
			BeAssignableToTypeOf(&url.Error{})))
	})

	It("fails if fetching the resource takes too long", func() {
		listener, _ := net.Listen("tcp", "127.0.0.1:0")
		go serveSlow(listener, initialResourceContent, 3*timeout)
		resourceUrl := unsafeParseUrl(fmt.Sprintf("http://%s/%s", listener.Addr().String(), "pvc.yaml"))

		_, err := kustomizer.ApplyLabel(resourceUrl, ownerLabel)

		Expect(err).To(SatisfyAll(
			Not(BeNil()),
			BeAssignableToTypeOf(&url.Error{})))
	})
})

func serve(listener net.Listener, response string) {
	serveSlow(listener, response, 0)
}

func serveSlow(listener net.Listener, response string, sleepDuration time.Duration) {
	err := http.Serve(listener, http.HandlerFunc(func(responseWriter http.ResponseWriter, req *http.Request) {
		time.Sleep(sleepDuration)
		responseWriter.WriteHeader(200)
		responseWriter.Header().Add("Content-Type", "application/octet-stream")
		_, _ = io.WriteString(responseWriter, response)
	}))

	if err != nil {
		panic(err)
	}
	return
}

func localFile(name string, content string) (*os.File, string) {
	file, err := ioutil.TempFile("", name)
	if err != nil {
		panic(err)
	}
	path, err := filepath.Abs(file.Name())
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(file.Name(), []byte(content), os.FileMode(0600))
	if err != nil {
		panic(err)
	}
	return file, fmt.Sprintf("file://%s", path)
}

func unsafeParseUrl(raw string) *url.URL {
	result, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return result
}

func unsafeClose(file *os.File) {
	err := file.Close()
	if err != nil {
		panic(err)
	}
}
