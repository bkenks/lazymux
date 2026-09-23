#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.13"
# ///
# MISE description="Build and install lazymux to $GOBIN (--dev for lazymux-dev)"

"""Build lazymux for this machine and install it to $GOBIN.

mise run install          build and install lazymux
mise run install --dev    build and install lazymux-dev, sandboxed to ~/lazymux-dev
"""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from _lib import build_lazymux, build_lazymux_dev, install_binary


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        prog="mise run install", description="Build and install lazymux to $GOBIN."
    )
    parser.add_argument(
        "--dev",
        action="store_true",
        help="install lazymux-dev, sandboxed to ~/lazymux-dev, instead of lazymux",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    built = build_lazymux_dev() if args.dev else build_lazymux()
    dest = install_binary(built)
    print(f"installed {dest}")


main()
