#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

./mvnw -q -B -V test
