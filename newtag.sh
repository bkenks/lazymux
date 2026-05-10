#!/usr/bin/env bash
set -euo pipefail

# Cuts a new release: commits anything in the working tree, pushes,
# bumps the version tag, and pushes the tag.
#
# Usage:
#   ./newtag.sh             # patch bump (default)
#   ./newtag.sh minor       # 0.5.7 → 0.6.0
#   ./newtag.sh major
#   ./newtag.sh v1.2.3      # explicit version
#
# Once the tag is on origin, install with:
#   go install github.com/bkenks/lazymux@<tag>

bump=${1:-patch}

git fetch --tags origin >/dev/null 2>&1 || true
latest=$(git tag --list 'v*' --sort=-v:refname | head -n1)
latest=${latest:-v0.0.0}

case "$bump" in
  patch|minor|major)
    IFS='.' read -r major minor patch <<<"${latest#v}"
    case "$bump" in
      major) major=$((major + 1)); minor=0; patch=0 ;;
      minor) minor=$((minor + 1)); patch=0 ;;
      patch) patch=$((patch + 1)) ;;
    esac
    tag="v${major}.${minor}.${patch}"
    ;;
  v[0-9]*.[0-9]*.[0-9]*)
    tag="$bump"
    ;;
  *)
    echo "error: '$bump' must be patch|minor|major or vX.Y.Z" >&2
    exit 1
    ;;
esac

if [[ -n $(git status --porcelain) ]]; then
  git add -A
  git commit -m "$tag"
fi

git push
git tag "$tag"
git push origin "$tag"

echo "released $tag — install: go install github.com/bkenks/lazymux@$tag"
