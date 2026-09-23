#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.13"
# ///
# MISE description="Install lazymux-dev to $GOBIN"
# MISE depends=["dev"]

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from _lib import DEV_BINARY_NAME, bin_dir, install_binary

dest = install_binary(bin_dir() / DEV_BINARY_NAME)
print(f"installed {dest}")
