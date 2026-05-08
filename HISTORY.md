# History

## 0.1.2 - 2026-05-08

- MongoDB-basierter Document-Store fuer `domain.Document` implementiert.
- Mongo-Verbindungsaufbau an MCSPhotoIndex angelehnt (URI mit `hosts`, Auth optional ueber `authDatabase`).
- Index-Erstellung fuer Dokumente hinzugefuegt (`path` unique, `tags`, Textindex auf `title` und `text`).
- DI und API-Handler auf Store-Interface umgestellt.
- Mongo-Konfiguration in `service.yaml` auf `hosts`/`authDatabase` erweitert (Legacy-Felder bleiben kompatibel).
- InMemoryStore fuer Dokumente entfernt; kein Fallback mehr ohne MongoDB.

## 0.1.1 - 2026-05-08

- Build-Dokumentation unter `docs/BUILD.md` erstellt (lokaler Build, Skript-Build, Zertifikats-Generierung, Docker-Build).
- README um einen Verweis auf die zentrale Build-Dokumentation erweitert.
- Docker-Build-Flows dokumentiert inklusive Multi-Stage Build fuer Frontend und Backend.

## 0.1.0 - 2026-04-25

- Initiale GitHub-Verknüpfung mit dem Repository `willie68/schematic` hergestellt.
- Monorepo-Struktur mit Backend, Frontend und Projektdokumentation eingebracht.
- Git-Workflow mit `main` und issuebasierten Branches dokumentiert.