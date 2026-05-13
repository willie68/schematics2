# Docker Build & Deployment Guide

Dokumentation für das Build-System der Schematic2 Backend-Komponente.

## Überblick

Das Projekt wird als Docker-Image gebaut und kann in zwei Szenarien deployt werden:

1. **Direkter Zugriff**: Container wird direkt angesteuert (z.B. auf Port 9743)
2. **Reverse-Proxy Deployment**: Container wird über Apache Reverse-Proxy mit Basis-Path `/schematics2` angesteuert

Der Build-Prozess wird durch das Script `buildDocker.cmd` gesteuert.

---

## Quick Start

### Szenario 1: Direkter Zugriff (Standard)

```powershell
cd backend\scripts
.\buildDocker.cmd
```

**Resultat:**
- Frontend wird mit `BASE_PATH=/client` gebaut
- Image: `mcs/schematics2:latest`
- Zugriff: `https://192.168.178.14:9743/client`

### Szenario 2: Reverse-Proxy Deployment

```powershell
cd backend\scripts
.\buildDocker.cmd /schematics2
```

**Resultat:**
- Frontend wird mit `BASE_PATH=/schematics2` gebaut
- Image: `mcs/schematics2:latest`
- Zugriff: `https://wkla.no-ip.biz/schematics2/client`
- Registry: `192.168.178.14:5000/mcs/schematics2:latest`

---

## Was passiert beim Build?

### 1. TLS Certificate Generierung
```
Step 1: Generating TLS certificate
```
- Erstellt/erneuert `configs/cert/server.crt` und `configs/cert/server.key`
- Gültig für 10 Jahre
- Subject: `schematic2` (oder `CN=schematic2`)
- Zusätzliche DNS-Namen: `host.docker.internal`, `localhost`

### 2. Build-Informationen sammeln
```
Build Information:
  BUILD_TIME: 2026-05-13T17:58:29.123+02:00
  VCS_REF: abc1234
  BASE_PATH: /client (oder /schematics2)
```
- `BUILD_TIME`: ISO 8601 Timestamp (UTC)
- `VCS_REF`: Git Commit-Hash (short form)
- `BASE_PATH`: Frontend Base-Path für Asset-Loading

### 3. Docker Multi-Stage Build

#### Stage 1: Frontend Builder (node:22-alpine)
```
- Installiert npm dependencies
- Baut Frontend mit BASE_PATH
- Output: dist/ Verzeichnis
```

#### Stage 2: Backend Builder (golang:1.26-alpine3.22)
```
- Lädt Go dependencies
- Kopiert Frontend dist/ von Stage 1
- Generiert TLS Certificate
- Baut Go Binary mit ldflags Injection
  - VERSION=0.2.14 (aus internal/version/version.go)
  - BUILD_TIME={timestamp}
  - COMMIT={git hash}
```

#### Stage 3: Runtime (alpine:3.22)
```
- Minimal Image
- Kopiert Binary + Configs
- HEALTHCHECK: /readyz Endpoint alle 30s
- EXPOSE: 8080 (HTTP), 8443 (HTTPS)
```

### 4. Registry Push
```powershell
docker tag mcs/schematics2:latest 192.168.178.14:5000/mcs/schematics2:latest
docker push 192.168.178.14:5000/mcs/schematics2:latest
```
- Pushed zu lokaler Docker Registry (HTTP, insecure)
- Auf Ubuntu: `docker pull 192.168.178.14:5000/mcs/schematics2:latest`

---

## Versions-Management

### Version erhöhen vor Build

1. **Datei: `backend/internal/version/version.go`**
   ```go
   var (
       Version = "0.2.14"  // ← Hier erhöhen
       BuildTime = ""
       Commit = ""
   )
   ```

2. **Datei: `HISTORY.md`**
   ```markdown
   ## 0.2.14 - 2026-05-13
   
   - **Backend**: Build & Deployment improvements
   - **Frontend**: Reverse-Proxy support
   ```

3. **Git Commit**
   ```powershell
   git add .
   git commit -m "Release: v0.2.14"
   git push
   ```

4. **Build starten**
   ```powershell
   cd backend\scripts
   .\buildDocker.cmd /schematics2  # oder /client für direkten Zugriff
   ```

