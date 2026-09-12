#!/bin/sh
set -eu

image=${1:?usage: test-container-smoke.sh IMAGE}
suffix=$$
container=blockbusterr-smoke-$suffix
volume=blockbusterr-smoke-$suffix

cleanup() {
  docker rm -f "$container" >/dev/null 2>&1 || true
  docker volume rm "$volume" >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

docker volume create "$volume" >/dev/null
docker run -d --name "$container" -p 127.0.0.1::9090 -v "$volume:/app/data" "$image" >/dev/null
port=$(docker port "$container" 9090/tcp | sed 's/.*://')

i=0
until curl --fail --silent --show-error "http://127.0.0.1:$port/jobs" >/dev/null; do
  i=$((i + 1))
  if [ "$i" -ge 30 ]; then
    docker logs "$container" >&2
    exit 1
  fi
  sleep 1
done

curl --fail --silent --show-error "http://127.0.0.1:$port/static/css/app.css" >/dev/null
[ "$(docker exec "$container" awk '/^Uid:/{print $2}' /proc/1/status)" = 10001 ]
docker exec --user 10001:10001 "$container" touch /app/data/smoke-write
