#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.13"
# ///
#MISE description="Cut a release: bump the version tag, build every platform, push, publish"

"""Cut a lazymux release.

    mise run release patch          v1.0.2 -> v1.0.3
    mise run release minor          v1.0.2 -> v1.1.0
    mise run release major          v1.0.2 -> v2.0.0
    mise run release v1.4.0         explicit version
    mise run release patch --dry-run    print the plan, change nothing

Preflight refuses to release from a dirty tree, a non-release branch, or a
branch that has diverged from its remote. The full platform matrix is built and
the test suite runs *before* the tag is created, so a broken build never gets
tagged. Every artifact is attached to the Forgejo release.
"""

from __future__ import annotations

import argparse
import re
import shutil
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from _lib import PLATFORMS, build_matrix, capture, die, dist_dir, repo_root, run, write_checksums

RELEASE_BRANCH = "main"
REMOTE = "origin"
SEMVER = re.compile(r"^v(\d+)\.(\d+)\.(\d+)$")
BUMPS = ("patch", "minor", "major")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        prog="mise run release",
        description="Build every platform, tag, push and publish a lazymux release.",
    )
    parser.add_argument(
        "version",
        metavar="patch|minor|major|vX.Y.Z",
        help="how to bump the latest tag, or an explicit version",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="print what would happen without tagging, pushing or publishing",
    )
    parser.add_argument(
        "-y",
        "--yes",
        action="store_true",
        help="skip the confirmation prompt (required when stdin is not a TTY)",
    )
    parser.add_argument(
        "--no-publish",
        action="store_true",
        help="push the tag but do not create a Forgejo release",
    )
    return parser.parse_args()


def latest_tag() -> str:
    """Return the highest semver tag, or v0.0.0 if the repo has none."""
    tags = capture("git", "tag", "--list", "v*", cwd=repo_root()).splitlines()
    versions = [t for t in (tag.strip() for tag in tags) if SEMVER.match(t)]
    if not versions:
        return "v0.0.0"
    return max(versions, key=lambda v: tuple(int(p) for p in SEMVER.match(v).groups()))


def next_version(current: str, spec: str) -> str:
    """Apply a bump keyword to `current`, or validate `spec` as an explicit version."""
    if spec not in BUMPS:
        if not SEMVER.match(spec):
            die(f"{spec!r} is not a bump keyword ({', '.join(BUMPS)}) or a vX.Y.Z version")
        return spec

    major, minor, patch = (int(p) for p in SEMVER.match(current).groups())
    if spec == "major":
        return f"v{major + 1}.0.0"
    if spec == "minor":
        return f"v{major}.{minor + 1}.0"
    return f"v{major}.{minor}.{patch + 1}"


def check_newer(current: str, new: str) -> None:
    """Refuse to go backwards, which would produce a confusing tag history."""
    if current == "v0.0.0":
        return
    current_parts = tuple(int(p) for p in SEMVER.match(current).groups())
    new_parts = tuple(int(p) for p in SEMVER.match(new).groups())
    if new_parts <= current_parts:
        die(f"{new} is not newer than the latest tag {current}")


def preflight(new: str) -> None:
    """Refuse to release from a tree that is dirty, off-branch, or out of sync."""
    root = repo_root()

    if capture("git", "status", "--porcelain", cwd=root):
        die("working tree is dirty — commit or stash your changes first")

    branch = capture("git", "rev-parse", "--abbrev-ref", "HEAD", cwd=root)
    if branch != RELEASE_BRANCH:
        die(f"on branch {branch!r}, but releases are cut from {RELEASE_BRANCH!r}")

    run("git", "fetch", "--tags", REMOTE)

    local = capture("git", "rev-parse", "HEAD", cwd=root)
    remote = capture("git", "rev-parse", f"{REMOTE}/{RELEASE_BRANCH}", cwd=root)
    if local != remote:
        die(
            f"{branch} has diverged from {REMOTE}/{RELEASE_BRANCH} — "
            "push or pull before releasing"
        )

    if capture("git", "tag", "--list", new, cwd=root):
        die(f"tag {new} already exists")


def confirm(prompt: str, assume_yes: bool) -> None:
    if assume_yes:
        return
    if not sys.stdin.isatty():
        die("stdin is not a TTY — re-run with --yes to confirm non-interactively")
    if input(f"{prompt} [y/N] ").strip().lower() not in ("y", "yes"):
        die("aborted")


def publish(version: str, assets: list[Path]) -> None:
    """Create the Forgejo release via tea, attaching every build artifact."""
    if shutil.which("tea") is None:
        print(
            f"warning: `tea` not found — tag {version} was pushed, but no Forgejo "
            "release was created. Install tea, or create it in the web UI and "
            f"upload the artifacts from {dist_dir()}.",
            file=sys.stderr,
        )
        return
    cmd: list[str | Path] = [
        "tea",
        "release",
        "create",
        "--tag",
        version,
        "--title",
        version,
        "--note",
        f"lazymux {version}",
    ]
    for asset in assets:
        cmd += ["--asset", asset]
    run(*cmd)


def main() -> None:
    args = parse_args()

    current = latest_tag()
    new = next_version(current, args.version)
    check_newer(current, new)

    print(f"releasing {current} -> {new}")
    preflight(new)

    # Build and test before tagging, so a broken tree never gets a tag. The
    # matrix is built here rather than via `mise run build --all` because the
    # artifacts must embed `new`, and that tag does not exist yet.
    run("go", "vet", "./...")
    run("go", "test", "./...")
    artifacts = build_matrix(new)
    assets = [*artifacts, write_checksums(artifacts)]

    if args.dry_run:
        publish_line = (
            "  (publish skipped)"
            if args.no_publish
            else f"  tea release create --tag {new} with {len(assets)} assets"
        )
        print(
            f"\ndry run — would have:\n"
            f"  git tag -a {new} -m 'Release {new}'\n"
            f"  git push {REMOTE} {new}\n"
            f"{publish_line}\n"
            f"\nbuilt artifacts left in {dist_dir()}"
        )
        return

    confirm(f"tag {new}, push it to {REMOTE}, and publish {len(assets)} assets?", args.yes)

    run("git", "tag", "-a", new, "-m", f"Release {new}")
    try:
        run("git", "push", REMOTE, new)
    except SystemExit:
        # Leave no dangling local tag if the push was rejected.
        subprocess.run(["git", "tag", "-d", new], cwd=repo_root(), check=False)
        raise

    if not args.no_publish:
        publish(new, assets)

    print(f"\nreleased {new} ({len(PLATFORMS)} platforms)")
    for asset in assets:
        print(f"  {asset.name}")


main()
