#!/usr/bin/env sh
set -e

MODULE="github.com/bkenks/lazymux@latest"

die() {
    printf "\033[31m%s\033[0m\n" "$1" >&2
    exit 1
}

if ! command -v go >/dev/null 2>&1; then
    die "Go is required to install lazymux. Install it from https://go.dev/dl and re-run this script."
fi

printf "Installing lazymux via 'go install %s'...\n" "$MODULE"
go install "$MODULE"

GOBIN="$(go env GOBIN)"
if [ -z "$GOBIN" ]; then
    GOBIN="$(go env GOPATH)/bin"
fi

printf "\n✅ Installed to %s/lazymux\n" "$GOBIN"

if ! echo "$PATH" | grep -q "$GOBIN"; then
    printf "\n⚠️  Add this to your shell config if 'lazymux' isn't found:\n"
    printf "export PATH=\"%s:\$PATH\"\n\n" "$GOBIN"
fi
