#!/bin/bash

TAG=$(cat /github/workspace/.release)

function get_bump_mode {
    BUMP_MODE=$1

    if [ -z "${BUMP_MODE}" ]; then
        BUMP_MODE="patch"
    fi

    if [ "${BUMP_MODE}" != "major" -a "${BUMP_MODE}" != "minor" -a "${BUMP_MODE}" != "patch" ]; then
        echo "bump-semver [option] [version] with option being major, minor or patch"
        exit 1
    fi

    echo $BUMP_MODE
}

function bump_version {
    BUMP_MODE=$2
    CURRENT_VERSION=$1

    MAJOR=$(echo "${CURRENT_VERSION#v}" | cut -f1 -d.)
    MINOR=$(echo "${CURRENT_VERSION#v}" | cut -f2 -d.)
    PATCH=$(echo "${CURRENT_VERSION#v}" | cut -f3 -d.)

    if [ "${BUMP_MODE}" == "major" ]; then
        NEW_MAJOR="$(( ${MAJOR} + 1 ))"
        NEW_VERSION="${NEW_MAJOR}.0.0"
    elif [ "${BUMP_MODE}" == "minor" ]; then
        NEW_MINOR="$(( ${MINOR} + 1 ))"
        NEW_VERSION="${MAJOR}.${NEW_MINOR}.0"
    elif [ "${BUMP_MODE}" == "patch" ]; then
        NEW_PATCH="$(( ${PATCH} + 1 ))"
        NEW_VERSION="${MAJOR}.${MINOR}.${NEW_PATCH}"
    else
        echo "yo, please select something to bump 1st (-.-) You can choose among: major, minor, and patch"
    fi

    if [ -n "${NEW_VERSION}" ]; then
        echo "${NEW_VERSION}"
    fi
}

BUMP_MODE=$(get_bump_mode $2)
if [ $? == "1" ]; then
    echo $BUMP_MODE
    exit $?
fi

CURRENT_VERSION=$TAG
if [ -z $CURRENT_VERSION ]; then
    CURRENT_VERSION="1.0.0"
    >&2 echo "warning: no previous version found. Created 1.0.0 as the 1st version."
    echo "${CURRENT_VERSION}" > .release
else
    VERSION=$(bump_version $CURRENT_VERSION $BUMP_MODE)
    echo "${VERSION}" > .release
fi