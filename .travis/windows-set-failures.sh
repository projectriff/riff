#!/bin/bash

if [ "$TRAVIS_EVENT_TYPE" != "pull_request" ]
then
    export TRAVIS_ALLOW_FAILURE=true
fi