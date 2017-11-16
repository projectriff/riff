#!/bin/sh

set -e

while [ ! -e /pipes/input ]
do
	echo "Waiting for input pipe to appear"
	sleep 1
done


if [ -z $FUNCTION_URI ]; then
	echo "Required variable FUNCTION_URI is not defined"
	exit 1
fi

while read X; do
	source "$FUNCTION_URI" "$X";
done </pipes/input >/pipes/output
