#!/usr/bin/env bash

NAME="$1"
TAG="$2"
BUILD_SCR="build.sh"

git add .
git commit -m "$NAME $TAG"
git push
git tag "$TAG"
git push origin "$TAG"

if [[ -f $BUILD_SCR ]]; then
    bash "./$BUILD_SCR"
fi
