# Schematic2

Schematic2 ist der Nachfolger von WilliesSchematicsWorld als Monorepo.

Das Repository ist mit GitHub unter https://github.com/willie68/schematic verknüpft.

**Version: Backend 0.2.22, Frontend 0.2.7**

## Features

- **Dokumentsuche**: Web-Frontend zur Suche in Schaltplänen, Dokumentationen und weiteren Dateien (PNG, JPG, PDF, ...)
- **Tag-System**: Vorindizierung über Tags und Volltextsuche über indizierte Inhalte
- **Image Viewer**: Native Zoom, Pan, Rotate, und Download für Bilddateien
- **Authentifizierung**: Eigener Authentifizierungs- und Autorisierungsdienst mit User Registration
- **Private Documents**: Unterstützung für private und öffentliche Dokumente
- **Effektdatenbank**: Verwaltung und Suche von Effekten mit Sortierung
  - i18n German Übersetzungen für Effekt-Typen
  - Sortierung nach Typ, Hersteller, Modell, Spannung, Strom
  - Bild-Upload und Anschluss-Information

## Ziele

- Dezentrales Dokumentenmanagementsystem für Schaltpläne und technische Dokumente
- Benutzerfreundliche Suche und Kategorisierung
- Sichere Verwaltung privater Dokumente

## Monorepo-Struktur

- `backend/`: Go REST API (go-micro-orientierter Aufbau mit `internal/` und DI über `do`)
- `frontend/`: Vue 3.5 + PrimeVue 3.53 Web-Frontend
- `docs/`: Projekt- und Architekturdokumentation

## Deployment

### Docker-Build

```bash
# Standard: Reverse-Proxy unter /schematics2 (mit Apache/Nginx)
.\scripts\buildDocker.cmd

# Oder: Direkter Client-Zugriff ohne Base-Path
.\scripts\buildDocker.cmd /client
```

**Automatische Injection:**
- BUILD_TIME: ISO 8601 Zeitstempel
- VCS_REF: Git Short Hash
- BASE_PATH: Frontend Base-Path + Backend Redirect-Pfad

Versionsnummern werden automatisch aus Quelltext injiziert (Backend: `internal/version/version.go`, Frontend: `package.json`).

### Reverse-Proxy (Apache)

Beispiel Apache VirtualHost Konfiguration für `/schematics2`:

```apache
ProxyPreserveHost On
SSLProxyEngine On
SSLProxyVerify none

# Redirect nackter Pfad auf mit Slash
RedirectMatch ^/schematics2$ /schematics2/

# Trailing Slash ERFORDERLICH auf beiden Seiten!
ProxyPass /schematics2/ https://192.168.178.14:9743/
ProxyPassReverse /schematics2/ https://192.168.178.14:9743/
```

⚠️ **Wichtig**: Trailing Slash (`/`) auf beiden Seiten des ProxyPass – ohne ihn entstehen `//` Pfade und das Frontend erhält falsche API URLs.

## Build-Dokumentation

- Build-Anleitung: `docs/BUILD.md`

## Backend-Tools

### Unified Import (Tags, Manufacturers, Schematics, Effect Types, Effects)

Alle Daten in einem Command importieren. Der Base-Ordner muss folgende Struktur haben:

```
base-dir/
├── manufacturers/       # JSON Dateien mit Herstellern
├── tags/               # JSON Dateien mit Tags
├── schematics/         # Verzeichnisse mit schematic.json und Dateien
├── effecttypes/        # Verzeichnisse mit effecttype.json und Bildern
└── effects/            # Verzeichnisse mit effect.json und Bildern
```

**Standard-Import (alle Daten):**
```bash
cd backend
go run cmd/import-all/main.go -base-dir ./testdata
```

**Nur spezifische Daten importieren:**
```bash
# Nur Effects
go run cmd/import-all/main.go -base-dir ./testdata -manufacturers=false -tags=false -schematics=false -effecttypes=false

# Nur Tags und Hersteller
go run cmd/import-all/main.go -base-dir ./testdata -schematics=false -effecttypes=false -effects=false
```

**Weitere Flags:**
- `-dry-run` - Nur validieren, keine Änderungen schreiben
- `-skip-existing` - Existierende Dokumente überspringen (default: true)
- `-max-errors` - Maximale Fehleranzahl vor Abbruch (default: 50, 0=unbegrenzt)

