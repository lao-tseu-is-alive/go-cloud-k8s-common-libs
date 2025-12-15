#!/bin/bash

# Safety options
set -o errexit   # Exit on any command failure
set -o nounset   # Treat unset variables as error
set -o pipefail  # Catch errors in pipelines

SOURCE_CODE="pkg/version/version.go"

if [[ ! -f "$SOURCE_CODE" ]]; then
  echo "## 💥💥 ERROR: Source file '${SOURCE_CODE}' not found!"
  exit 1
fi

echo "## Updating build information in ${SOURCE_CODE}"

# Get current timestamp and git revision
NOW=$(date +%Y-%m-%dT%H:%M:%S%z)  # More precise format (includes seconds and timezone)
REVISION=$(git describe --dirty --always)

# Update the constants in the Go file
# Uses safer sed syntax that works on both GNU and macOS sed
sed -i.bak "s/^\([[:space:]]*BuildStamp[[:space:]]*=[[:space:]]*\).*/\1\"$NOW\"/" "$SOURCE_CODE" && rm "${SOURCE_CODE}.bak"
sed -i.bak "s/^\([[:space:]]*REVISION[[:space:]]*=[[:space:]]*\).*/\1\"$REVISION\"/" "$SOURCE_CODE" && rm "${SOURCE_CODE}.bak"

# Reformat the file
gofmt -w "$SOURCE_CODE"

echo "## Extracting app metadata from ${SOURCE_CODE}"

# Extract values using safer grep + awk patterns
APP_NAME=$(awk -F '=' '/APP[[:space:]]*=[[:space:]]*/ {gsub(/[[:space:]]|"|;/, "", $2); print $2}' "$SOURCE_CODE")
APP_VERSION=$(awk -F '=' '/VERSION[[:space:]]*=[[:space:]]*/ {gsub(/[[:space:]]|"|;/, "", $2); print $2}' "$SOURCE_CODE")
APP_REVISION=$(awk -F '=' '/REVISION[[:space:]]*=[[:space:]]*/ {gsub(/[[:space:]]|"|;/, "", $2); print $2}' "$SOURCE_CODE")
APP_BuildStamp=$(awk -F '=' '/BuildStamp[[:space:]]*=[[:space:]]*/ {gsub(/[[:space:]]|"|;/, "", $2); print $2}' "$SOURCE_CODE")

# Check if extraction succeeded
if [[ -z "$APP_NAME" || -z "$APP_VERSION" ]]; then
  echo "## 💥💥 WARNING: Could not extract APP_NAME or VERSION from ${SOURCE_CODE}"
  echo "    Check if the constants are defined as expected (e.g., const APP = \"myapp\")"
  exit 1
fi

echo "## Found:"
echo "   APP          : ${APP_NAME}"
echo "   VERSION      : ${APP_VERSION}"
echo "   REVISION     : ${APP_REVISION}"
echo "   BuildStamp   : ${APP_BuildStamp}"
