#!/usr/bin/env python3
"""
Build platform-specific wheels for sparktype.

This script takes pre-built binaries from goreleaser and packages them
into platform-specific Python wheels.

Usage:
    python build-wheels.py <version> <dist_dir>

Where:
    <version>  - The version being released (e.g., "0.1.0")
    <dist_dir> - Directory containing goreleaser output binaries
"""

import argparse
import json
import os
import shutil
import stat
import subprocess
import sys
import tempfile
from pathlib import Path


def load_platforms_config():
    """Load shared platform configuration."""
    # Config is at scripts/platforms.json relative to repo root
    script_dir = Path(__file__).parent
    repo_root = script_dir.parent.parent.parent
    config_path = repo_root / "scripts" / "platforms.json"

    with open(config_path) as f:
        return json.load(f)


def get_archive_name(version: str, goos: str, goarch: str) -> str:
    """Get the goreleaser archive name for a platform."""
    ext = "zip" if goos == "windows" else "tar.gz"
    return f"sparktype_{version}_{goos}_{goarch}.{ext}"


def extract_binary(archive_path: Path, goos: str, dest_dir: Path) -> Path:
    """Extract the binary from an archive."""
    import tarfile
    import zipfile

    binary_name = "sparktype.exe" if goos == "windows" else "sparktype"

    if archive_path.suffix == ".zip" or str(archive_path).endswith(".zip"):
        with zipfile.ZipFile(archive_path, "r") as zf:
            for name in zf.namelist():
                if name.endswith(binary_name):
                    extracted = dest_dir / binary_name
                    with zf.open(name) as src, open(extracted, "wb") as dst:
                        dst.write(src.read())
                    return extracted
    else:
        with tarfile.open(archive_path, "r:gz") as tf:
            for member in tf.getmembers():
                if member.name.endswith(binary_name) and member.isfile():
                    extracted = dest_dir / binary_name
                    with tf.extractfile(member) as src, open(extracted, "wb") as dst:
                        dst.write(src.read())
                    return extracted

    raise FileNotFoundError(f"Binary not found in archive: {archive_path}")


def build_wheel(
    version: str,
    platform_tag: str,
    binary_path: Path,
    binary_name: str,
    pypi_dir: Path,
    output_dir: Path,
) -> Path:
    """Build a wheel for a specific platform."""
    # Ensure output_dir is absolute to avoid issues with cwd changes
    output_dir = output_dir.resolve()

    with tempfile.TemporaryDirectory() as tmpdir:
        tmpdir = Path(tmpdir)

        # Copy package structure
        pkg_dir = tmpdir / "sparktype"
        shutil.copytree(pypi_dir / "sparktype", pkg_dir)

        # Create bin directory and copy binary
        bin_dir = pkg_dir / "bin"
        bin_dir.mkdir(exist_ok=True)
        dest_binary = bin_dir / binary_name
        shutil.copy2(binary_path, dest_binary)

        # Make binary executable on Unix
        if not binary_name.endswith(".exe"):
            dest_binary.chmod(
                dest_binary.stat().st_mode | stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH
            )

        # Copy pyproject.toml and update version
        pyproject_content = (pypi_dir / "pyproject.toml").read_text()
        # Replace the placeholder version
        pyproject_content = pyproject_content.replace(
            'version = "0.0.0-development"', f'version = "{version}"'
        )
        # Also handle the old format in case it wasn't updated
        pyproject_content = pyproject_content.replace(
            'version = "0.1.0"', f'version = "{version}"'
        )
        (tmpdir / "pyproject.toml").write_text(pyproject_content)

        # Copy README
        readme_src = pypi_dir / "README.md"
        if readme_src.exists():
            shutil.copy2(readme_src, tmpdir / "README.md")

        # Update __version__ in __init__.py
        init_path = pkg_dir / "__init__.py"
        if init_path.exists():
            init_content = init_path.read_text()
            init_content = init_content.replace(
                '__version__ = "0.0.0-development"', f'__version__ = "{version}"'
            )
            init_path.write_text(init_content)

        # Update __version__ in cli.py
        cli_path = pkg_dir / "cli.py"
        cli_content = cli_path.read_text()
        cli_content = cli_content.replace(
            '__version__ = "0.0.0-development"', f'__version__ = "{version}"'
        )
        cli_content = cli_content.replace(
            '__version__ = "0.1.0"', f'__version__ = "{version}"'
        )
        cli_path.write_text(cli_content)

        # Build wheel into a temp directory first, then rename and move to output_dir
        env = os.environ.copy()

        # Create a temp directory for this build's output
        build_output = tmpdir / "wheel_output"
        build_output.mkdir()

        # Use pip wheel directly - simpler and avoids build isolation issues
        result = subprocess.run(
            [
                sys.executable,
                "-m",
                "pip",
                "wheel",
                "--no-deps",
                "--wheel-dir",
                str(build_output),
                str(tmpdir),
            ],
            capture_output=True,
            text=True,
            env=env,
            cwd=tmpdir,
        )

        # Always print build output for debugging
        if result.stdout.strip():
            print(f"  pip wheel stdout: {result.stdout.strip()}")
        if result.stderr.strip():
            print(f"  pip wheel stderr: {result.stderr.strip()}", file=sys.stderr)

        if result.returncode != 0:
            raise RuntimeError(
                f"Failed to build wheel for {platform_tag}: {result.stderr}"
            )

        # Find the built wheel in the temp directory
        built_wheels = list(build_output.glob("sparktype-*.whl"))
        if not built_wheels:
            print(
                f"  No wheels found in build output: {list(build_output.iterdir())}",
                file=sys.stderr,
            )
            raise RuntimeError(f"No wheel created for {platform_tag}")

        # Should be exactly one wheel
        source_wheel = built_wheels[0]
        print(f"  Built: {source_wheel.name}")

        # Rename with platform tag and move to final output directory
        new_name = source_wheel.name.replace("py3-none-any", f"py3-none-{platform_tag}")
        dest_wheel = output_dir / new_name
        shutil.move(str(source_wheel), str(dest_wheel))
        print(f"  Renamed to: {dest_wheel.name}")

        return dest_wheel


