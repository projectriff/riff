#!/bin/sh

set -e

while [ ! -e /pipes/input ]
do
	echo "Waiting for input pipe to appear"
	sleep 1
done

# todo: only use this as a default if no env var set
FUNCTION_URI=/functions/script.sh

while read X; do
	source "$FUNCTION_URI" "$X";
done </pipes/input >/pipes/output
