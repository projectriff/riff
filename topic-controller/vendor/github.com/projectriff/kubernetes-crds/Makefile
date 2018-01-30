.PHONY: clean codegen codegen-verify


# This is ONE of the generated files (alongside everything in pkg/client)
# that serves as make dependency tracking
GENERATED_SOURCE = pkg/apis/projectriff.io/v1/zz_generated.deepcopy.go

GO_SOURCES = $(shell find pkg/apis -type f -name '*.go' ! -path $(GENERATED_SOURCE))

PKGS = $(shell go list ./pkg/...)

build: $(GO_SOURCES) vendor
	go build $(PKGS)

codegen: $(GENERATED_SOURCE)

$(GENERATED_SOURCE): $(GO_SOURCES) hack/vendor
	hack/vendor/k8s.io/code-generator/generate-groups.sh all \
      github.com/projectriff/kubernetes-crds/pkg/client \
      github.com/projectriff/kubernetes-crds/pkg/apis \
      projectriff.io:v1 \
      --go-header-file  hack/boilerplate.go.txt
	hack/vendor/k8s.io/code-generator/generate-internal-groups.sh defaulter \
      github.com/projectriff/kubernetes-crds/pkg/client \
      '' \
      github.com/projectriff/kubernetes-crds/pkg/apis \
      projectriff.io:v1 \
      --go-header-file  hack/boilerplate.go.txt

codegen-verify: hack/vendor
	hack/vendor/k8s.io/code-generator/generate-groups.sh all \
      github.com/projectriff/kubernetes-crds/pkg/client \
      github.com/projectriff/kubernetes-crds/pkg/apis \
      projectriff.io:v1 \
      --go-header-file  hack/boilerplate.go.txt \
      --verify-only

clean:
	rm -fR pkg/client
	rm -f $(GENERATED_SOURCE)

vendor: glide.lock
	glide install -v

glide.lock: glide.yaml
	glide up -v

hack/vendor: hack/glide.lock
	# Note the absence of -v
	cd hack && glide install

hack/glide.lock: hack/glide.yaml
	# Note the absence of -v
	cd hack && glide up
