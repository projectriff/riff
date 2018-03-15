#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "\n${RED}This task only schedules a FATS job, it does not indicate FATS was successful${NC}";

body="{
    \"request\": {
        \"message\": \"Triggerd by ${TRAVIS_REPO_SLUG}#${TRAVIS_JOB_NUMBER}\",
        \"branch\": \"master\"
    }
}"

request=$(
    curl -s -X POST \
        -H "Content-Type: application/json" \
        -H "Accept: application/json" \
        -H "Travis-API-Version: 3" \
        -H "Authorization: token ${TRAVIS_API_TOKEN}" \
        -d "$body" \
        https://api.travis-ci.org/repo/projectriff%2Ffats/requests
)
request_id=`echo $request | jq '.request.id'`
sleep 5
request=$(
    curl -s \
        -H "Accept: application/json" \
        -H "Travis-API-Version: 3" \
        -H "Authorization: token ${TRAVIS_API_TOKEN}" \
        https://api.travis-ci.org/repo/projectriff%2Ffats/request/${request_id}
)

echo -e "View results at https://travis-ci.org/projectriff/fats/builds/`echo $request | jq -r '.builds[0].id'`"
