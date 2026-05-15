# Build Guide

Diese Dokumentation beschreibt alle Build-Moeglichkeiten fuer Schematics2 (lokal und Docker).

## Voraussetzungen

- Go 1.26+
- Node.js 22+ und npm
- Docker (optional, fuer Container-Build)

## Build-Artefakte

- Backend Binary (Windows): backend/bin/schematics2.exe
- Backend Binary (Linux): backend/bin/schematics2
- Frontend Dist: frontend/dist
- Embedded Frontend fuer Backend: backend/internal/webclient/dist
- TLS-Zertifikate: backend/configs/cert/server.crt und backend/configs/cert/server.key

## Zertifikate erzeugen

Das Kommando erzeugt ein selbstsigniertes Zertifikat (Public/Private) anhand von backend/configs/service.yaml.

```bash
cd backend
go run ./cmd/gencert
```

Optional mit expliziten Pfaden:

```bash
go run ./cmd/gencert -config configs/service.yaml -cert configs/cert/server.crt -key configs/cert/server.key
```

## Frontend Build

```bash
cd frontend
npm run build
```

Hinweis:
- Nach dem Build laeuft automatisch postbuild.
- postbuild kopiert frontend/dist nach backend/internal/webclient/dist.

## Backend Build

```bash
cd backend
go build -ldflags="-s -w" -o ./bin/schematics2 ./cmd/server
```

## Komplett-Build (Windows)

```bat
cd backend
scripts\build.cmd
```

Ablauf:
1. Frontend Build
2. Zertifikatserzeugung (cmd/gencert)
3. Backend Binary Build

## Komplett-Build (Linux/macOS)

```bash
cd backend
./scripts/build.sh
```

Ablauf:
1. Frontend Build
2. Zertifikatserzeugung (cmd/gencert)
3. Backend Binary Build

## Build mit Make (Backend-Skripte)

```bash
cd backend/scripts
make gencert
make build
```

Wichtiger Hinweis:
- Das vorhandene backend/scripts/Makefile stammt in Teilen noch aus einem aelteren Projekt und enthaelt derzeit Targets/Binaernamen, die nicht komplett zu schematics2 passen.
- Fuer den regulaeren Build sollten bevorzugt scripts/build.cmd oder scripts/build.sh genutzt werden.

## Docker Build

Das Dockerfile baut Frontend und Backend in Multi-Stage und erzeugt im Build auch Zertifikate.

```bash
cd backend
docker build -f ./build/package/Dockerfile ../ -t schematics2:latest
```

Warum ../ als Context:
- Das Dockerfile benoetigt sowohl backend/ als auch frontend/.

Container starten:

```bash
docker run -p 8080:8080 -p 8443:8443 -v "${PWD}/configs:/app/configs" schematics2:latest
```

PowerShell Beispiel:

```powershell
docker run -p 8080:8080 -p 8443:8443 -v ${PWD}\configs:/app/configs schematics2:latest
```

## Healthcheck

Im Container ist folgender Healthcheck konfiguriert:
- GET http://localhost:8080/health

## Typischer Workflow

1. Abhaengigkeiten installieren:
   - frontend: npm install
   - backend: go mod tidy
2. Komplettbuild lokal:
   - Windows: backend/scripts/build.cmd
   - Linux/macOS: backend/scripts/build.sh
3. Optional Docker Image bauen:
   - docker build -f backend/build/package/Dockerfile . -t schematics2:latest (vom Repo-Root)
