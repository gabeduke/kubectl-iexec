#!/usr/bin/env bash

set -x -e +o pipefail

export GOPATH=${HOME}/go
export GO111MODULE=on

go test -race -covermode=atomic -coverprofile=coverage.txt ./...