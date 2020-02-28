#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

readonly bucket=gs://projectriff/release

cache_control='Cache-Control: public'
if echo $VERSION | grep -iqF snapshot; then
  cache_control="${cache_control}, max-age=60"
else
  cache_control="${cache_control}, max-age=3600"
fi

gsutil -h "${cache_control}" rsync -a public-read -d ${bucket}/snapshots/${VERSION_SLUG}/ ${bucket}/${VERSION}/
