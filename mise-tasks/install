#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.13"
# ///
#MISE description="Install lazymux to $GOBIN"
#MISE depends=["build"]

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from _lib import bin_dir, install_binary

dest = install_binary(bin_dir() / "lazymux")
print(f"installed {dest}")
