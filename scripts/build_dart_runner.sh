#!/usr/bin/env bash

# Note: CGO_ENABLED is set to zero in this script because we are
# not using any CGO dependencies. If CGO happens to be enabled in
# the greater environment, we want to override it here to prevent
# cross compilation failures. For details on why we do this, see
# https://stackoverflow.com/questions/77293369/dynamic-linked-go-program-when-cross-compile

if [ -z $(ls scripts/build_dart_runner.sh 2> /dev/null) ]; then
    echo "Run this script from the project root directory"
    exit
fi

COMMIT=$(git rev-parse --short HEAD)
TAG=$(git describe --tags 2> /dev/null)
DATE=$(date +%Y-%m-%d)
OS=$(uname -s)
ARCH=$(uname -m)

BUILD_TAGS=""

# uname returns MINGW64_NT-10.0 on Windows 10 Cygwin
# and MSYS_NT-10.0 on Windows 10 cmd.
if [[ "$OS" == *"_NT-"* ]]; then
	OS="Windows $(uname -m)"
	BUILD_TAGS="-tags windows"
fi

TAG="${TAG:=Beta 0.1}"
VERSION="DART Runner $TAG for $OS (Build $COMMIT $DATE)"

echo "Building MacOS amd64 version in ./dist/mac-x64/dart-runner"
mkdir -p dist/mac-x64
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o dist/mac-x64/dart-runner -ldflags "-X 'main.Version=$VERSION'" $BUILD_TAGS

echo "Building MacOS arm-64 (M-chip) version in ./dist/mac-arm64/dart-runner"
mkdir -p dist/mac-arm64
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o dist/mac-arm64/dart-runner -ldflags "-X 'main.Version=$VERSION'" $BUILD_TAGS

# Note: When running `./scripts/run tests`, post_build_test.go uses its own special build command for Windows.
echo "Building Windows amd64 version in ./dist/windows-x64/dart-runner"
mkdir -p dist/windows-x64
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o dist/windows-x64/dart-runner -ldflags "-X 'main.Version=$VERSION'" $BUILD_TAGS

echo "Building Windows arm64 version in ./dist/windows-arm64/dart-runner"
mkdir -p dist/windows-arm64
GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -o dist/windows-arm64/dart-runner -ldflags "-X 'main.Version=$VERSION'" $BUILD_TAGS

echo "Building Linux amd64 version in ./dist/linux-x64/dart-runner"
mkdir -p dist/linux-x64
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dist/linux-x64/dart-runner -ldflags "-X 'main.Version=$VERSION'" $BUILD_TAGS

echo "Building Linux arm64 version in ./dist/linux-arm64/dart-runner"
mkdir -p dist/linux-arm64
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o dist/linux-arm64/dart-runner -ldflags "-X 'main.Version=$VERSION'" $BUILD_TAGS

echo "Version info from latest build:"
if [[ "$OS" == "Darwin" ]]; then
    if [[ "$ARCH" == "x86_64" ]]; then
        dist/mac-x64/dart-runner --version
    else
        dist/mac-arm64/dart-runner --version
    fi
elif [[ "$OS" == "Linux" ]]; then
    dist/linux/dart-runner --version
else
    dist/windows/dart-runner --version
fi
