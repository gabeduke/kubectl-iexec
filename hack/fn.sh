#!/usr/bin/env bash

set -ef pipefail
set -x

go_fmt () {
    echo "Running go fmt"

    # Use an eval to avoid glob expansion
    FIND_EXEC="find . -type f -iname '*.go'"

    # Get a list of files that we are interested in
    CHECK_FILES=$(eval "${FIND_EXEC}")

    set +e
    test -z "$(gofmt -l -d -e "${CHECK_FILES}")"
    SUCCESS=$?
    set -e

    # Exit if `go fmt` passes.
    if [ $SUCCESS -eq 0 ]; then
        exit 0
    fi


    # Get list of unformatted files.
    set +e
    ISSUE_FILES=$(gofmt -l "${CHECK_FILES}")
    echo "${ISSUE_FILES}"
    set -e

    # Iterate through each unformatted file.
    OUTPUT=""
    for FILE in $ISSUE_FILES; do
        DIFF=$(gofmt -d -e "${FILE}")
        OUTPUT="$OUTPUT
\`${FILE}\`
$(format_code "${DIFF}" diff)
"
    done

    echo "${OUTPUT}"
    post "go fmt" "${OUTPUT}"

    exit $SUCCESS
}

go_test() {
    echo "Running go test"

    # short mode for printing results to PR
    set +e
    go test ./... -v -short -race | grep FAIL > test-results.txt
    SUCCESS=${PIPESTATUS[0]}
    set -e

    if [ "$SUCCESS" != 0 ]; then
        FAILED=$(cat test-results.txt)
        FAILED=$(format_code "${FAILED}" less)
        post "Go Test" "${FAILED}"
        exit "$SUCCESS"
    fi

    # verbose for upload to codecov
    set +e
    go test ./... -race -coverprofile=coverage.txt -covermode=atomic
    SUCCESS=$?
    set -e

    if [ $SUCCESS != 0 ]; then
        exit $SUCCESS
    fi

    set +x
    if [ -z "$CODECOV_TOKEN" ]
    then
        echo "No Codecov token provided. Skipping.."
        exit $SUCCESS
    else
        curl -s https://codecov.io/bash | bash -s -- -t "$CODECOV_TOKEN" -f ./coverage.txt
        exit $SUCCESS
    fi
    set -x

}

format_code() {
    local CODE="${1}"
    local SYNTAX="${2}"
    CODE="
\`\`\`${SYNTAX}
${CODE}
\`\`\`
    "
    echo "${CODE}"
}

post() {

    STEP="${1}"
    FAILED="${2}"

    # Post results back as comment.
    COMMENT="#### \`${STEP}\`
${FAILED}
"
    PAYLOAD=$(echo '{}' | jq --arg body "$COMMENT" '.body = $body')
    COMMENTS_URL=$(cat < /github/workflow/event.json | jq -r .pull_request.comments_url)

    if [ "COMMENTS_URL" != null ]; then
        set +e
        curl -s -S -H "Authorization: token $GITHUB_TOKEN" --header "Content-Type: application/json" --data "$PAYLOAD" "$COMMENTS_URL" > /dev/null
        set -e
    fi
}

