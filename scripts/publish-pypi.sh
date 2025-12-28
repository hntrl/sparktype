#!/bin/bash
set -e

VERSION="$1"
if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    exit 1
fi

# Remove leading 'v' if present
VERSION="${VERSION#v}"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
DIST_DIR="${ROOT_DIR}/dist"
PYPI_DIR="${ROOT_DIR}/distributions/pypi"
WHEELS_DIR="${ROOT_DIR}/dist/wheels"

echo "Building Python wheels for sparktype v${VERSION}..."

# Create wheels directory
mkdir -p "$WHEELS_DIR"

# Build platform-specific wheels
python3 "${PYPI_DIR}/scripts/build-wheels.py" "$VERSION" "$DIST_DIR" --output "$WHEELS_DIR"

# Upload to PyPI
echo "Uploading wheels to PyPI..."
if [ -z "$PYPI_TOKEN" ]; then
    echo "::error::PYPI_TOKEN is not set. Cannot publish to PyPI."
    echo "Please ensure the PYPI_TOKEN secret is configured in the repository settings."
    exit 1
fi

python3 -m pip install --quiet twine

# Upload all wheels
python3 -m twine upload \
    --username __token__ \
    --password "$PYPI_TOKEN" \
    --non-interactive \
    "$WHEELS_DIR"/*.whl

echo "Successfully published wheels to PyPI!"

