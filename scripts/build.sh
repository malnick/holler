#!/bin/bash
set -e

PLATFORM=$(uname | tr [:upper:] [:lower:])
GIT_REF=$(git describe --tags --always)
SOURCE_DIR=$(git rev-parse --show-toplevel)
VERSION=${GIT_REF}
REVISION=$(git rev-parse --short HEAD)

export PATH="${GOPATH}/bin:${PATH}"
export CGO_ENABLED=0

function build_proxy {
    go build -a -o ${BUILD_DIR}/holler-${COMPONENT}-${GIT_REF} \
        -ldflags "-X main.VERSION=${VERSION} -X main.REVISION=${REVISION}" \
        cmd/holler/*.go
}

function main {
    COMPONENT="$1"
    BUILD_DIR="${SOURCE_DIR}/build/${COMPONENT}"

    build_${COMPONENT} ${BUILD_DIR}
}

main "$@"
