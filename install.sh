#!/usr/bin/env bash
set -euo pipefail

readonly FORGE_URL="https://fj.ktbcloud.com"
readonly REPO="bkenks/lazymux"
readonly INSTALL_DIR="${HOME}/.local/bin"
tmp_dir=""

fail() {
  echo "lazymux install: $*" >&2
  exit 1
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || fail "$1 is required but not installed"
}

detect_os() {
  case "$(uname -s)" in
  Darwin) echo darwin ;;
  Linux) echo linux ;;
  *) fail "unsupported OS $(uname -s); download a binary from ${FORGE_URL}/${REPO}/releases" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
  x86_64 | amd64) echo amd64 ;;
  arm64 | aarch64) echo arm64 ;;
  *) fail "unsupported CPU $(uname -m); download a binary from ${FORGE_URL}/${REPO}/releases" ;;
  esac
}

get_latest_tag() {
  local release_json tag
  release_json="$(curl -fsSL "${FORGE_URL}/api/v1/repos/${REPO}/releases/latest")" ||
    fail "couldn't reach ${FORGE_URL} to look up the latest release"
  tag="$(printf '%s' "$release_json" | sed -n 's/.*"tag_name":"\([^"]*\)".*/\1/p')"
  [ -n "$tag" ] || fail "couldn't read the latest release tag from ${FORGE_URL}"
  echo "$tag"
}

verify_checksum() {
  local dir="$1" artifact="$2" expected actual
  expected="$(awk -v name="$artifact" '$2 == name || $2 == "*" name { print $1 }' "$dir/SHA256SUMS")"
  [ -n "$expected" ] || fail "SHA256SUMS has no entry for $artifact"
  if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "$dir/$artifact" | awk '{ print $1 }')"
  else
    actual="$(shasum -a 256 "$dir/$artifact" | awk '{ print $1 }')"
  fi
  [ "$expected" = "$actual" ] || fail "checksum mismatch for $artifact; not installing it"
}

main() {
  require_command curl
  require_command awk
  command -v sha256sum >/dev/null 2>&1 || require_command shasum

  local os arch tag artifact download_url
  os="$(detect_os)"
  arch="$(detect_arch)"
  tag="$(get_latest_tag)"
  artifact="lazymux-${tag}-${os}-${arch}"
  download_url="${FORGE_URL}/${REPO}/releases/download/${tag}"

  tmp_dir="$(mktemp -d)"
  trap 'rm -r "$tmp_dir"' EXIT

  echo "Downloading lazymux ${tag} for ${os}/${arch}..."
  curl -fsSL -o "$tmp_dir/$artifact" "$download_url/$artifact" ||
    fail "couldn't download $download_url/$artifact"
  curl -fsSL -o "$tmp_dir/SHA256SUMS" "$download_url/SHA256SUMS" ||
    fail "couldn't download $download_url/SHA256SUMS"
  verify_checksum "$tmp_dir" "$artifact"

  mkdir -p "$INSTALL_DIR"
  install -m 0755 "$tmp_dir/$artifact" "$INSTALL_DIR/lazymux"
  echo "Installed lazymux ${tag} to ${INSTALL_DIR}/lazymux"

  case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *) echo "Add ${INSTALL_DIR} to your PATH to run lazymux." ;;
  esac
}

main "$@"
