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

## Backend API (Start)

- `GET /health`
- `POST /api/v1/auth/login`
- `GET /api/v1/auth/me`
- `POST /api/v1/documents/index` (auth required)
- `GET /api/v1/documents/search` (auth required, `q`, mehrfach `tag`)

## Persistenz

- `domain.Document` wird ueber MongoDB gespeichert.
- Die Verbindung wird aus `backend/configs/service.yaml` unter `mongodb` gelesen.
- Wichtige Felder: `hosts`, `username`, `password`, `database`, `authDatabase`.
- Es gibt keinen InMemory-Fallback mehr: MongoDB ist fuer den Backend-Start erforderlich.

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
