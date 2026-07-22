#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.13"
# ///
#MISE description="Install lazymux-dev to $GOBIN"
#MISE depends=["dev"]

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from _lib import bin_dir, install_binary

dest = install_binary(bin_dir() / "lazymux-dev")
print(f"installed {dest}")
