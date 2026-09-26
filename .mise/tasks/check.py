#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.13"
# ///
# MISE description="Run every check: go vet, go test, golangci-lint, ruff, ty, shellcheck and shfmt"

"""Run the checks CI, the pre-commit hook and `mise run release` all gate on.

    mise run check

Go is vetted, tested and linted from the repo root; the mise task scripts are
linted, format-checked and type-checked; install.sh is linted and format-checked.
The first failing check stops the run.
"""

from __future__ import annotations

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from _lib import run

TASKS_DIR = ".mise/tasks"
INSTALL_SCRIPT = "install.sh"
PYTHON_VERSION = "3.13"


def main() -> None:
    """Run each check in turn, exiting non-zero at the first failure."""
    run("go", "vet", "./...")
    run("go", "test", "./...")
    run("golangci-lint", "run")
    run("ruff", "check", TASKS_DIR)
    run("ruff", "format", "--check", TASKS_DIR)
    run(
        "ty",
        "check",
        "--python-version",
        PYTHON_VERSION,
        "--extra-search-path",
        TASKS_DIR,
        TASKS_DIR,
    )
    run("shellcheck", INSTALL_SCRIPT)
    run("shfmt", "--indent", "2", "--diff", INSTALL_SCRIPT)


main()
