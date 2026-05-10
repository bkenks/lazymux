#!/usr/bin/env bash
set -euo pipefail

# Installs the current working tree to $GOPATH/bin (same path as
# `go install github.com/bkenks/lazymux@latest`) so dev builds and
# released builds live in one place.
go install .
