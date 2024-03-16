#!/usr/bin/env bash

set -e

TARGET_OS="$1"
TARGET_ARCH="$2"
BUILD_DIR="${PWD}/build"

if [ "$TARGET_OS" = "" ]; then
    echo "Target OS expected: [linux, darwin]"
    exit 1
fi

if [ "$TARGET_ARCH" = "" ]; then
    echo "Target architecture expected: [amd64, arm64]"
    exit 1
fi

check_build_dir() {
    if [ ! -f "$BUILD_DIR" ]; then
        echo "Creating build directory ./build"
        mkdir -p "$BUILD_DIR"
    fi
}

build_binary() {
    echo "Building $1 binary for $2 architecture"
    GOOS="$1" GOARCH="$2" go build -o "$BUILD_DIR/disconfig_$1_$2" *.go
}

check_build_dir
build_binary "$TARGET_OS" "$TARGET_ARCH"
