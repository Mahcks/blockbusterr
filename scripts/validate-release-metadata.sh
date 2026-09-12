#!/bin/sh
set -eu

tag=${1:-}
prerelease=${2:-}

if printf '%s' "$tag" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+-(beta|rc)\.[0-9]+$'; then
  [ "$prerelease" = true ] || {
    echo "Prerelease tag $tag must be published as a prerelease." >&2
    exit 1
  }
  echo latest-beta
elif printf '%s' "$tag" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
  [ "$prerelease" = false ] || {
    echo "Stable tag $tag must not be published as a prerelease." >&2
    exit 1
  }
  echo latest
else
  echo "Unsupported release tag: $tag" >&2
  exit 1
fi
