#!/bin/sh

set -e

while [ ! -e /pipes/input ]
do
	echo "Waiting for input pipe to appear"
	sleep 1
done

while read X; do
	source "$FUNCTION_URI" "$X";
done </pipes/input >/pipes/output
