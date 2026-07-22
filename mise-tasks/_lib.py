"""Shared helpers for the lazymux mise tasks.

This file is intentionally not executable, so mise does not expose it as a task.
Task scripts import it by adding their own directory to `sys.path`.
"""

from __future__ import annotations

import os
import shutil
import subprocess
import sys
from pathlib import Path

MODULE = "github.com/bkenks/lazymux"
DEV_DIR_NAME = "lazymux-dev"


def die(message: str) -> None:
    """Print an error to stderr and exit non-zero."""
    print(f"error: {message}", file=sys.stderr)
    raise SystemExit(1)


def capture(*cmd: str | Path, cwd: Path | None = None) -> str:
    """Run a command and return its stripped stdout."""
    result = subprocess.run(
        [str(c) for c in cmd], cwd=cwd, check=True, capture_output=True, text=True
    )
    return result.stdout.strip()


def repo_root() -> Path:
    """Locate the repo root, preferring the root mise resolved."""
    root = os.environ.get("MISE_PROJECT_ROOT")
    if root:
        return Path(root)
    try:
        return Path(capture("git", "rev-parse", "--show-toplevel"))
    except (subprocess.CalledProcessError, FileNotFoundError):
        die("not inside a git repo and $MISE_PROJECT_ROOT is unset")
        raise  # unreachable; keeps type checkers happy


def bin_dir() -> Path:
    return repo_root() / "build" / "bin"


def run(*cmd: str | Path, cwd: Path | None = None) -> None:
    """Echo a command, then run it from the repo root, failing on non-zero exit."""
    printable = " ".join(str(c) for c in cmd)
    print(f"$ {printable}", file=sys.stderr)
    result = subprocess.run([str(c) for c in cmd], cwd=cwd or repo_root())
    if result.returncode != 0:
        die(f"command failed with exit {result.returncode}: {printable}")


def build_version() -> str:
    """Describe the working tree as a version string, matching the old Makefile."""
    try:
        return capture("git", "describe", "--tags", "--always", "--dirty", cwd=repo_root())
    except (subprocess.CalledProcessError, FileNotFoundError):
        return "dev"


def go_build(output: Path, ldflags: str) -> Path:
    """Compile the module at the repo root into `output` with the given ldflags."""
    output.parent.mkdir(parents=True, exist_ok=True)
    run("go", "build", "-ldflags", ldflags, "-o", output, ".")
    return output


def build_lazymux(version: str | None = None) -> Path:
    """Build the release binary. Defaults to the `git describe` version."""
    version = version or build_version()
    return go_build(bin_dir() / "lazymux", f"-X main.buildVersion={version}")


def build_lazymux_dev(version: str | None = None) -> Path:
    """Build the dev binary, whose config/repo dir is redirected to ~/lazymux-dev."""
    version = version or build_version()
    ldflags = (
        f"-X main.buildVersion={version}-dev -X {MODULE}/internal/config.dirName={DEV_DIR_NAME}"
    )
    return go_build(bin_dir() / "lazymux-dev", ldflags)


def gobin() -> Path:
    """Resolve $GOBIN, falling back to $(go env GOPATH)/bin."""
    try:
        path = capture("go", "env", "GOBIN") or f"{capture('go', 'env', 'GOPATH')}/bin"
    except (subprocess.CalledProcessError, FileNotFoundError):
        die("could not run `go env` — is the Go toolchain installed? (try `mise install`)")
        raise  # unreachable
    return Path(path)


def install_binary(src: Path) -> Path:
    """Copy a built binary into $GOBIN with mode 0755."""
    if not src.is_file():
        die(f"{src} does not exist")
    dest_dir = gobin()
    dest_dir.mkdir(parents=True, exist_ok=True)
    dest = dest_dir / src.name
    shutil.copy2(src, dest)
    dest.chmod(0o755)
    return dest
