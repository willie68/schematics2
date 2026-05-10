# Schematic2

Schematic2 ist der Nachfolger von WilliesSchematicsWorld als Monorepo.

Das Repository ist mit GitHub unter https://github.com/willie68/schematic verknüpft.

## Ziele

- Web-Frontend zur Suche in Schaltplänen, Dokumentationen und weiteren Dateien (PNG, JPG, PDF, ...)
- Vorindizierung über Tags
- Volltextsuche über indizierte Inhalte
- Eigener Authentifizierungs- und Autorisierungsdienst

## Monorepo-Struktur

- `backend/`: Go REST API (go-micro-orientierter Aufbau mit `internal/` und DI über `do`)
- `frontend/`: Vue + PrimeVue Web-Frontend
- `docs/`: Projekt- und Architekturdokumentation

## Build-Dokumentation

- Build-Anleitung: `docs/BUILD.md`

## Backend-Tools

### Tag-Import
Tags aus JSON-Dateien in MongoDB importieren:

```bash
cd backend
go run cmd/import-tags/main.go [-tags-dir testdata/tags]
```

Siehe `backend/cmd/import-tags/README.md` für Details.

### Manufacturers-Import
Hersteller aus JSON-Dateien in MongoDB importieren:

```bash
cd backend
go run cmd/import-manufacturers/main.go [-manufacturers-dir testdata/manufacturers]
```

Siehe `backend/cmd/import-manufacturers/README.md` für Details.

## Backend API (Start)

- `GET /health`
- `POST /api/v1/auth/login`
- `GET /api/v1/auth/me`
- `GET /api/v1/tags` - Liste alle Tags auf
- `GET /api/v1/tags/suggest?q=<prefix>&limit=<n>` - Schlag Tags vor (Prefix-Match, normalisiert)
- `GET /api/v1/manufacturers/suggest?q=<prefix>&limit=<n>` - Schlag Hersteller vor (Prefix-Match, case-sensitive)
- `POST /api/v1/documents/index` (auth required) - Indexiere ein Dokument mit Tags
- `GET /api/v1/documents/search?q=<query>&tag=<t1>&tag=<t2>` (auth required) - Suche mit Volltext und Tags

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

### 1) Backend

```bash
cd backend
go mod tidy
go run cmd/api/main.go
```

Umgebungsvariablen siehe `backend/configs/service.env.example`.

### 2) Frontend

```bash
cd frontend
npm install
npm run dev
```

Frontend läuft standardmäßig auf `http://localhost:5173`.

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
