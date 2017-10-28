#!/bin/sh

set -e

while [ ! -e /pipes/input ]
do
	echo "Waiting for input pipe to appear"
	sleep 1
done

python -u ./funcrunner.py < pipes/input > pipes/output
