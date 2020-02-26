#!/usr/bin/env bash

_term() {
  echo "Caught SIGTERM signal!"
  pkill -SIGTERM subscribe
}

trap _term TERM

tail -f /dev/null & wait
