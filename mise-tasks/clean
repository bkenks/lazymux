#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.13"
# ///
#MISE description="Remove build/bin and build/dist"

import shutil
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from _lib import bin_dir, dist_dir

removed = False
for target in (bin_dir(), dist_dir()):
    if target.exists():
        shutil.rmtree(target)
        print(f"removed {target}")
        removed = True

if not removed:
    print("nothing to clean")
