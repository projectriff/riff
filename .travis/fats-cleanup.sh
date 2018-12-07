#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

fats_dir=`dirname "${BASH_SOURCE[0]}"`/fats

# attempt to cleanup fats
if [ -d "$fats_dir" ]; then
  source $fats_dir/cleanup.sh
fi
