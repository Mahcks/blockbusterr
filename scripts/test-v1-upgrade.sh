#!/bin/sh
set -eu

candidate_image=${1:-blockbusterr:ci}
v1_image=${V1_IMAGE:-ghcr.io/mahcks/blockbusterr:v1.5.0}
data_dir=$(mktemp -d)
rollback_dir=$(mktemp -d)
jobs_json=$(mktemp)
suffix=$$
v1_name="blockbusterr-v1-upgrade-$suffix"
v2_name="blockbusterr-v2-upgrade-$suffix"
rollback_name="blockbusterr-v1-rollback-$suffix"

cleanup() {
  docker rm -f "$v1_name" "$v2_name" "$rollback_name" >/dev/null 2>&1 || true
  docker run --rm --entrypoint sh -v "$data_dir:/data" "$candidate_image" -c "chown -R $(id -u):$(id -g) /data" >/dev/null 2>&1 || true
  docker run --rm --entrypoint sh -v "$rollback_dir:/data" "$candidate_image" -c "chown -R $(id -u):$(id -g) /data" >/dev/null 2>&1 || true
  rm -rf "$data_dir" "$rollback_dir"
  rm -f "$jobs_json"
}
trap cleanup EXIT INT TERM

cp scripts/testdata/v1-upgrade-config.yaml "$data_dir/config.yaml"

docker run -d --network none --name "$v1_name" -v "$data_dir:/app/data" "$v1_image" >/dev/null
i=0
until docker exec "$v1_name" wget -qO- http://127.0.0.1:9090/v1 >/dev/null 2>&1; do
  sleep 1
  i=$((i + 1))
  [ "$i" -lt 60 ] || {
    docker logs "$v1_name"
    echo "$v1_image did not become ready" >&2
    exit 1
  }
done
docker stop -t 15 "$v1_name" >/dev/null
docker run --rm --entrypoint sh -v "$data_dir:/data" "$candidate_image" -c '
  chmod 0775 /data
  for path in /data/blockbusterr.db*; do
    [ ! -e "$path" ] || chmod 0666 "$path"
  done
'

before=$(python3 - "$data_dir/blockbusterr.db" <<'PY'
import sqlite3, sys
db = sqlite3.connect(sys.argv[1])
# Preserve actual history, not just an empty database, through upgrade and rollback.
run_id = db.execute("INSERT INTO job_runs (started_at, finished_at, job_id, job_name, media_type, status) VALUES (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'upgrade-fixture', 'Upgrade history fixture', 'movie', 'completed')").lastrowid
db.execute("INSERT INTO activity_logs (run_id, job_id, job_type, media_type, title, status) VALUES (?, 'upgrade-fixture', 'Trending Movies', 'movie', 'Upgrade history fixture', 'skipped')", (run_id,))
db.commit()
tables = {row[0] for row in db.execute("SELECT name FROM sqlite_master WHERE type='table'")}
print(":".join(str(db.execute(f"SELECT COUNT(*) FROM {table}").fetchone()[0]) if table in tables else "0" for table in ("activity_logs", "job_runs")))
PY
)
docker run --rm --entrypoint sh -v "$data_dir:/data" "$candidate_image" -c '
  for path in /data/blockbusterr.db*; do
    [ ! -e "$path" ] || chmod 0644 "$path"
  done
'

docker run -d --network none --name "$v2_name" -e DATA_DIR=/app/data -e BLOCKBUSTERR_DRY_RUN=false -v "$data_dir:/app/data" "$candidate_image" >/dev/null
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
# Roll back into a separate empty volume; never reuse the migrated v2 database.
docker stop -t 15 "$v2_name" >/dev/null
docker run --rm --entrypoint sh -v "$data_dir:/upgraded:ro" -v "$rollback_dir:/rollback" "$candidate_image" -c '
  cp -a /upgraded/backups/pre-v2/. /rollback/
  cmp /upgraded/backups/pre-v2/config.yaml /rollback/config.yaml
  cmp /upgraded/backups/pre-v2/blockbusterr.db /rollback/blockbusterr.db
'
docker run -d --network none --name "$rollback_name" -v "$rollback_dir:/app/data" "$v1_image" >/dev/null
i=0
until docker exec "$rollback_name" wget -qO- http://127.0.0.1:9090/v1 >/dev/null 2>&1; do
  i=$((i + 1))
  [ "$i" -lt 60 ] || {
    docker logs "$rollback_name"
    echo "$v1_image did not become ready after restoring the pre-v2 snapshot" >&2
    exit 1
  }
  sleep 1
done
docker stop -t 15 "$rollback_name" >/dev/null
cmp "$data_dir/backups/pre-v2/config.yaml" "$rollback_dir/config.yaml" || {
  echo "rollback changed the original configuration" >&2
  exit 1
}
python3 - "$data_dir/backups/pre-v2/blockbusterr.db" "$rollback_dir/blockbusterr.db" <<'PYTHON'
import sqlite3, sys
with sqlite3.connect(f"file:{sys.argv[1]}?mode=ro", uri=True) as before:
    with sqlite3.connect(f"file:{sys.argv[2]}?mode=ro", uri=True) as after:
        assert after.execute("PRAGMA integrity_check").fetchone()[0] == "ok", "rollback database is corrupt"
        # v1 legitimately appends an initial Job Run on startup. Every original
        # user-table row must still exist unchanged, including our seeded history.
        for (table,) in before.execute("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'"):
            quote = lambda name: '"' + name.replace('"', '""') + '"'
            columns = ",".join(quote(row[1]) for row in before.execute(f"PRAGMA table_info({quote(table)})"))
            query = f"SELECT {columns} FROM {quote(table)}"
            assert set(before.execute(query)) <= set(after.execute(query)), f"rollback lost or changed rows in {table}"
PYTHON
echo "$v1_image -> $candidate_image upgrade, restart, and snapshot rollback passed without intervention"
