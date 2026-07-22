"""Shared helpers for the lazymux mise tasks.

This file is intentionally not executable, so mise does not expose it as a task.
Task scripts import it by adding their own directory to `sys.path`.
"""

from __future__ import annotations

import hashlib
import os
import shutil
import subprocess
import sys
from pathlib import Path

MODULE = "github.com/bkenks/lazymux"
DEV_DIR_NAME = "lazymux-dev"

# Release matrix. lazymux is pure Go (no cgo in the dependency graph), so these
# cross-compile with nothing but GOOS/GOARCH — no C toolchain required.
PLATFORMS: tuple[tuple[str, str], ...] = (
    ("darwin", "amd64"),
    ("darwin", "arm64"),
    ("linux", "amd64"),
    ("linux", "arm64"),
    ("windows", "amd64"),
    ("windows", "arm64"),
)

CHECKSUM_FILE = "SHA256SUMS"


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
    """Where host builds land — the binary `install` picks up."""
    return repo_root() / "build" / "bin"


def dist_dir() -> Path:
    """Where cross-compiled release artifacts land."""
    return repo_root() / "build" / "dist"


def run(*cmd: str | Path, cwd: Path | None = None, env: dict[str, str] | None = None) -> None:
    """Echo a command, then run it from the repo root, failing on non-zero exit."""
    printable = " ".join(str(c) for c in cmd)
    print(f"$ {printable}", file=sys.stderr)
    result = subprocess.run(
        [str(c) for c in cmd],
        cwd=cwd or repo_root(),
        env={**os.environ, **env} if env else None,
    )
    if result.returncode != 0:
        die(f"command failed with exit {result.returncode}: {printable}")


def build_version() -> str:
    """Describe the working tree as a version string, matching the old Makefile."""
    try:
        return capture("git", "describe", "--tags", "--always", "--dirty", cwd=repo_root())
    except (subprocess.CalledProcessError, FileNotFoundError):
        return "dev"


def host_platform() -> tuple[str, str]:
    """The GOOS/GOARCH this machine builds for by default."""
    try:
        return capture("go", "env", "GOOS"), capture("go", "env", "GOARCH")
    except (subprocess.CalledProcessError, FileNotFoundError):
        die("could not run `go env` — is the Go toolchain installed? (try `mise install`)")
        raise  # unreachable


def go_build(
    output: Path, ldflags: str, goos: str | None = None, goarch: str | None = None
) -> Path:
    """Compile the module at the repo root into `output` with the given ldflags.

    Passing goos/goarch cross-compiles. CGO is disabled for cross builds so the
    result is a static binary and no C cross-toolchain is needed.
    """
    output.parent.mkdir(parents=True, exist_ok=True)
    env = None
    if goos and goarch:
        env = {"GOOS": goos, "GOARCH": goarch, "CGO_ENABLED": "0"}
    run("go", "build", "-ldflags", ldflags, "-o", output, ".", env=env)
    return output


def release_ldflags(version: str) -> str:
    return f"-X main.buildVersion={version}"


def build_lazymux(version: str | None = None) -> Path:
    """Build the host release binary into build/bin. Defaults to the describe version."""
    version = version or build_version()
    return go_build(bin_dir() / "lazymux", release_ldflags(version))


def build_lazymux_dev(version: str | None = None) -> Path:
    """Build the dev binary, whose config/repo dir is redirected to ~/lazymux-dev."""
    version = version or build_version()
    ldflags = (
        f"-X main.buildVersion={version}-dev -X {MODULE}/internal/config.dirName={DEV_DIR_NAME}"
    )
    return go_build(bin_dir() / "lazymux-dev", ldflags)


def asset_name(version: str, goos: str, goarch: str) -> str:
    """Release artifact filename, e.g. lazymux-v1.2.3-windows-amd64.exe."""
    suffix = ".exe" if goos == "windows" else ""
    return f"lazymux-{version}-{goos}-{goarch}{suffix}"


def build_matrix(version: str, platforms: tuple[tuple[str, str], ...] = PLATFORMS) -> list[Path]:
    """Cross-compile `version` for every platform, returning the artifact paths."""
    out = dist_dir()
    artifacts: list[Path] = []
    for index, (goos, goarch) in enumerate(platforms, start=1):
        print(f"[{index}/{len(platforms)}] {goos}/{goarch}", file=sys.stderr)
        target = out / asset_name(version, goos, goarch)
        artifacts.append(go_build(target, release_ldflags(version), goos, goarch))
    return artifacts


def write_checksums(artifacts: list[Path]) -> Path:
    """Write a sha256sum-compatible SHA256SUMS next to the artifacts."""
    if not artifacts:
        die("no artifacts to checksum")
    path = artifacts[0].parent / CHECKSUM_FILE
    lines = []
    for artifact in sorted(artifacts, key=lambda p: p.name):
        digest = hashlib.sha256(artifact.read_bytes()).hexdigest()
        lines.append(f"{digest}  {artifact.name}")
    path.write_text("\n".join(lines) + "\n")
    return path


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
