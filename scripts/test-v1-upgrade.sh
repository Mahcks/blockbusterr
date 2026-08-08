#!/bin/sh
set -eu

candidate_image=${1:-blockbusterr:ci}
v1_image=${V1_IMAGE:-ghcr.io/mahcks/blockbusterr:v1.5.0}
data_dir=$(mktemp -d)
jobs_json=$(mktemp)
suffix=$$
v1_name="blockbusterr-v1-upgrade-$suffix"
v2_name="blockbusterr-v2-upgrade-$suffix"

cleanup() {
  docker rm -f "$v1_name" "$v2_name" >/dev/null 2>&1 || true
  docker run --rm --entrypoint sh -v "$data_dir:/data" "$candidate_image" -c "chown -R $(id -u):$(id -g) /data" >/dev/null 2>&1 || true
  rm -rf "$data_dir"
  rm -f "$jobs_json"
}
trap cleanup EXIT INT TERM

cp scripts/testdata/v1-upgrade-config.yaml "$data_dir/config.yaml"

docker run -d --name "$v1_name" -v "$data_dir:/app/data" "$v1_image" >/dev/null
i=0
while [ ! -f "$data_dir/blockbusterr.db" ] && [ "$i" -lt 30 ]; do
  sleep 1
  i=$((i + 1))
done
[ -f "$data_dir/blockbusterr.db" ] || {
  docker logs "$v1_name"
  echo "v1 did not initialize its database" >&2
  exit 1
}
docker stop -t 15 "$v1_name" >/dev/null
docker run --rm --entrypoint sh -v "$data_dir:/data" "$candidate_image" -c 'chmod 0775 /data && chmod 0666 /data/blockbusterr.db'

before=$(python3 - "$data_dir/blockbusterr.db" <<'PY'
import sqlite3, sys
db = sqlite3.connect(sys.argv[1])
tables = {row[0] for row in db.execute("SELECT name FROM sqlite_master WHERE type='table'")}
print(":".join(str(db.execute(f"SELECT COUNT(*) FROM {table}").fetchone()[0]) if table in tables else "0" for table in ("activity_logs", "job_runs")))
PY
)
docker run --rm --entrypoint sh -v "$data_dir:/data" "$candidate_image" -c 'chmod 0644 /data/blockbusterr.db'

docker run -d --name "$v2_name" -e DATA_DIR=/app/data -e BLOCKBUSTERR_DRY_RUN=false -v "$data_dir:/app/data" "$candidate_image" >/dev/null
i=0
until docker exec "$v2_name" wget -qO- http://127.0.0.1:9090/v1/jobs/list >"$jobs_json" 2>/dev/null; do
  i=$((i + 1))
  if [ "$i" -ge 60 ]; then
    docker logs "$v2_name"
    echo "v2 did not become ready after upgrading v1 data" >&2
    exit 1
  fi
  sleep 1
done

after=$(python3 - "$data_dir/blockbusterr.db" <<'PY'
import sqlite3, sys
db = sqlite3.connect(f"file:{sys.argv[1]}?mode=ro", uri=True)
print(":".join(str(db.execute(f"SELECT COUNT(*) FROM {table}").fetchone()[0]) for table in ("activity_logs", "job_runs")))
PY
)

[ "$before" = "$after" ] || {
  echo "startup unexpectedly changed activity/job-run counts: $before -> $after" >&2
  exit 1
}
grep -q '"id":"trending_movies"' "$jobs_json" || {
  echo "migrated job is missing from the v2 API" >&2
  exit 1
}
grep -q '"limit":37' "$jobs_json" || {
  echo "migrated job limit was not preserved" >&2
  exit 1
}
docker exec "$v2_name" sh -c '
  test -f /app/data/backups/pre-v2/blockbusterr.db &&
  test -f /app/data/config.yaml.backup &&
  cmp /app/data/config.yaml.backup /app/data/backups/pre-v2/config.yaml &&
  grep -q "rule_set_id: default-movies" /app/data/config.yaml &&
  grep -q "Horror" /app/data/config.yaml &&
  grep -q "Reality" /app/data/config.yaml &&
  grep -q "global_limit_movies: 7" /app/data/config.yaml &&
  grep -q "upgrade-fixture-client" /app/data/config.yaml &&
  grep -q "upgrade-fixture-secret" /app/data/config.yaml
' || {
  echo "automatic backup or migrated configuration verification failed" >&2
  docker logs "$v2_name"
  exit 1
}

docker exec "$v2_name" sh -c 'test "$(awk '\''/^Uid:/{print $2}'\'' /proc/1/status)" = 10001 && test "$(awk '\''/^Gid:/{print $2}'\'' /proc/1/status)" = 10001'
backup_hash=$(docker exec "$v2_name" sha256sum /app/data/backups/pre-v2/blockbusterr.db /app/data/backups/pre-v2/config.yaml)
docker restart "$v2_name" >/dev/null
i=0
until docker exec "$v2_name" wget -qO- http://127.0.0.1:9090/v1/jobs/list >/dev/null 2>&1; do
  i=$((i + 1))
  [ "$i" -lt 60 ] || {
    docker logs "$v2_name"
    echo "v2 did not become ready after its second start" >&2
    exit 1
  }
  sleep 1
done
[ "$backup_hash" = "$(docker exec "$v2_name" sha256sum /app/data/backups/pre-v2/blockbusterr.db /app/data/backups/pre-v2/config.yaml)" ] || {
  echo "second startup modified the pre-v2 backup" >&2
  exit 1
}
[ "$after" = "$(python3 - "$data_dir/blockbusterr.db" <<'PY'
import sqlite3, sys
db = sqlite3.connect(f"file:{sys.argv[1]}?mode=ro", uri=True)
print(":".join(str(db.execute(f"SELECT COUNT(*) FROM {table}").fetchone()[0]) for table in ("activity_logs", "job_runs")))
PY
)" ] || {
  echo "second startup changed activity or job-run history" >&2
  exit 1
}
echo "v1.5.0 -> v2 upgrade passed without intervention"