def main():
    parser = argparse.ArgumentParser(description="Build platform-specific wheels")
    parser.add_argument("version", help="Version being released")
    parser.add_argument("dist_dir", help="Directory containing goreleaser archives")
    parser.add_argument(
        "--output", "-o", default="./wheels", help="Output directory for wheels"
    )
    args = parser.parse_args()

    version = args.version.lstrip("v")
    dist_dir = Path(args.dist_dir)
    output_dir = Path(args.output)
    output_dir.mkdir(parents=True, exist_ok=True)

    # Get the pypi distribution directory
    script_dir = Path(__file__).parent
    pypi_dir = script_dir.parent

    # Load shared platform config
    config = load_platforms_config()

    print(f"Building wheels for sparktype v{version}")
    print(f"Source archives: {dist_dir}")
    print(f"Output directory: {output_dir}")

    built_wheels = []

    for platform_id, platform_config in config["platforms"].items():
        goos = platform_config["go"]["os"]
        goarch = platform_config["go"]["arch"]
        pypi_tag = platform_config["pypi"]["tag"]
        binary_name = platform_config.get("binaryName", "sparktype")

        archive_name = get_archive_name(version, goos, goarch)
        archive_path = dist_dir / archive_name

        if not archive_path.exists():
            print(f"Warning: Archive not found: {archive_path}, skipping {pypi_tag}")
            continue

        print(f"Building wheel for {pypi_tag}...")

        with tempfile.TemporaryDirectory() as tmpdir:
            tmpdir = Path(tmpdir)

            # Extract binary
            try:
                binary_path = extract_binary(archive_path, goos, tmpdir)
            except FileNotFoundError as e:
                print(f"Error: {e}", file=sys.stderr)
                continue

            # Build wheel
            try:
                wheel_path = build_wheel(
                    version,
                    pypi_tag,
                    binary_path,
                    binary_name,
                    pypi_dir,
                    output_dir,
                )
                built_wheels.append(wheel_path)
                print(f"  Built: {wheel_path.name}")
            except Exception as e:
                print(f"Error building wheel for {pypi_tag}: {e}", file=sys.stderr)
                continue

    print(f"\nBuilt {len(built_wheels)} wheels:")
    for whl in built_wheels:
        print(f"  {whl.name}")

    return 0 if built_wheels else 1


if __name__ == "__main__":
    sys.exit(main())