**Effect Types & Effects Import Details:**
- JSON Struktur: `effecttype.json` / `effect.json` in jedem Verzeichnis
- Effect Type Bilder: Werden nach `internal/repository/effecttypes` kopiert, Go Embed eingebunden
- Effect Bilder: Werden ins Blob-Repository gespeichert (mit ContainerInfo)
- Effect Validierung: 
  - `effectType` wird gegen die `effecttypes` Collection validiert
  - `manufacturer` wird gegen die `manufacturers` Collection validiert (oder neu erstellt)
- Domain: `Effect.Image` → `Effect.Images` (Array von ContainerInfo für Blob-Storage)

### Legacy: Einzelne Import-Commands

Die folgenden separaten Commands sind deprecated, aber noch vorhanden:
- `cmd/import-tags/main.go` - Nur Tags importieren
- `cmd/import-manufacturers/main.go` - Nur Hersteller importieren
- `cmd/import-schematics/main.go` - Nur Schematics importieren
- `cmd/import-effecttypes/main.go` - Nur Effect Types importieren
- `cmd/import-effects/main.go` - Nur Effects importieren

Für neue Importe bitte `cmd/import-all/main.go` verwenden!

## Backend API (Start)

- `GET /health`
- `GET /api/v1/info` - Backend Version und Status
- `POST /api/v1/auth/login` - Admin oder Email-Login mit Benutzername/Email und Passwort
- `POST /api/v1/auth/register` - Registriere neuen Benutzer (öffentlich, Rate-Limited: 10s Mindestdauer)
- `GET /api/v1/auth/me` - Aktuell authentifizierter Benutzer (JWT erforderlich)
- `GET /api/v1/tags` - Liste alle Tags auf
- `GET /api/v1/tags/suggest?q=<prefix>&limit=<n>` - Schlag Tags vor (Prefix-Match, case-insensitiv, normalisiert)
- `GET /api/v1/manufacturers/suggest?q=<prefix>&limit=<n>` - Schlag Hersteller vor (Prefix-Match, case-insensitiv, case-preserved)
- `POST /api/v1/documents/index` (auth required) - Indexiere ein Dokument mit Tags
- `GET /api/v1/documents/search?q=<query>&tag=<t1>&tag=<t2>` (auth required) - Suche mit Volltext und Tags

### Effects API

- `GET /api/v1/effects/search?q=<query>&skip=<n>&limit=<n>` - Durchsuche Effects mit Regex-Filter und Pagination
  - `q` - Suchtext (durchsucht effectType, manufacturer, model, tags, comment)
  - `skip` - Überspringe n Ergebnisse (Pagination)
  - `limit` - Begrenzen Sie auf n Ergebnisse (10, 20, 50, ...)
  - Response: `{ results: [...], total: n }`
- `GET /api/v1/effects/<id>/image` - Hole das erste Bild eines Effects (als Blob)
- `GET /api/v1/effecttypes` - Liste alle Effekt-Typen mit i18n Namen
- `GET /api/v1/connectors/<name>` - Hole Connector-Bild (PNG/JPG aus embedded filesystem)
- `POST /api/v1/effects` (auth required) - Erstelle neuen Effect mit multipart/form-data
  - Form-Felder: `effectType`, `manufacturer`, `model`, `voltage`, `current`, `connector`, `image`
  - Response: `{ id, ... }`

### Authentifizierung

- **Admin-Login**: Username `admin` mit Passwort (aus Env-Var `ADMIN_PASS`)
- **Email-Login**: Email und Passwort für selbstregistrierte Benutzer
- **JWT Token**: Wird in `Authorization: Bearer <token>` Header gesendet
- **Token-Dauer**: 24 Stunden

### Benutzerregistrierung

```bash
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securePassword123",
  "firstName": "John",
  "lastName": "Doe",
  "street": "123 Main St",
  "zipCode": "12345",
  "city": "TestCity"
}
```

**Anforderungen:**
- Email (eindeutig in der Datenbank)
- Passwort (mindestens 8 Zeichen)
- Alle Adressfelder erforderlich
- **Flood-Protection**: Jede Registrierung dauert mindestens 10 Sekunden, nur ein Request gleichzeitig


## Persistenz

