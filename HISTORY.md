# History

## 0.1.14 - 2026-05-10

- Backend: Manufacturer-Suche case-insensitiv gemacht.
  - `SuggestManufacturers()` findet jetzt "fend" -> ["Fender", "Fendt", "Airfend"]
  - MongoDB: `$regex` mit `$options: "i"` für case-insensitive Matching
  - Hersteller werden weiterhin mit Original-Schreibweise gespeichert (case-preserved)
  - Frontend erhält die korrekte Groß-/Kleinschreibung in Suggestions

## 0.1.13 - 2026-05-10

- Backend: Manufacturers Import Tool erstellt.
  - Neues Tool: `cmd/import-manufacturers` zum Importieren von Hersteller-JSON-Dateien.
  - Liest 390+ Hersteller aus `testdata/manufacturers` und importiert sie in MongoDB.
  - **Case-Preserved**: Hersteller behalten Original-Schreibweise (anders als Tags).
  - Alle existierenden Hersteller werden vor Import gelöscht.
  - MongoDB-Struktur: `{_id: "Samsung"}` (ohne separate Name-Feld).

## 0.1.12 - 2026-05-10

- Backend & Frontend: Manufacturers Collection System implementiert.
  - MongoDB: Neue `manufacturers` Collection für persistente Hersteller-Verwaltung.
  - Backend: `SuggestManufacturers()` Methode für Prefix-basierte Autocomplete.
  - Backend: `updateManufacturer()` bei jedem Document-Save auto-hinzufügen.
  - Backend: GET `/api/v1/manufacturers/suggest` Endpoint implementiert.
  - Frontend: AutoComplete für Hersteller-Feld in UploadDialog.
  - Frontend: `onManufacturerSuggest()` für API-Aufrufe.
  - **Wichtig:** Hersteller sind Case-Sensitive (anders als Tags die normalisiert sind).
  - Frontend erfolgreich kompiliert und deployed in backend/internal/webclient.

## 0.1.11 - 2026-05-10

- Backend: Tag-Import Tool erstellt.
  - Neues Tool: `cmd/import-tags` zum Importieren von Tag-JSON-Dateien.
  - Liest 169 Tags aus `testdata/tags` und importiert sie in MongoDB.
  - Tags werden normalisiert (Kleinbuchstaben, Whitespace-Trimming).
  - Alle existierenden Tags werden vor Import gelöscht.

## 0.1.10 - 2026-05-10

- Frontend: Toast-Benachrichtigungssystem implementiert.
  - Neues `useToast` Komposable für Toast-Verwaltung.
  - Toast-Komponente mit Stapelung von rechts unten nach oben.
  - Automatisches Verschwinden nach 5 Sekunden.
  - Toast-Typen: success, error, info, warning.
  - Toast in App.vue integriert.
  - Bei Suche: Toast mit Anzahl gefundener Dokumente anzeigen.
  - Frontend erfolgreich kompiliert.

## 0.1.9 - 2026-05-10

- Backend: InMemoryIndex entfernt (veraltet, MongoIndex ist Standard).
  - Gelöschte Dateien: `index.go`, `index_test.go`
  - Nur MongoIndex bleibt bestehen.
  - 4 MongoIndex Tests bestehen, Backend erfolgreich kompiliert.

## 0.1.8 - 2026-05-10

- Backend: Suchlogik zu MongoDB delegiert.
  - Entfernt alle in-Memory Filterung in MongoIndex.
  - Neue `SearchFilter` Domain-Klasse für standardisierte Such-Parameter.
  - MongoDocumentStore.Search() implementiert echte MongoDB-Queries (bson.D).
  - Tag-Filter: Normalisierung + `$all` Operator für UND-Logik.
  - Text-Filter: Unterstützung für MongoDB Text-Index.
  - MongoIndex.Search() nur noch Normalisierung + Filter-Konstruktion.
  - Score entfernt (nicht nötig, Sortierung nach _id).
  - Alle 7 Tests bestehen, Backend erfolgreich kompiliert.

## 0.1.7 - 2026-05-10

- Backend: MongoDB-basierter Index implementiert.
  - Ersetzt InMemoryIndex mit MongoIndex für persistente Suche.
  - Hybrid-Ansatz: Dokumente aus DB laden, in-Memory filtern, Ergebnisse bewerten.
  - Unterstützt Nur-Query-, Nur-Tag- und kombinierte Such-Szenarien.
  - 7 Tests bestehen.
  - DI-Konfiguration aktualisiert: `NewMongoIndex` als aktiver Index.
  - Vollständig kompiliert und getestet.

## 0.1.6 - 2026-05-10

- Backend: Tag-basierte Suche implementiert (ohne Volltext-Query).
  - Index.Search() behandelt Query-leere Anfragen korrekt.
  - Nur-Tag-Suche: Gibt alle Dokumente mit den angeforderten Tags zurück.
  - Kombinierte Suche: Volltext-Query Ergebnisse nach Tags gefiltert (UND-Logik).
  - Neue Tests fuer alle Such-Szenarien (3 Tests bestehen).

## 0.1.5 - 2026-05-10

- Frontend SearchView: Enter-Taste triggert Suche.
  - Suchfeld: Enter führt Suche direkt aus.
  - Tags-Feld: Erstes Enter schließt Tag ab (wenn Suggestion offen), zweites Enter triggert Suche.
  - Ermöglicht nahtlose Tastatur-Navigation ohne Button-Klick.

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