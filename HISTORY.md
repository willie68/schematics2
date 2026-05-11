# History

## 0.2.3 - 2026-05-11

- **Backend & Domain**: Effects Import mit Blob-Storage und Validierung.
  - Domain erweitert: `Effect.Image` (string) → `Effect.Images` ([]*ContainerInfo) für multiple Bilder
  - Neue Collection `effects` in MongoDB
  - Import-Tool: `cmd/import-effects/main.go` - einzelner Import für effects
  - Integriert in `cmd/import-all/main.go` - effects können selektiv importiert werden
  - Validierung: 
    - `effectType` wird gegen `effecttypes` Collection validiert
    - `manufacturer` wird gegen `manufacturers` Collection validiert oder neu erstellt
  - Bilder-Handling: `image` Dateien werden ins Blob-Repository gespeichert (wie Schematics)
  - foreignId wird ignoriert
  - 67 Effects erfolgreich importierbar

## 0.2.2 - 2026-05-11

- **Backend**: Effect Types Import mit automatischem Bild-Handling.
  - Neue Collection `effecttypes` in MongoDB für Effekt-Typen
  - Import-Tool: `cmd/import-effecttypes/main.go` - einzelner Import für effect types
  - Integriert in `cmd/import-all/main.go` - effect types können selektiv importiert werden
  - Bilder-Handling: `typeImage` Dateien werden automatisch nach `internal/repository/effecttypes` kopiert
  - Go Embed: `internal/repository/effecttypes/embedd.go` mit `//go:embed` für Zugriff auf eingebettete Bilder
  - Funktion `GetImage(filename)`, `ListImages()`, `GetImages()` für Bild-Zugriff zur Laufzeit
  - Mapping: `nls` aus Backup wird zu `i18n` in Go/MongoDB, `foreignId` wird ignoriert
  - 20 Effect Types erfolgreich importierbar

## 0.2.1 - 2026-05-11

- **Backend**: Vereinheitlichte Import-Tools mit neuem `cmd/import-all` Command.
  - Consolidation: 3 separate import Commands (`import-tags`, `import-manufacturers`, `import-schematics`) in einem Command kombiniert
  - Flexibilität: Optionale Flags zum selektiven Importieren (nur bestimmte Datentypen)
  - Convenience: Base-Ordner mit Unterverzeichnis-Struktur für alle Daten:
    - `base-dir/manufacturers/` - Hersteller JSON Dateien
    - `base-dir/tags/` - Tags JSON Dateien
    - `base-dir/schematics/` - Schematic-Verzeichnisse
  - Flags: `-base-dir`, `-manufacturers`, `-tags`, `-schematics`, `-dry-run`, `-skip-existing`, `-max-errors`
  - Legacy: Alte Commands bleiben für Backward-Kompatibilität erhalten

## 0.2.0 - 2026-05-10

- **Frontend**: Professioneller Image Viewer mit Zoom, Pan, Rotate und Download
  - Implementierung: PrimeVue native `Image` Komponente mit Preview-Overlay
  - Features: 
    - Zoom in/out mit Scrollrad oder Preview-Controls
    - 90° Rotation left/right ohne manuelles Sync-Problem
    - Pan (Drag) funktioniert nach Rotation korrekt
    - Download-Button zum Speichern von Dateien
  - Images skalieren proportional (`object-fit: contain`) für vollständige Anzeige
  - Cleanup: Entfernte externe Libs (panzoom, viewerjs) → nur noch PrimeVue
- **SearchView.vue**: Vereinfachte Button-Toolbars (nur Download-Button, sonst direkt auf Bild klicken)

## 0.1.19 - 2026-05-10

- Backend & Frontend: User Management System mit Email-Registrierung und Flood Protection.
  - **Domain**: User struct mit Email, Password (hashed), FirstName, LastName, Address (Street, ZipCode, City).
  - **MongoDB**: Neue `users` Collection mit eindeutigem Email-Index für Duplikat-Vermeidung.
  - **Backend Services**: 
    - `users.Service` mit Rate Limiting (10 Sekunden Mindestdauer pro Request, sequenziell).
    - `Register()` Method: Validiert alle Felder, hasht Passwort, speichert mit Flood-Protection.
    - `Authenticate()` Method: Validiert Email/Passwort gegen gespeicherte Benutzer.
  - **Backend API**:
    - `POST /api/v1/auth/register` (öffentlich): Registriert neuen Benutzer mit 10s Mindestdauer.
    - `POST /api/v1/auth/login` (erweitert): Unterstützt sowohl Admin als auch Email-Login.
  - **Frontend**:
    - `RegisterView.vue`: Vollständiges Registrierungsformular mit Adressangaben (Street, ZipCode, City).
    - `LoginView.vue`: Tabs für Admin-Login und Email-Login.
  - **Tests**: 15 Unit Tests für User Service (erfolg, validierung, duplikate, authentifizierung).
  - **Dependencies**: `github.com/google/uuid` für UUID-Generierung.

## 0.1.18 - 2026-05-10

