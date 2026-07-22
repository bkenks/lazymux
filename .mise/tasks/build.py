#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.13"
# ///
#MISE description="Build lazymux (host by default; --all cross-compiles the release matrix)"

"""Build the lazymux binary.

    mise run build                      build/bin/lazymux for this machine
    mise run build --all                build/dist/* for every release platform
    mise run build --platform linux/amd64   build/dist/* for one platform
    mise run build --list               print the release matrix and exit

Plain `mise run build` stays host-only so the local edit-rebuild loop and
`mise run install` stay fast. `release` calls this with --all.
"""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from _lib import (
    PLATFORMS,
    build_lazymux,
    build_matrix,
    build_version,
    die,
    dist_dir,
    write_checksums,
)


def parse_platform(value: str) -> tuple[str, str]:
    """Parse a `goos/goarch` pair, validating it against the release matrix."""
    if value.count("/") != 1:
        die(f"{value!r} is not a goos/goarch pair, e.g. linux/amd64")
    goos, goarch = value.split("/")
    if (goos, goarch) not in PLATFORMS:
        supported = ", ".join(f"{o}/{a}" for o, a in PLATFORMS)
        die(f"{value} is not in the release matrix ({supported})")
    return goos, goarch


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        prog="mise run build", description="Build the lazymux binary."
    )
    parser.add_argument(
        "--all",
        action="store_true",
        help="cross-compile every platform in the release matrix into build/dist",
    )
    parser.add_argument(
        "--platform",
        action="append",
        metavar="GOOS/GOARCH",
        help="cross-compile one platform into build/dist (repeatable)",
    )
    parser.add_argument(
        "--list", action="store_true", help="print the release matrix and exit"
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()

    if args.list:
        for goos, goarch in PLATFORMS:
            print(f"{goos}/{goarch}")
        return

    if args.all and args.platform:
        die("--all and --platform are mutually exclusive")

    version = build_version()

    if not args.all and not args.platform:
        output = build_lazymux(version)
        print(f"built {output} ({version})")
        return

    targets = PLATFORMS if args.all else tuple(parse_platform(p) for p in args.platform)
    artifacts = build_matrix(version, targets)
    checksums = write_checksums(artifacts)

    print(f"\nbuilt {len(artifacts)} artifact(s) for {version} in {dist_dir()}")
    for artifact in artifacts:
        print(f"  {artifact.name}")
    print(f"  {checksums.name}")


main()
