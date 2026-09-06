# vX.Y.Z

## New Features
- 

## Fixes
- 

## Improvements
- 

## Breaking Changes
- 

## Configuration
- 

## Upgrading
1. **Stop Blockbusterr and back up its complete data directory.** The configuration and SQLite database must come from the same stopped snapshot.
```sh
docker stop blockbusterr
cp -a data "data.backup-$(date +%Y%m%d-%H%M%S)"
```

2. **Pull and restart**
```sh
docker pull ghcr.io/mahcks/blockbusterr:vX.Y.Z
docker compose down && docker compose up -d
```

*Or for docker run users*
```sh
docker stop blockbusterr
docker rm blockbusterr
docker run -d \\
  --name blockbusterr \\
  -p 9090:9090 \\
  -v $(pwd)/data:/app/data \\
  ghcr.io/mahcks/blockbusterr:vX.Y.Z
```

---

_Replace X.Y.Z with your version. Add or remove sections as needed._
