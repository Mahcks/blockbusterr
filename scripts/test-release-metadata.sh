#!/bin/sh
set -eu

validator=$(dirname "$0")/validate-release-metadata.sh

[ "$($validator v2.0.0-beta.7 true)" = latest-beta ]
[ "$($validator v2.0.0-rc.1 true)" = latest-beta ]
[ "$($validator v2.0.0 false)" = latest ]

for invalid in \
  'v2.0.0-beta.7 false' \
  'v2.0.0-rc.1 false' \
  'v2.0.0 true' \
  '2.0.0 false' \
  'v2.0 false'; do
  set -- $invalid
  if "$validator" "$1" "$2" >/dev/null 2>&1; then
    echo "accepted invalid release metadata: $invalid" >&2
    exit 1
  fi
done
