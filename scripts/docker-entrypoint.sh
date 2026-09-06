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

  backup_root="$data_dir/backups"
  backup_dir="$backup_root/pre-v2"
  upgrade_marker="$backup_root/.v2-initialized"
  if [ -f "$data_dir/blockbusterr.db" ] && [ ! -e "$upgrade_marker" ] && [ ! -e "$backup_dir" ]; then
    backup_tmp="$backup_root/.pre-v2-$$"
    mkdir -p "$backup_tmp" || {
      echo "Blockbusterr could not create the automatic pre-v2 backup; refusing to migrate." >&2
      exit 1
    }
    for name in config.yaml config.yml blockbusterr.db blockbusterr.db-journal blockbusterr.db-wal blockbusterr.db-shm; do
      path="$data_dir/$name"
      [ ! -e "$path" ] || cp -p "$path" "$backup_tmp/$name" || {
        echo "Blockbusterr could not back up $path; refusing to migrate." >&2
        exit 1
      }
    done
    mv "$backup_tmp" "$backup_dir" || {
      echo "Blockbusterr could not finalize the automatic pre-v2 backup; refusing to migrate." >&2
      exit 1
    }
    echo "Blockbusterr saved the pre-v2 data snapshot to $backup_dir."
  fi
  mkdir -p "$backup_root" && touch "$upgrade_marker" || {
    echo "Blockbusterr could not record completed v2 storage initialization." >&2
    exit 1
  }

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
