#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.13"
# ///
# MISE description="Build and install gitkeeper to $GOBIN (--dev for gitkeeper-dev)"

"""Build gitkeeper for this machine and install it to $GOBIN.

mise run install          build and install gitkeeper
mise run install --dev    build and install gitkeeper-dev, sandboxed to gitkeeper-dev dirs
"""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from _lib import build_gitkeeper, build_gitkeeper_dev, install_binary


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        prog="mise run install", description="Build and install gitkeeper to $GOBIN."
    )
    parser.add_argument(
        "--dev",
        action="store_true",
        help="install gitkeeper-dev, sandboxed to gitkeeper-dev dirs, instead of gitkeeper",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    built = build_gitkeeper_dev() if args.dev else build_gitkeeper()
    dest = install_binary(built)
    print(f"installed {dest}")


main()
