#!/usr/bin/env python3
"""
sparktype CLI wrapper - runs the bundled Go binary.
"""

import os
import subprocess
import sys
from pathlib import Path

__version__ = "0.0.0-development"


def get_binary_path() -> Path:
    """Get the path to the bundled binary."""
    package_dir = Path(__file__).parent
    binary_name = "sparktype.exe" if sys.platform == "win32" else "sparktype"
    return package_dir / "bin" / binary_name


def main():
    """Main entry point."""
    binary_path = get_binary_path()

    if not binary_path.exists():
        print(f"Error: sparktype binary not found at {binary_path}", file=sys.stderr)
        print("This may indicate an installation issue.", file=sys.stderr)
        print(
            "Try reinstalling: pip install --force-reinstall sparktype", file=sys.stderr
        )
        sys.exit(1)

    # Ensure binary is executable (on Unix)
    if sys.platform != "win32" and not os.access(binary_path, os.X_OK):
        try:
            binary_path.chmod(0o755)
        except OSError as e:
            print(f"Error: Could not make binary executable: {e}", file=sys.stderr)
            sys.exit(1)

    # Run the binary with all arguments
    try:
        result = subprocess.run(
            [str(binary_path)] + sys.argv[1:],
            check=False,
        )
        sys.exit(result.returncode)
    except KeyboardInterrupt:
        sys.exit(130)
    except Exception as e:
        print(f"Error running sparktype: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
