#!/bin/sh
set -u

readonly app_uid=10001
readonly app_gid=10001
data_dir=${DATA_DIR:-/app/data}

if [ "$(id -u)" = "0" ]; then
  case "$data_dir" in
    ""|/|/app)
      echo "Blockbusterr refused to repair unsafe DATA_DIR: $data_dir" >&2
      exit 1
      ;;
  esac

  if ! mkdir -p "$data_dir" || ! chown -h "$app_uid" "$data_dir"; then
    echo "Blockbusterr could not assign $data_dir to UID $app_uid; continuing so startup can report the exact storage error." >&2
  fi

  for name in config.yaml config.yml blockbusterr.db blockbusterr.db-journal blockbusterr.db-wal blockbusterr.db-shm; do
    path="$data_dir/$name"
    if [ -e "$path" ] && ! chown -h "$app_uid" "$path"; then
      echo "Blockbusterr could not assign $path to UID $app_uid." >&2
    fi
  done

  exec su-exec "$app_uid:$app_gid" "$@"
fi

exec "$@"
