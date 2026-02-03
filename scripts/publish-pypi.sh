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

# Delay between uploads in seconds
UPLOAD_DELAY=10

echo "Building Python wheels for sparktype v${VERSION}..."

# Create wheels directory
mkdir -p "$WHEELS_DIR"

# Build platform-specific wheels
# Note: build dependencies (build, wheel, twine) should be installed by the CI workflow
python3 "${PYPI_DIR}/scripts/build-wheels.py" "$VERSION" "$DIST_DIR" --output "$WHEELS_DIR"

# Upload to PyPI
echo "Uploading wheels to PyPI..."
if [ -z "$PYPI_TOKEN" ]; then
    echo "::error::PYPI_TOKEN is not set. Cannot publish to PyPI."
    echo "Please ensure the PYPI_TOKEN secret is configured in the repository settings."
    exit 1
fi

# Upload wheels sequentially with delays to avoid rate limiting
FIRST=true
for WHEEL in "$WHEELS_DIR"/*.whl; do
    if [ ! -f "$WHEEL" ]; then
        echo "::error::No wheel files found in $WHEELS_DIR"
        exit 1
    fi

    # Add delay between uploads (except for the first one)
    if [ "$FIRST" = true ]; then
        FIRST=false
    else
        echo "⏳ Waiting ${UPLOAD_DELAY}s before next upload..."
        sleep "$UPLOAD_DELAY"
    fi

    WHEEL_NAME=$(basename "$WHEEL")
    echo "📦 Uploading $WHEEL_NAME..."

    # Retry logic for uploads
    MAX_RETRIES=3
    RETRY=0
    while [ $RETRY -lt $MAX_RETRIES ]; do
        if python3 -m twine upload \
            --username __token__ \
            --password "$PYPI_TOKEN" \
            --non-interactive \
            --skip-existing \
            "$WHEEL" 2>&1; then
            echo "✅ Uploaded $WHEEL_NAME"
            break
        else
            RETRY=$((RETRY + 1))
            if [ $RETRY -lt $MAX_RETRIES ]; then
                DELAY=$((RETRY * 15))
                echo "⏳ Upload failed, retrying in ${DELAY}s (attempt $RETRY/$MAX_RETRIES)..."
                sleep "$DELAY"
            else
                echo "::error::Failed to upload $WHEEL_NAME after $MAX_RETRIES attempts"
                exit 1
            fi
        fi
    done
done

echo "Successfully published wheels to PyPI!"
