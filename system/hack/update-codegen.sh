#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(cd $(dirname "${BASH_SOURCE[0]}")/.. && pwd)

CODEGEN_PKG=${CODEGEN_PKG:-$(go mod download -json 2>/dev/null | jq -r 'select(.Path == "k8s.io/code-generator").Dir')}

TMP_DIR="$(mktemp -d)"
trap 'rm -rf ${TMP_DIR}' EXIT
export GOPATH=${GOPATH:-${TMP_DIR}}

TMP_REPO_PATH="${TMP_DIR}/src/github.com/projectriff/system"
mkdir -p "$(dirname "${TMP_REPO_PATH}")" && ln -s "${SCRIPT_ROOT}" "${TMP_REPO_PATH}"

bash "${CODEGEN_PKG}"/generate-groups.sh "client" \
  github.com/projectriff/system/pkg/client github.com/projectriff/system/pkg/apis \
  "build:v1alpha1 core:v1alpha1 knative:v1alpha1 streaming:v1alpha1" \
  --output-base "${TMP_DIR}/src" \
  --go-header-file "${SCRIPT_ROOT}"/hack/boilerplate.go.txt

# refine generated files

patchClient() {
  local find="$1"
  local replace="$2"

  find pkg/client -type f -name "*.go" -print0 | xargs -0 sed -i '' -e "s|${find}|${replace}|g"
}

# the fake client uses a naive GVK to GVR transform which doesn't use the correct pluralization. I feel bad, you should feel bad too.

patchClient \
  'schema.GroupVersionResource{Group: "streaming.projectriff.io", Version: "v1alpha1", Resource: "gateways"}' \
  'schema.GroupVersionResource{Group: "streaming.projectriff.io", Version: "v1alpha1", Resource: "gatewaies"}'
patchClient \
  'schema.GroupVersionResource{Group: "streaming.projectriff.io", Version: "v1alpha1", Resource: "inmemorygateways"}' \
  'schema.GroupVersionResource{Group: "streaming.projectriff.io", Version: "v1alpha1", Resource: "inmemorygatewaies"}'
patchClient \
  'schema.GroupVersionResource{Group: "streaming.projectriff.io", Version: "v1alpha1", Resource: "kafkagateways"}' \
  'schema.GroupVersionResource{Group: "streaming.projectriff.io", Version: "v1alpha1", Resource: "kafkagatewaies"}'
patchClient \
  'schema.GroupVersionResource{Group: "streaming.projectriff.io", Version: "v1alpha1", Resource: "pulsargateways"}' \
  'schema.GroupVersionResource{Group: "streaming.projectriff.io", Version: "v1alpha1", Resource: "pulsargatewaies"}'

make fmt
