#!/usr/bin/env bash

set -o pipefail
set -e

APP_DIR="/go/src/github.com/${GITHUB_REPOSITORY}/"
mkdir -p "${APP_DIR}" && cp -r ./ "${APP_DIR}" && cd "${APP_DIR}"

go_fmt () {
    echo "Running go fmt"

    # Use an eval to avoid glob expansion
    FIND_EXEC="find . -type f -iname '*.go'"

    # Get a list of files that we are interested in
    CHECK_FILES=$(eval ${FIND_EXEC})

    set +e
    test -z "$(gofmt -l -d -e ${CHECK_FILES})"
    SUCCESS=$?
    set -e

    # Exit if `go fmt` passes.
    if [ $SUCCESS -eq 0 ]; then
        exit 0
    fi


    # Get list of unformatted files.
    set +e
    ISSUE_FILES=$(gofmt -l ${CHECK_FILES})
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

go_lint () {
    echo "Running go lint"
    go get ./...

    set +e
    golangci-lint run --out-format tab > lint-results.txt
    SUCCESS=$?
    set -e

    if [ $SUCCESS != 0 ]; then
        FAILED=$(cat lint-results.txt)
        FAILED=$(format_code "${FAILED}" less)
        echo "${FAILED}"
        post "Go Lint" "${FAILED}"
        exit $SUCCESS
    fi

    exit $SUCCESS
}

go_test() {
    echo "Running go test"

    set +e
    go test ./... -v -short | grep FAIL > test-results.txt
    SUCCESS=${PIPESTATUS[0]}
    set -e

    if [ $SUCCESS != 0 ]; then
        FAILED=$(cat test-results.txt)
        FAILED=$(format_code "${FAILED}" less)
        post "Go Test" "${FAILED}"
        exit $SUCCESS
    fi

    exit $SUCCESS
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
    COMMENTS_URL=$(cat /github/workflow/event.json | jq -r .pull_request.comments_url)

    if [ "COMMENTS_URL" != null ]; then
        set +e
        curl -s -S -H "Authorization: token $GITHUB_TOKEN" --header "Content-Type: application/json" --data "$PAYLOAD" "$COMMENTS_URL" > /dev/null
        set -e
    fi
}

### MAIN

case "$1" in
'fmt')
go_fmt
;;
'lint')
go_lint
;;
'test')
go_test
;;
esac