# History

## 0.1.4 - 2026-05-10

- Blob-Speicherung: Kompression pro Eintrag mit zstd/gzip/none-Optionen hinzugefuegt.
  - Konfiguration ueber `repository.compressionType` in `service.yaml` (Standard: "none").
  - Format: 4-Byte Original-Laenge + 1-Byte Kompressionstyp + komprimierte Daten.
  - Universelle Lesbarkeit: Service auto-erkennt Kompressionstyp unabhaengig von Konfiguration.
  - Gemischte Kompressiontypen in einem .cnt-Container moeglich.
- Blob-Metadaten-Persistierung: .inf JSON-Dateien parallel zu .cnt-Containern.
  - Jeder .cnt-Container hat zugehorige N.inf-Datei mit Metadaten-Array.
  - Resilience bei Datenbankausfall: Container-Struktur kann aus .inf-Dateien rekonstruiert werden.
  - Neue Methoden: `ListAllContainerInfos()` fuer Iteration ueber alle Container und Eintraege.
- Thread-Safety: Pro-Container Mutex-Locks fuer sichere konkurrierende Zugriffe.
  - `muCurrent`: Schutz fuer Rotation-Logik (currentFile, currentNum, currentSize).
  - `containerLocks`: Pro-Container Locks fuer atomare Writes zu .cnt und .inf.
  - Ermöglicht echte Parallelität: Parallel writes zu verschiedenen Containern, Serialisierung bei gleicher Container.
  - Schutz gegen Datenverlust durch gleichzeitige Schreibvorgaenge.
- Tests: 8 neue Test-Methoden (6 fuer .inf-Persistierung + 2 fuer Concurrency).
  - Alle 18 Blob-Service-Tests erfolgreich (14 urspruengliche + 6 inf + 2 concurrent).
  - Validierung: .inf-Persistierung, Kompressions-Szenarien, parallele Loads, Serialisierung.

## 0.1.3 - 2026-05-10

- Frontend Upload-Dialog: Tag-Eingabe mit Chip-List und Autocomplete-Vorschlaege implementiert.
  - Enter erzeugt neue Tags, Auswahl uebernimmt vorhandene Tags.
  - Duplikat-Schutz auf Tag-Namen (case-insensitive).
- Frontend Suche: Tag-Eingabe von Freitext zu Multi-Select-Autocomplete umgestellt.
  - Nur vordefinierte Tags sind zulässig (forceSelection).
  - Auswahl als Chips sichtbar und entfernbar.
- Backend: `indexDocument`-Handler robust gegen Tags als Strings oder Objekte gemacht.
  - `parseTags()` Hilfsfunktion extrahiert Tag-Namen korrekt unabhaengig vom Format.
  - Deduplizierung und Normalisierung erfolgen serverseitig.

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