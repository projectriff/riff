#!/bin/sh

set -e

source $FUNCTION_URI

while read X; do echo $(userfunction "$X"); done </pipes/input >/pipes/output
