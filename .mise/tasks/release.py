#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.13"
# ///
#MISE description="Cut a release: check the tree, tag the new version and push it"

"""Cut a lazymux release.

    mise run release patch          v1.0.2 -> v1.0.3
    mise run release minor          v1.0.2 -> v1.1.0
    mise run release major          v1.0.2 -> v2.0.0
    mise run release v1.4.0         explicit version
    mise run release patch --dry-run    print the plan, change nothing

Preflight refuses to release from a dirty tree, a non-release branch, or a
branch that has diverged from its remote, and the test suite runs *before* the
tag is created, so a broken tree never gets tagged.

Tags are `vX.Y.Z`. The prefix is what Go modules require, so it is also what keeps
`go install github.com/bkenks/lazymux@latest` resolving to the newest release.

Pushing the tag is the whole job: Woodpecker (`.woodpecker.yml`) picks up the
`v*` tag, cross-compiles the release matrix and creates the Forgejo release with
every binary attached.
"""

from __future__ import annotations

import argparse
import re
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from _lib import capture, die, repo_root, run

RELEASE_BRANCH = "main"
REMOTE = "origin"
# Tags are written as vX.Y.Z, the form Go modules require. The `v` is optional
# when matching so a bare version can be passed on the command line.
SEMVER = re.compile(r"^v?(\d+)\.(\d+)\.(\d+)$")
BUMPS = ("patch", "minor", "major")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        prog="mise run release",
        description="Test, tag and push a lazymux release for CI to build and publish.",
    )
    parser.add_argument(
        "version",
        metavar="patch|minor|major|vX.Y.Z",
        help="how to bump the latest tag, or an explicit version",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="print what would happen without tagging or pushing",
    )
    parser.add_argument(
        "-y",
        "--yes",
        action="store_true",
        help="skip the confirmation prompt (required when stdin is not a TTY)",
    )
    return parser.parse_args()


def latest_tag() -> str:
    """Return the highest semver tag as vX.Y.Z, or v0.0.0 if the repo has none."""
    tags = capture("git", "tag", "--list", cwd=repo_root()).splitlines()
    matches = [SEMVER.match(tag.strip()) for tag in tags]
    versions = [tuple(int(p) for p in m.groups()) for m in matches if m]
    if not versions:
        return "v0.0.0"
    return "v" + ".".join(str(part) for part in max(versions))


def next_version(current: str, spec: str) -> str:
    """Apply a bump keyword to `current`, or validate `spec` as an explicit version."""
    if spec not in BUMPS:
        match = SEMVER.match(spec)
        if not match:
            die(f"{spec!r} is not a bump keyword ({', '.join(BUMPS)}) or a vX.Y.Z version")
        # Accept a bare version on the command line, tag it prefixed anyway.
        return "v" + ".".join(match.groups())

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


def main() -> None:
    args = parse_args()

    current = latest_tag()
    new = next_version(current, args.version)
    check_newer(current, new)

    print(f"releasing {current} -> {new}")
    preflight(new)

    # Test before tagging, so a broken tree never gets a tag. The release
    # binaries are built by CI from the tag, not here.
    run("go", "vet", "./...")
    run("go", "test", "./...")

    if args.dry_run:
        print(
            f"\ndry run — would have:\n"
            f"  git tag -a {new} -m 'Release {new}'\n"
            f"  git push {REMOTE} {new}\n"
            f"  (CI then builds the matrix and publishes the release)"
        )
        return

    confirm(f"tag {new} and push it to {REMOTE}?", args.yes)

    run("git", "tag", "-a", new, "-m", f"Release {new}")
    try:
        run("git", "push", REMOTE, new)
    except SystemExit:
        # Leave no dangling local tag if the push was rejected.
        subprocess.run(["git", "tag", "-d", new], cwd=repo_root(), check=False)
        raise

    print(
        f"\npushed {new} — Woodpecker is building the matrix and will create the "
        "release with every binary attached."
    )


main()
