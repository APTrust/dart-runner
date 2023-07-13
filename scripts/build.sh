#!/usr/bin/env bash

if [ -z $(ls scripts/build.sh 2> /dev/null) ]; then
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
mkdir -p dist/mac
GOOS=darwin GOARCH=amd64 go build -o dist/mac-x64/dart-runner -ldflags "-X 'main.Version=$VERSION'" $BUILD_TAGS

echo "Building MacOS arm-64 (M-chip) version in ./dist/mac-arm64/dart-runner"
mkdir -p dist/mac
GOOS=darwin GOARCH=arm64 go build -o dist/mac-arm64/dart-runner -ldflags "-X 'main.Version=$VERSION'" $BUILD_TAGS

echo "Building Windows amd64 version in ./dist/windows/dart-runner"
mkdir -p dist/windows
GOOS=windows GOARCH=amd64 go build -o dist/windows/dart-runner -ldflags "-X 'main.Version=$VERSION'" $BUILD_TAGS

echo "Building Linux amd64 version in ./dist/linux/dart-runner"
mkdir -p dist/linux
GOOS=linux GOARCH=amd64 go build -o dist/linux/dart-runner -ldflags "-X 'main.Version=$VERSION'" $BUILD_TAGS

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
