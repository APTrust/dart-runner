#!/bin/bash


if [ -z $(ls scripts/build.sh 2> /dev/null) ]
then
    echo "Run this script from the project root directory"
    exit
fi

COMMIT=$(git rev-parse --short HEAD)
TAG=$(git describe --tags 2> /dev/null)
DATE=$(date +%Y-%m-%d)
OS=$(uname -ms)

TAG="${TAG:=Beta 0.1}"
VERSION="DART Runner $TAG for $OS (Build $COMMIT $DATE)"

mkdir -p dist
go build -o dist/dart-runner -ldflags "-X 'main.Version=$VERSION'"

echo "Executable is in dist/dart-runner"
dist/dart-runner --version