- Frontend & Backend: Private Documents Filter korrekt implementiert.
  - Domain: SearchFilter um `IsAuthenticated` bool erweitert für Authentifizierungs-Status.
  - Backend: `/documents/search` prüft Authentifizierung und wendet Filter an:
    - **Gäste**: Nur `privateFile == false` Dokumente sichtbar (vollständig blockiert)
    - **Angemeldete ohne Filter**: Alle Dokumente sichtbar (keine Einschränkung)
    - **Angemeldete mit Private-Filter Button aktiv**: Nur `privateFile == true` Dokumente
  - Frontend: Private Button wird für Gäste `:disabled` gesetzt (keine Änderung möglich)
  - Frontend: search() setzt `privateOnly = false` für Gäste vor API-Call (doppelte Sicherheit)
  - Sicherheit: Mehrschichtig - Handler-Logik + MongoDB Filter + Frontend UI
  - Tests: Alle aktualisiert und bestanden (isAuthenticated Parameter hinzugefügt)
  - Bugfix: `isAuthenticated()` mit sicherer Typ-Assertion `ok` pattern

## 0.1.17 - 2026-05-10

- Frontend & Backend: Private Documents Filter für Gäste deaktiviert.
  - Frontend: Lock-Icon Button ist für Gäste `:disabled` (nicht sichtbar/klickbar).
  - Frontend: `search()` setzt `privateOnly = false` automatisch für Gäste vor API-Call.
  - Backend: `/documents/search` Endpoint ist jetzt öffentlich (nicht protected).
  - Backend: `isAuthenticated()` Hilfsfunktion prüft ob Request authentifiziert ist.
  - Backend: Handler erzwingt `privateOnly = false` für Gäste (kein Auth-Context).
  - Result: Gäste können nur öffentliche Dokumente suchen, nicht private.
  - Sicherheit: Mehrschichtig - Frontend GUI + Backend API Validation.

## 0.1.16 - 2026-05-10

- Frontend: Vollständiger Paginierungs-Stack mit Skip/Limit.
  - Limit ComboBox: [10, 20, 50, 100] Dokumente pro Seite.
  - Prev/Next Buttons für Seiten-Navigation.
  - Paginierungs-Counter: "X–Y of Z Dokumente".
  - API-Integration: ?skip=X&limit=20 Query-Parameter.

- Frontend & Backend: Tabellen-Redesign mit erweiterten Spalten.
  - Entfernt: ID-Spalte.
  - Neu: Subtitle, Tags (Badge-Chips mit grauem Hintergrund), PrivateFile (Lock-Icon), Owner.
  - Manufacturer, Model, Subtitle, Owner: Sortierbar.
  - Tags, PrivateFile: Nicht sortierbar.

- Frontend: Layout-Optimierungen.
  - Suchfelder: Gleichbreite 2-Spalten Grid (1fr 1fr auto auto auto).
  - Private Filter Button: Lock-Icon mit Warning-Severity wenn aktiv.
  - Search/Upload Buttons: Inline mit Icon-only + Tooltips.
  - Limit-Auswahl: Separate rechts-ausgerichtete Reihe.

- Frontend: Tag-Autocomplete Enter-Key Bug behoben.
  - Entfernt: handleTagKeydown() Konflikt mit PrimeVue native Enter-Handling.
  - Result: PrimeVue selektiert Tag korrekt bei Enter-Taste.

- Backend & Frontend: Server-Side Sorting implementiert.
  - Domain: SearchFilter um SortField/SortOrder erweitert.
  - MongoDB: Dynamische Sort-Feldwahl und Richtung via options.Find().SetSort().
  - InMemoryIndex: Feld-basierte In-Memory Sortierung (manufacturer/model/subtitle/owner).
  - Frontend: DataTable @sort Event triggert Backend mit sortField/sortOrder Parametern.
  - Tests: mongo_test.go, index.go, handler.go, mocks_test.go alle aktualisiert und passing.

- Backend & Frontend: Private Documents Filter implementiert.
  - Frontend: Neuer Lock-Icon Button vor dem Search-Button.
  - Frontend: Nur für angemeldete Benutzer sichtbar.
  - Frontend: Toggle-State mit visueller Rückmeldung (warning severity when active).
  - Backend: SearchFilter um PrivateOnly bool erweitert.
  - Backend: MongoStore Filter kombiniert privateFile==true mit anderen Filtern via $and.
  - Backend: InMemoryIndex filtert privateFile bei jedem Matching-Szenario.
  - API: Handler parst ?privateOnly=true Query-Parameter.
  - Tests: Alle Mocks und Handler-Aufrufe mit neuer Signatur aktualisiert.

- Build Status: go build ./... erfolgreich, go test ./internal/... all passing.

## 0.1.15 - 2026-05-10

- Frontend: Versionsnummer im Header angezeigt und klickbar.
  - Neue Datei: `frontend/src/config.js` mit APP_VERSION Konstante.
  - App-Header zeigt "Version 0.1.15" unter dem Untertitel an.
  - Version ist klickbar und öffnet Info-Dialog (cursor:pointer).
  - Neue Komponente: `frontend/src/components/InfoDialog.vue`
  
- Backend: Neuer `/api/v1/info` Endpoint mit Versionsinformation.
  - Neue Konstante: `BackendVersion = "0.1.15"` in handler.go
  - Endpoint gibt JSON mit Version und Status zurück: `{"version":"0.1.15", "status":"ok"}`
  - Info-Dialog zeigt Backend-Version und Live-Status an.
  - Version manuell mit HISTORY.md synchronisieren.

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