#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.13"
# ///
#MISE description="Build build/bin/lazymux-dev, sandboxed to ~/lazymux-dev"

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from _lib import build_lazymux_dev, build_version

version = build_version()
output = build_lazymux_dev(version)
print(f"built {output} ({version}-dev)")
