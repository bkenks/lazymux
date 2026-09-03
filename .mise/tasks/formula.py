#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.13"
# ///
#MISE description="Render Formula/lazymux.rb from the built release artifacts"

"""Render the Homebrew formula for a built release.

    mise run formula                    version from `git describe`
    mise run formula --version 1.4.0    explicit version

This repo is its own Homebrew tap (`brew tap bkenks/lazymux <repo url>`), so the
formula lives here rather than in a separate homebrew-* repo. The formula
installs the prebuilt binary for the running platform, so no Go toolchain is
needed to `brew install`.

Checksums come from the `SHA256SUMS` written by `mise run build --all`, so run
that first. CI does exactly that, then commits the rendered formula.
"""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from _lib import CHECKSUM_FILE, asset_name, build_version, die, dist_dir, repo_root

# The platforms Homebrew supports, as (goos, goarch) -> formula block.
BREW_PLATFORMS: tuple[tuple[str, str], ...] = (
    ("darwin", "arm64"),
    ("darwin", "amd64"),
    ("linux", "arm64"),
    ("linux", "amd64"),
)

FORMULA_PATH = Path("Formula") / "lazymux.rb"

TEMPLATE = '''class Lazymux < Formula
  desc "TUI repo manager that also serves its repo inventory over MCP"
  homepage "https://fj.ktbcloud.com/bkenks/lazymux"
  version "{version}"
  license "GPL-3.0-or-later"

  on_macos do
    on_arm do
      url "{darwin_arm64_url}"
      sha256 "{darwin_arm64_sha}"
    end
    on_intel do
      url "{darwin_amd64_url}"
      sha256 "{darwin_amd64_sha}"
    end
  end

  on_linux do
    on_arm do
      url "{linux_arm64_url}"
      sha256 "{linux_arm64_sha}"
    end
    on_intel do
      url "{linux_amd64_url}"
      sha256 "{linux_amd64_sha}"
    end
  end

  def install
    bin.install Dir["lazymux-*"].first => "lazymux"
  end

  service do
    run [opt_bin/"lazymux", "mcp", "serve"]
    keep_alive true
    working_dir Dir.home
    log_path var/"log/lazymux-mcp.log"
    error_log_path var/"log/lazymux-mcp.log"
  end

  test do
    assert_match "lazymux #{{version}}", shell_output("#{{bin}}/lazymux --version")
  end
end
'''

DOWNLOAD_URL = "https://fj.ktbcloud.com/bkenks/lazymux/releases/download/{version}/{asset}"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        prog="mise run formula",
        description="Render the Homebrew formula from the built release artifacts.",
    )
    parser.add_argument(
        "--version",
        metavar="VERSION",
        help="release version the formula points at (default: git describe)",
    )
    return parser.parse_args()


def read_checksums() -> dict[str, str]:
    """Parse build/dist/SHA256SUMS into {filename: sha256}."""
    path = dist_dir() / CHECKSUM_FILE
    if not path.is_file():
        die(f"{path} not found — run `mise run build --all` first")
    sums: dict[str, str] = {}
    for line in path.read_text().splitlines():
        digest, _, name = line.partition("  ")
        if digest and name:
            sums[name] = digest
    return sums


def render(version: str, sums: dict[str, str]) -> str:
    fields = {"version": version}
    for goos, goarch in BREW_PLATFORMS:
        asset = asset_name(version, goos, goarch)
        if asset not in sums:
            die(f"{asset} is missing from {CHECKSUM_FILE} — was it built for {version}?")
        fields[f"{goos}_{goarch}_url"] = DOWNLOAD_URL.format(version=version, asset=asset)
        fields[f"{goos}_{goarch}_sha"] = sums[asset]
    return TEMPLATE.format(**fields)


def main() -> None:
    args = parse_args()
    version = args.version or build_version()
    output = repo_root() / FORMULA_PATH
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(render(version, read_checksums()))
    print(f"wrote {output} for {version}")


main()
