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
1. **Backup current config first!**
```sh
cp data/config.yaml data/config.yaml.backup
```

2. **Pull and restart**
```sh
docker pull ghcr.io/mahcks/blockbusterr:vX.Y.Z
docker-compose down && docker-compose up -d
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