Die neue Version wird automatisch:
- In `cmd/server/main.go` beim Start geloggt
- Im API Endpoint `/api/v1/info` ausgeliefert
- Im Docker Image Tags dokumentiert (über LABEL)

---

## Deployment auf Ubuntu Server

### 1. Image vom Registry pullen

```bash
ssh ubuntu@192.168.178.14
docker pull 192.168.178.14:5000/mcs/schematics2:latest
```

### 2. Container starten

**Für Reverse-Proxy (empfohlen):**
```bash
docker run -d \
  --name schematic2 \
  -p 8080:8080 \
  -p 8443:8443 \
  -e CLIENT_BASE_PATH=/schematics2 \
  -v /path/to/configs:/app/configs \
  -v /path/to/data:/app/data \
  192.168.178.14:5000/mcs/schematics2:latest
```

**Für direkten Zugriff:**
```bash
docker run -d \
  --name schematic2 \
  -p 8080:8080 \
  -p 8443:8443 \
  -v /path/to/configs:/app/configs \
  -v /path/to/data:/app/data \
  192.168.178.14:5000/mcs/schematics2:latest
```

### 3. Health Check überprüfen

```bash
docker ps  # Status: healthy?

# Manuell testen
curl -k https://localhost:8443/readyz
curl -k https://localhost:8443/livez
```

### 4. Apache Reverse-Proxy konfigurieren

```apache
<VirtualHost *:443>
    ServerName wkla.no-ip.biz
    SSLEngine on
    
    # Zu schematic2 Container
    ProxyPreserveHost On
    ProxyPass /schematics2 http://192.168.178.14:8080/
    ProxyPassReverse /schematics2 http://192.168.178.14:8080/
    
    # WebSocket Support
    RewriteEngine On
    RewriteCond %{HTTP:Upgrade} websocket [NC]
    RewriteCond %{HTTP:Connection} upgrade [NC]
    RewriteRule ^/schematics2/(.*) http://192.168.178.14:8080/$1 [P,L]
</VirtualHost>
```

---

## Troubleshooting

### Container ist nicht healthy

```bash
# Logs ansehen
docker logs schematic2

# Health Status prüfen
docker inspect schematic2 | grep -A 20 "Health"
```

**Häufige Probleme:**
- `mongodb` nicht erreichbar: `MONGO_HOSTS` env var prüfen
- `files` Verzeichnis nicht vorhanden: `chown` für `/app/configs` prüfen
- Port bereits belegt: `docker ps` zeigt Konflikte

### Frontend lädt nicht oder ist weiß

1. **Direkter Zugriff**: Assets sollten unter `/client/assets/` sein
2. **Reverse-Proxy**: Assets sollten unter `/schematics2/client/assets/` sein
3. **Browser Console**: Fehler in Network Tab prüfen
4. **vite.config.js**: Richtige `BASE_PATH` Umgebungsvariable beim Build?

### Version stimmt nicht überein

- `docker inspect mcs/schematics2:latest | grep -A 5 "Labels"`
- API: `curl https://localhost:8443/api/v1/info`
- Logs: `docker logs schematic2 | grep "starting schematic2 backend"`

---

## Umgebungsvariablen

### Optionale Laufzeit-Konfiguration

```bash
docker run -e MONGO_HOSTS=host:27017 \
           -e MONGO_DATABASE=schematics \
           -e LOG_LEVEL=debug \
           ... mcs/schematics2:latest
```

Siehe `backend/configs/service.yaml` für alle verfügbaren Optionen.

---

## CI/CD Integration

Das buildDocker.cmd kann in CI/CD Pipelines integriert werden:

```yaml
# Beispiel für GitHub Actions
- name: Build & Push schematic2
  run: |
    cd backend/scripts
    .\buildDocker.cmd /schematics2
  shell: pwsh
```

Build-Metadaten werden automatisch injiziert:
- `--build-arg BUILD_TIME`: Aktueller Timestamp (UTC)
- `--build-arg VCS_REF`: Git Commit-Hash
- `--build-arg BASE_PATH`: Je nach Deployment-Szenario

---

## Weitere Ressourcen

- [Dockerfile](../build/package/Dockerfile)
- [Version Management](../internal/version/version.go)
- [Service Main](../cmd/server/main.go)
- [HISTORY](../../HISTORY.md)
