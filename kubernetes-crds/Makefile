.PHONY: clean codegen codegen-verify

GO_SOURCES = $(shell find pkg -type f -name '*.go')

# This is ONE of the generated files (alongside everything in pkg/client)
# that serves as make dependency tracking
GENERATED_SOURCE = pkg/apis/projectriff.io/v1/zz_generated.deepcopy.go

codegen: $(GENERATED_SOURCE)

$(GENERATED_SOURCE):
	vendor/k8s.io/code-generator/generate-groups.sh all \
      github.com/projectriff/kubernetes-crds/pkg/client \
      github.com/projectriff/kubernetes-crds/pkg/apis \
      projectriff.io:v1 \
      --go-header-file  hack/boilerplate.go.txt

codegen-verify:
	vendor/k8s.io/code-generator/generate-groups.sh all \
      github.com/projectriff/kubernetes-crds/pkg/client \
      github.com/projectriff/kubernetes-crds/pkg/apis \
      projectriff.io:v1 \
      --go-header-file  hack/boilerplate.go.txt \
      --verify-only

clean:
	rm -fR pkg/client
	rm -f $(GENERATED_SOURCE)

vendor: Gopkg.toml
	dep ensure