- `domain.Document` wird ueber MongoDB gespeichert.
- Der Such-Index ist MongoDB-basiert (MongoIndex) für persistente, skalierbare Suche.
- Die Verbindung wird aus `backend/configs/service.yaml` unter `mongodb` gelesen.
- Wichtige Felder: `hosts`, `username`, `password`, `database`, `authDatabase`, `directConnection`.
- Es gibt keinen InMemory-Fallback mehr: MongoDB ist fuer den Backend-Start erforderlich.
- Tags werden normalisiert (Kleinbuchstaben, Whitespace-Trimming, Duplikat-Entfernung) vor der Speicherung.
- Tag-Counter werden in einer separaten Collection `tags` ueber Upsert aktualisiert.
- Suche unterstützt: Nur-Query, Nur-Tags, und kombinierte Anfragen mit UND-Logik.

## Blob-Speicherung (Dateien)

- Dateien werden in rotierenden Container-Dateien (`*.cnt`) im Repository-Verzeichnis gespeichert.
- Container-Format: `[4-byte original-length][1-byte compression-type][variable-length data]`
- Container rotieren, wenn die konfigurierte maximale Groesse erreicht wird.
- Komprimierung ist optional und pro Datei konfigurierbar:
  - `"none"` (Standard): Keine Komprimierung
  - `"gzip"`: Standard-Komprimierung (gut für Text)
  - `"zstd"` (empfohlen): Zstandard-Komprimierung (besser bei bereits komprimierten Formaten wie PDF, JPEG, TIF)
- Ein Container kann gemischte komprimierte und unkomprimierte Daten enthalten.
- Kompression ist in `backend/configs/service.yaml` unter `repository.compressionType` konfigurierbar.

## Schnellstart

### Voraussetzungen

- Go 1.26+
- Node.js 22+
- MongoDB 5.0+ (lokal oder via Docker)
- PowerShell (für buildDocker.cmd auf Windows)

### 1) MongoDB starten (Docker)

```bash
docker run -d --name mongodb -p 27017:27017 mongo:7
```

Oder für Entwicklung direkt lokal installieren.

### 2) Backend

```bash
cd backend
go mod tidy
go run cmd/server/main.go
```

Der Server läuft standardmäßig auf `https://localhost:8443`.

**Umgebungsvariablen** (optional, defaults sind vorhanden):
- `JWTSecret`: JWT Secret (default: aus configs/secrets.yaml)
- `AdminUser`: Admin Username (default: admin)
- `AdminPass`: Admin Password (default: admin123)

Siehe `backend/configs/service.yaml` für alle Konfigurationsoptionen.

### 3) Frontend

```bash
cd frontend
npm install
npm run dev
```

Frontend läuft standardmäßig auf `http://localhost:5173`.

#### Frontend-Architektur

- **Composition API mit `<script setup>`**: Moderne Vue 3 Syntax für bessere Performance
- **PrimeVue 3.53**: Umfangreiche UI-Komponenten (DataTable, Dialog, Dropdown, AutoComplete, etc.)
- **Axios Interceptors**: Zentrale Authentifizierung und 401-Error-Handling
- **Komponenten**:
  - `EffectsView.vue` - Hauptseite für Effects-Datenbank mit Suche, Pagination, Detail-Modal
  - `EffectUploadDialog.vue` - Wiederverwendbare Komponente zum Hochladen neuer Effects
    - Props: `visible` (v-model), `effectTypes` (Array)
    - Events: `@update:visible`, `@effect-created`
    - Features: Form-Validierung, Manufacturer-Autocomplete, Connector-Dropdown, Bild-Upload

## Git-Workflow

- Der stabile Integrationszweig ist `main`.
- Neue Arbeiten erfolgen grundsätzlich in Issue-Branches auf Basis von `main`.
- Branch-Namenskonvention: `issue/<nummer>-<kurzer-slug>`.
- Beispiel: `issue/42-search-index`.
- Ohne vorhandenes GitHub-Issue wird kein dauerhafter Arbeitsbranch angelegt.

## Installation und Nutzung

- Repository klonen.
- Backend-Abhängigkeiten mit `go mod tidy` im Verzeichnis `backend` auflösen.
- Frontend-Abhängigkeiten mit `npm install` im Verzeichnis `frontend` installieren.
- Backend und Frontend wie oben beschrieben starten.
- Änderungen über einen Issue-Branch entwickeln und anschließend in `main` integrieren.

## Nächste Schritte

- Persistente Speicherung für Dokumente/Index (z. B. MongoDB + Suchindex)
- Dateiupload-Pipeline für PDF/OCR/Bild-Metadaten
- Feingranulare Rollen- und Ressourcenrechte
- Swagger/OpenAPI für alle Endpunkte
