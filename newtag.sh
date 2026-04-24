#!/usr/bin/env bash

NAME="$1"
TAG="$2"

git add .
git commit -m "$NAME $TAG"
git push
git tag "$TAG"
git push origin "$TAG"
