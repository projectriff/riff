#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

if [ $STAGE = "remote" ]; then
  make release
  
  bucket=gs://projectriff/riff-cli/releases
  gsutil cp -a public-read -n riff-*{.tgz,.zip} ${bucket}/builds/v${VERSION_SLUG}/
else
  make install
fi
