#!/bin/bash

# Safety settings
set -o errexit   # Exit immediately on any error
set -o nounset   # Treat unset variables as error
set -o pipefail  # Catch errors in pipelines

FILE="getAppInfo.sh"

# Try to source getAppInfo.sh from current dir or ./scripts
if [[ -f "$FILE" ]]; then
  echo "## Will execute $FILE"
  # shellcheck disable=SC1090
  source "$FILE"
elif [[ -f "./scripts/$FILE" ]]; then
  echo "## Will execute ./scripts/$FILE"
  # shellcheck disable=SC1090
  source "./scripts/$FILE"
else
  echo "## 💥💥 ERROR: $FILE was not found in current directory or ./scripts/"
  exit 1
fi

# At this point, getAppInfo.sh should have set APP_NAME and APP_VERSION
if [[ -z "${APP_NAME:-}" || -z "${APP_VERSION:-}" ]]; then
  echo "## 💥💥 ERROR: APP_NAME or APP_VERSION not set by $FILE"
  exit 1
fi

echo "## APP: ${APP_NAME}, version: ${APP_VERSION} detected"

# Check if the working tree is clean
if output=$(git status --porcelain) && [[ -z "$output" ]]; then
  echo "## Git working tree is clean ✓"

  # Check if the tag already exists
  if git tag -l "v${APP_VERSION}" | grep -q "^v${APP_VERSION}$"; then
    echo "## 💥💥 ERROR: Tag 'v${APP_VERSION}' already exists!"
    echo "    Remove it manually if needed: git tag -d v${APP_VERSION} && git push origin --delete v${APP_VERSION}"
    exit 1
  else
    echo "## ✓🚀 Tag 'v${APP_VERSION}' does not exist — proceeding to create it"

    # Create the tag
    git tag -a "v${APP_VERSION}" -m "Release v${APP_VERSION}"

    # Update build stamp and revision in version.go
    if [[ -x "./scripts/setAppBuildStampInfo.sh" ]]; then
      "./scripts/setAppBuildStampInfo.sh"
    else
      echo "## 💥💥 ERROR: ./scripts/setAppBuildStampInfo.sh not found or not executable"
      exit 1
    fi

    # Commit the updated version file
    git add pkg/version/version.go
    git commit -m "chore: update build info for v${APP_VERSION}"

    # Push everything
    echo "## Pushing main branch..."
    git push origin main

    echo "## Pushing tags..."
    git push origin --tags

    echo "## 🎉 Successfully created and pushed tag v${APP_VERSION}"
  fi
else
  echo "## 💥💥 ERROR: Git working tree is DIRTY!"
  echo "    You must commit or stash all changes before creating a release tag."
  echo ""
  echo "Current status:"
  git status
  exit 1
fi
