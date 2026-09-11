# History

## 0.3.10 - 2026-09-11 (Backend & Frontend)

- **Frontend**: App-Logo und Favicon (SVG/PNG) im Header und in `index.html`.
- **Frontend**: Mobile Suchzeile — Search-Feld breiter, Lücke vor dem Suchbutton geschlossen.
- **Frontend**: Hilfe in der mobilen Suche startet zugeklappt.
- **HISTORY.md**: Merge-Konflikt in 0.3.9 bereinigt.
- **Backend**: API-Mocks (`mocks_test.go`) aktualisiert.
- **Tooling**: Cursor Release-Skill unter `.cursor/skills/release/`.
- **Frontend/Backend**: Versionsstände auf `0.3.10` angehoben.

## 0.3.9 - 2026-08-23 (Backend & Frontend)

- **GELF via TCP**: `logging.go` unterstützt GELF-Versand via TCP (zusätzlich zu UDP).
  - Neue Config-Option `logging.gelf-protocol: tcp` (Standard: `udp`).
- **Backend**: Nil-Pointer-Check in `GetHash()` ergänzt, Absturz bei fehlendem Container verhindert.
- **Frontend/Backend**: Versionsstände auf `0.3.9` angehoben.

## 0.3.8 - 2026-07-01 (Backend & Frontend)

- **NS_ERROR_NET_RESET Fehler bei Upload/Presence/PATCH behoben**:
  - **Backend**: Hash-basierter Presence Check implementiert (nur Hash statt komplette Base64-Datei)
  - **Backend**: HTTP Server `MaxHeaderBytes` auf 50MB erhöht für große JSON-Payloads
  - **Backend**: Body Size Limit Middleware (200MB) nur auf `/documents` Endpoints angewendet
  - **Backend**: JSON Decoder optimiert (direktes Dekodieren statt Umweg über `map[string]any`)
  - **Backend**: Content-Length Header in Response gesetzt für korrekte Verbindung
  - **Frontend**: Axios mit `keepAlive: true` konfiguriert
  - **Frontend**: Presence Check mit SHA-256 Hash statt ganzer Datei
  - **Frontend**: PATCH Payload optimiert (nur Type-Updates für bestehende Files, keine Base64-Daten)
  - **Frontend**: Upload- und Edit-Dialog Timeouts erhöht (300s)
  - **Frontend**: Besseres Error-Handling für Netzwerkfehler (Timeout, Netzwerk, 413, 500)

## 0.3.7 - 2026-06-29 (Backend & Frontend)

- **Issue #1 - Deployment**:
  - **Backend**: Backup-Pfad in `newBackup` auf `cfg.Backup.BackupPath` korrigiert
  - **Backend**: Download-Blob-Ladevorgang mit `measurement` instrumentiert (`download-blob`)
  - **Backend/Frontend**: Versionsstände auf `0.3.7` angehoben
  - **Dokumentation**: README und HISTORY für das Deployment synchronisiert

## 0.3.4 - 2026-06-25 (Backend & Frontend)

- **Mobile Suche und Detailansicht deutlich verbessert**:
  - Header in der mobilen Suche/Effektdatenbank ausgeblendet, in der Einstiegsseite weiterhin sichtbar
  - In Such- und Effekttitel mobile Zurück-Navigation per Pfeil zur Startseite ergänzt
  - Hilfetext in der Suche auf Mobilgeräten als kompakter, einklappbarer Bereich umgesetzt
  - Splitter zwischen Trefferliste und Details auf Mobilgeräten verschlankt; Abstände reduziert
  - Bei Auswahl eines Treffers wird die Detailansicht mobil automatisch in voller Breite geöffnet
  - Dateitypen in der mobilen Dateitabelle als Icons statt Textlabels dargestellt

- **Mobile Dateiansicht überarbeitet**:
  - Dateiklick öffnet auf Mobilgeräten einen Vollbild-Viewer
  - Unterstützung für alle bisher unterstützten Dateitypen im mobilen Vollbildmodus (inkl. PDF/Bild-Vorschau und Fallback-Ansicht)
  - Datei-Selektion reagiert zuverlässig bei jedem Tap (Event-Handling angepasst)
  - Bildanzeige im mobilen Vollbild-Viewer ohne PrimeVue-Preview-Overlay (kein graues Auge-Popup)

- **Layout- und Usability-Fixes**:
  - Scroll- und Höhenlogik in der Detailansicht überarbeitet, damit der Dateibereich wieder erreichbar ist
  - Mobile Toasts kompakter und minimalistischer gestaltet, mit robuster Darstellung ohne Abschneiden

## 0.3.5 - 2026-06-28 (Backend & Frontend)

- **Hash-basierter Presence-Check für Datei-Uploads**:
  - **Backend**: `checkFilePresence` implementiert (`POST /api/v1/files/presence`)
  - **Backend**: Payload-Bytes werden per `GetHashFromPayload` gehasht und über `HasHash` gegen vorhandene Dateien geprüft
  - **Backend**: API-Tests für Presence-Check ergänzt (Treffer, kein Treffer, invalid payload)
- **Frontend Upload/Edit verbessert**:
  - **Frontend**: Presence-Check beim direkten Hinzufügen jeder Datei in Upload- und Edit-Dialog
  - **Frontend**: Visuelles Feedback je Datei:
    - Orangefarbenes Ausrufezeichen (Tooltip: "Dokument bereits hochgeladen") bei `presence=true`
    - Grüner Haken bei `presence=false`
- **Share-Link UX angepasst**:
  - **Frontend**: Share-Link-Erzeugung im Such-View nicht mehr an Login-Status gekoppelt
- **Hash-Service Testabdeckung erweitert**:
  - **Backend**: Umfangreiche Unit-Tests für Hash-Service und Rebuild-Logik ergänzt

## 0.3.4 - 2026-06-27 (Backend & Frontend)

- **Feature: Share Document (Issue #4) vollständig umgesetzt**:
  - **Backend**: Share-Erzeugung abgesichert (nur Owner/Admin, JWT erforderlich)
  - **Backend**: Neue öffentliche Share-Endpunkte:
    - `GET /api/v1/shares/{link}` für Dokumentzugriff über Share-Link
    - `GET /api/v1/shares/{link}/files/{filename}` für Dateidownload über Share-Link
  - **Backend**: Ablaufprüfung für Shares (`410 Gone` bei abgelaufenem Share)
  - **Backend**: Share-Persistenz in MongoDB robust gemacht (Mongo `_id` als ObjectID, Ausgabe als Hex-Share-ID)
  - **Frontend**: Dokumentansicht unterstützt `share`-Query-Parameter
  - **Frontend**: Dateiabrufe verwenden bei aktivem Share die neuen Share-Endpunkte
  - **Frontend**: "Link kopieren" erzeugt jetzt einen echten Share-Link über die API
  - **Frontend**: Generierte Share-Links zeigen auf `/client/search?share=...` statt auf Dokument-IDs
  - **Frontend**: Ein-Datei-Shared-Dokumente bleiben beim Klick auf die Dateizeile stabil geöffnet

## 0.3.3 - 2026-06-21 (Backend)

- **Backup-Workflow erweitert und robuster gemacht**:
  - Backup startet nach Service-Start mit initialer Verzögerung und läuft danach im konfigurierten Intervall
  - MongoDB-Collections werden speicherschonend per Streaming in JSON exportiert
  - `effects.json` und `documents.json` werden gestreamt verarbeitet, um referenzierte Blob-Dateien in den Backup-Baum zu exportieren
  - Dokumentdateien werden strukturiert unter `db/documents/<manufacturer>/<model>/` abgelegt und mit indexbasierten Dateinamen versehen
  - Pfad-/Dateinamen-Bereinigung für Windows (u.a. problematische Endungen und reservierte Namen)
  - Erfolgreiche Backups werden in mehrteilige ZIP-Archive mit maximal 500 MB pro Datei geschrieben
  - Nach erfolgreicher Archivierung wird der temporäre Backup-Ordner entfernt und alte Backup-ZIPs werden bereinigt

## 0.3.2 - 2026-06-16 (Backend)

- **Case-insensitive Login und Email-Normalisierung**:
  - **Backend**: `GetUserByEmail()` mit MongoDB Collation für case-insensitive Suche
  - **Backend**: Email-Index mit case-insensitiver Collation in `ensureIndexes()`
  - **Service**: Email in `Register()`, `Authenticate()`, `UserStatus()` normalisiert (toLowerCase)
  - **API-Handler**: Email in `me()` und `changePassword()` normalisiert
  - **Index-Migration**: Alter Email-Index wird vor Erstellung des neuen gelöscht (verhindert Konflikt)
  - **Ergebnis**: Benutzer können sich mit beliebiger Groß-/Kleinschreibung anmelden (User@Example.com, user@example.com, USER@EXAMPLE.COM)

## 0.3.1 - 2026-06-16 (Backend & Frontend)

- **Verbessertes Benutzerverwaltungs-Interface**:
  - **Frontend**: Button-Label und Icon ändern sich dynamisch basierend auf User-Status
    - "Deaktivieren" (rot, Icon: pi-ban) für aktive Benutzer
    - "Aktivieren" (grün, Icon: pi-check) für deaktivierte Benutzer
    - Severity wechselt: danger (rot) ↔ success (grün)
  - **Backend**: `userStore` Interface aktualisiert um `ListAllUsers()` und `GetUserByID()` Methoden
  - **Testing**: Mocks mit mockery v3.7.1 regeneriert
  - **Konfiguration**: `.mockery.yml` für mockery v3 konfiguriert

## 0.3.0 - 2026-06-16 (Backend & Frontend)

- **Benutzerverwaltungs-Interface für Admins**:
  - **Backend**: Neue Endpoints `/api/v1/users` (GET) zur Auflistung aller Benutzer und `/api/v1/users/{id}/deactivate` (PATCH) zum Aktivieren/Deaktivieren
  - **Backend**: Beide Endpoints nur für Admin-Benutzer erreichbar (HTTP 403 für Nicht-Admins)
  - **Frontend**: Neuer Menüpunkt "Benutzerverwaltung" (ersetzt "neuer Benutzer")
  - **Dialog**: Tabelle mit allen Benutzern (E-Mail, Vorname, Nachname, Adresse, Status)
  - **Suchfunktion**: Filterbar nach E-Mail und Namen (oben links)
  - **Inline-Buttons**: "Deaktivieren" (aktiv nur bei selektierter Zeile) und "Neuer Benutzer" (oben rechts)
  - **Zeilenauswahl**: Single-Select, Deaktivieren-Button wird nur bei selektierten Zeilen aktiviert
  - **Status-Anzeige**: Benutzer können als "Aktiv" (grün) oder "Deaktiviert" (rot) angezeigt werden

- **Repository-Erweiterung für Benutzerverwaltung**:
  - `ListAllUsers(ctx context.Context) ([]User, error)`: Ruft alle Benutzer sortiert nach E-Mail ab
  - `GetUserByID(ctx context.Context, userID string) (*User, error)`: Ruft Benutzer nach ID ab

## 0.2.49 - 2026-06-16 (Backend & Frontend)

- **Admin-gesteuerte Benutzeranlage**:
  - **Backend**: `/api/v1/auth/register` Endpoint nun nur für Admin-Benutzer erreichbar (HTTP 403 für Nicht-Admins)
  - **Frontend**: Menüpunkt "neuer Benutzer" erscheint nur für angemeldete Admins
  - **Dialog**: Admin-Dialog zur manuellen Benutzeranlage mit Feldern: Vorname, Nachname, E-Mail, Passwort (vorbelegt), Adresse
  - **Passwort-Vorbelegung**: Default-Passwort "Start1234" wird automatisch eingetragen und ist sichtbar (für schnelle Admin-Anlage)

- **Erzwungenes Passwort-Ändern beim ersten Login**:
  - **Backend**: User-Model um `ChangePassword` Bool-Flag erweitert
  - **Register-Service**: Neue Benutzer erhalten automatisch `ChangePassword: true`
  - **Login-Response**: Neue Felder `changePassword` (bool) und `email` (string) in der Login-Antwort
  - **Frontend LoginView**: Bei `changePassword: true` wird sofort das Passwort-Ändern-Formular angezeigt
  - **Erzwungener Flow**: Benutzer kann nicht zur App navigieren, bevor Passwort geändert ist
  - **Flag-Reset**: Nach erfolgreichem Passwort-Wechsel wird `ChangePassword` auf `false` gesetzt

## 0.2.48 - 2026-06-02 (Backend & Frontend)

- **Dokumentation auf aktuellen Codestand synchronisiert**:
  - `README.md` auf Version **0.2.48** aktualisiert (Backend + Frontend)
  - API-Abschnitt auf aktuelle Endpunkte gebracht:
    - User-Endpunkte (`/api/v1/users/me`, `/api/v1/users/change-password`)
    - Dokument-Endpunkte inkl. Direktzugriff und Datei-Download
    - Effects-Endpunkte inkl. Sortierparametern und CRUD-Übersicht
  - Feature-Übersicht erweitert (Kontoverwaltung, Datenschutz/Rechtstexte, Cookie-Hinweis)
  - Roadmap/Nächste Schritte an den aktuellen Implementierungsstand angepasst

## 0.2.44 - 2026-05-21 (Backend & Frontend)

- **Dokumenten-Link-Sharing und Direktzugriff**:
  - **Backend**: Neuer GET `/api/v1/documents/{id}` Endpoint für Direktzugriff auf einzelne Dokumente
    - Authentifizierung und Autorisierung integriert (private Dokumente nur für Owner)
  - **Frontend**: Icon-Button "Link kopieren" in der Dokumentauswahl
    - Erzeugt einen direkten Link mit Dokument-ID als Query-Parameter
    - Beim Öffnen des Links wird das Dokument automatisch geladen und im Vollbildmodus angezeigt
  - **Frontend**: Verbessertes Vollbildmodus-Layout
    - Zeigt Detail-Panel (Metadaten + Datei-Tabelle) links und Viewer rechts
    - Ermöglicht Wechsel zwischen Dateien eines Dokuments im Vollbildmodus

- **Zentrale Dokumenttyp-Verwaltung**:
  - **Frontend**: Neue `constants/docTypes.js` mit zentraler DOC_TYPES-Definition
    - Verhindert Duplikation in EditDialog und UploadDialog
    - Einfaches Hinzufügen neuer Dokumenttypen an einer Stelle
    - Hilfsfunktionen: `getDocTypeByValue()`, `getDocTypeLabel()`
  - Alle Dialoge nutzen die zentrale Konfiguration

- **Dokumentbestand-Anzeige**:
  - **Backend**: `infoResponse` um `DocumentCount` Feld erweitert
  - **Backend**: `CountAll()` Methode in MongoStore implementiert
  - **Frontend**: Dokumentanzahl vom `/api/v1/info` Endpoint geholt und auf SearchView angezeigt
    - Info-Box mit Gesamtzahl aller Dokumente in der Datenbank
    - Wird beim Laden der Seite automatisch aktualisiert

## 0.2.41 - 2026-05-18 (Backend & Frontend)

- **Erweiterte Datei-Upload-Unterstützung für große Dateien**:
  - **Frontend**: TXT und ZIP zu erlaubten Dateiformaten hinzugefügt
    - UploadDialog & EditDialog unterstützen jetzt: PDF, JPG, PNG, GIF, TIF, BMP, ZIP, TXT
    - Validierung aktualisiert mit neuen Formaten in Fehlermeldungen
  - **Backend**: HTTP-Timeout erhöht für große Uploads
    - ReadTimeout & WriteTimeout: 15 Sekunden → 300 Sekunden (5 Minuten)
    - Ermöglicht stabile Uploads auf langsamen Verbindungen
  - **Backend**: MaxBodySize Middleware implementiert
    - 50MB Upload-Limit mit klarer Fehlermeldung
    - Schützt Server vor extrem großen Requests
  - **Ergebnis**: 5MB+ Dateien können jetzt ohne Verbindungsabbruch hochgeladen werden

## 0.2.39 - 2026-05-18 (Frontend)

- **ZIP-Datei-Upload**: Unterstützung für ZIP-Archiv-Uploads hinzugefügt
  - Frontend: ZIP in accept-Attribut und MIME-Type Mapping erweitert
  - Ermöglicht das Hochladen von komprimierten Dokumenten

## 0.2.38 - 2026-05-18 (Backend & Frontend)

- **Suchfunktion vollständig integriert**: Indizes in Applikation + benutzerfreundliche Dokumentation
  - **Backend**: MongoDB B-tree Indizes in `ensureIndexes()` integriert statt separates Init-Skript
    - Automatische Index-Erstellung beim App-Start (Prepare-Methode)
    - Indizes auf: manufacturer, model, subtitle, description, tags, privateFile, owner, compound (privateFile+owner)
  - **Frontend**: Erweiterte Such-Hilfetext mit Syntax-Erklärung in SearchView
    - Nutzer sehen direkt alle Operatoren: `+Begriff` (UND), `-Begriff` (NOT), `Begriff*` (Prefix)
    - Hinweis auf OR-Verknüpfung von Leerzeichen und UND-Verknüpfung von Tags
  - **Dokumentation**: `backend/SEARCH.md` enthält vollständige API-Spezifikation und Performance-Tipps
  - **Ergebnis**: Suchfunktion ist vollständig dokumentiert und produktionsreif

## 0.2.37 - 2026-05-18 (Backend)

- **Erweiterte Suchfunktion mit Prefix- und Operatoren-Unterstützung**: Flexible Query-Verarbeitung
  - Query-Parser: Unterstützt Prefix-Matching (`fend*`), AND-Operator (`+term`), OR-Operator (Leerzeichen), NOT-Operator (`-term`)
  - MongoDB Filter-Builder: Nutzt `$regex` statt `$text` für bessere Performance und Flexibilität
  - Neue Typen: `ParsedQuery`, `QueryTerm` in `domain/model/index.go`
  - Parser-Logik: `parseQuery()` in `domain/index/index.go` mit umfassender Dokumentation
  - Filter-Builder: `buildMongoFilterFromParsedQuery()`, `buildRegexFilterForTerm()`, `buildPrivateFileFilter()` in `repository/store/mongo.go`
  - **Unit Tests**: 34 Tests für Query-Parser und Filter-Builder (alle bestanden ✅)
  - **Indizes**: MongoDB B-tree Indizes auf manufacturer, model, subtitle, description, tags
  - **Dokumentation**: `SEARCH.md` mit Beispielen, Performance-Tipps, Index-Management
  - **Beispiel-Queries**: 
    - `"fend*"` → Prefix-Match (fender, fendora, etc.)
    - `"fender +bassman"` → Beide Begriffe müssen vorhanden sein (UND)
    - `"fend* +bassm*"` → Prefix mit UND-Verknüpfung
    - `"+tube -broken"` → Muss "tube" haben, darf nicht "broken" haben

## 0.2.36 - 2026-05-18 (Backend & Frontend)

- **REST-konforme API**: Dokumenten-Erstellungs-Endpoint standardisiert
  - Backend: `POST /api/v1/documents` für Dokumentenerstellung (REST-konform)
  - Frontend: Entsprechende Anpassung in Upload-Dialog
  - **Ergebnis**: API folgt jetzt konsistent RESTful-Konventionen

## 0.2.33 - 2026-05-16 (Backend)

- **Unit Tests für TIFF-Konvertierung**: Umfangreiche Test-Suite hinzugefügt
  - `converter_test.go`: 11 Unit Tests für `convertTiffToPng()` und `isTiffMimeType()`
    - Tests für erfolgreiche Konvertierung, ungültige Daten, Corrupted data
    - Tests für verschiedene TIFF-MIME-Type-Varianten
    - Tests für Größen, Farbtreue und Performance
  - `download_handler_test.go`: 5 Integrations-Tests für downloadFile-Endpoint
    - Tests für ?format=png Parameter
    - Tests für alle TIFF-MIME-Type-Varianten
    - Tests für Non-TIFF Dateien (sollten nicht konvertiert werden)
    - Tests für Fehlerfall bei korrupten TIFF-Daten
  - **Alle 16 Tests bestehen erfolgreich**

## 0.2.32 - 2026-05-16 (Backend & Frontend)

- **TIFF-Viewer-Unterstützung**: Automatische Konvertierung TIFF→PNG für Viewer
  - Backend: Neue `convertTiffToPng()` Funktion mit `golang.org/x/image/tiff` Dekodierung
  - Backend: `downloadFile` Endpoint unterstützt `?format=png` Query-Parameter
  - Frontend: `SearchView.loadFileData()` verwendet `?format=png` für TIFF-Dateien
  - Frontend: `downloadImage()` lädt Original-TIFF für Download (keine Konvertierung)
  - **Ergebnis**: TIFF-Dateien werden im Viewer als PNG angezeigt, bleiben aber beim Download original

## 0.2.31 - 2026-05-16 (Frontend)

- **TIF/TIFF Datei-Unterstützung**: Vollständige Unterstützung für TIFF-Dateien hinzugefügt
  - `UploadDialog`: Accept-Attribut erweitert um `.tif` und `.tiff` Dateitypen
  - `UploadDialog`: Robuste `getMimeTypeFromFilename()` Funktion hinzugefügt
    - Bestimmt MIME-Type basierend auf Dateiendung (Fallback bei Browser-Inkompatibilität)
    - Unterstützt: PDF, JPEG, PNG, GIF, BMP, TIF/TIFF
  - `SearchView`: Viewer nutzt bereits alle TIFF-MIME-Type-Varianten (`image/tiff`, `image/x-tiff`, `image/vnd.tiff`)
  - Dokumentation aktualisiert

## 0.2.30 - 2026-05-15 (Backend & Frontend)

- **AppFooter**: Version-Information hinzugefügt
  - Zeigt aktuelle App- und Backend-Version zwischen Copyright und Footer-Links an
  - Format: "Versionen: App X.X.X, Backend X.X.X"
  - Backend-Version wird von `/api/v1/info` Endpoint abgerufen
  - Subtile Darstellung mit kleinerer Schrift (0.85rem) und hellerer Farbe

- **Info Endpoint**: Struktur aktualisiert für Frontend-Kompatibilität
  - Rückgabe von `version` und `status` im `/api/v1/info` Endpoint
  - Version wird dynamisch aus Konfiguration gelesen

## 0.2.28 - 2026-05-15 (Backend)

- **Effect Update**: Altes Bild wird beim Update gelöscht
  - `updateEffect` Handler: Bevor neues Bild gespeichert wird, wird altes mit `blob.DeleteByInfo()` gelöscht
  - Verhindert Speicherverschwendung durch verwaiste Blob-Dateien
  - Fehler beim Löschen wird geloggt, aber nicht fatal

## 0.2.19 - 2026-05-15 (Frontend)

- **EditEffectView**: Bildbereich nach KISS-Prinzip vereinfacht
  - Entfernt: PrimeVue FileUpload-Komponente mit komplexen Buttons
  - Hinzugefügt: Einfacher HTML `<input type="file">` Input
  - Verhaltensweise: Altes Bild bleibt, wenn keine neue Datei gewählt
  - Nach Speichern: loadEffect() lädt Updated-Daten + neues Bild

- **EffectsView**: Image-Feld-Referenzen korrigiert
  - Behob: `images` (plural) → `image` (singular) - Backend gibt einzelnes Bild zurück
  - 3 Stellen: Tabelle, Image Modal, Detail Modal
  - Bilder werden jetzt korrekt in Tabelle und Detailansicht angezeigt

- **EffectsView**: "Vorschlag"-Button hinzugefügt
  - Neuer Button oben rechts neben "Effektdatenbank"-Titel
  - Öffnet E-Mail-Client mit vorausgefülltem Subject "Vorschlag: Effekt"
  - Adresse: info@wk-music.de
  - Benutzer können Effektvorschläge einreichen

## 0.2.23 - 2026-05-14

- **Backend**: Document Owner automatisch vom angemeldeten User
  - `POST /api/v1/documents/index`: Owner wird vom JWT-Token gesetzt, nicht vom Frontend
  - Entfernt Validierung: `doc.Owner` wird immer vom `getAuthenticatedUser()` überschrieben
  - Sicherheit: Frontend kann Owner nicht mehr manipulieren

- **Frontend**: Owner-Feld aus Upload-Dialog entfernt
  - Benutzer können Owner nicht mehr manuell eingeben
  - Automatische Zuweisung vom Backend beim Upload
  - Vereinfachte UI: Nur noch "Privates Dokument" Toggle (ohne Owner-Eingabe)

## 0.2.22 - 2026-05-13 (Backend)

- **Version Management**: Refaktoriert nach Go Best Practices
  - Zentrale `internal/version/version.go` statt hardcodierter Konstanten
  - BuildTime, Commit, ClientBasePath via ldflags injiziert (Build-Zeit)
  - Keine Runtime Environment Variablen mehr nötig
  
- **Startup Logging**: Ausführliche strukturierte Logs
  - Version, BuildTime, Commit, ClientBasePath angezeigt
  - Sanitized Config-Info (MongoDB Hosts, Database, Paths)
  - Doppelt-Ausgabe von Ports entfernt

- **Docker Health Check**: Gefixt
  - Nutzt `/readyz` Endpoint (nicht `/health`)
  - Start-Period: 10 Sekunden
  - `/health` als Alias Route hinzugefügt

- **Path Normalisierung**: Apache Reverse-Proxy Robustheit
  - `path.Clean` Middleware normalisiert `//api/v1/...` → `/api/v1/...`
  - Workaround für Apache ProxyPass ohne trailing slash

- **Connector-Bilder**: URL-Decode-Bug gefixt
  - `url.QueryUnescape` → `url.PathUnescape`
  - `%2B` (Plus) wird jetzt korrekt als `+` dekodiert, nicht als Leerzeichen

## 0.2.7 - 2026-05-13 (Frontend)

- **API URLs**: `__API_BASE__` globale Konstante
  - Alle direkten `/api/v1/` URLs nutzen jetzt `__API_BASE__` Präfix
  - EffectsView, RegisterView: Bild-URLs korrekt mit Base-Path
  - Reverse-Proxy Deployment transparent

- **Version Management**: Zentralisiert
  - Single Source of Truth: `package.json` (version: "0.2.7")
  - Vite injiziert Version zur Build-Zeit als `__APP_VERSION__`
  - `config.js` nutzt injizierte Konstante

- **Lizenz**: MIT License hinzugefügt

## 0.2.13 - 2026-05-13

- **Frontend**: Manufacturer und Model auch editierbar
  - AutoComplete für Hersteller-Vorschläge
  - Validierung erforderlich für beide Felder
  
- **Backend**: PATCH Endpoint erweitert
  - Manufacturer und Model können aktualisiert werden
  - Validierung: TrimSpace um Leerzeichen zu entfernen

## 0.2.12 - 2026-05-13

- **Frontend**: Dokument-Edit Dialog
  - Neue EditDialog.vue Komponente (ähnlich wie UploadDialog)
  - Edit-Button in SearchView (nur für angemeldete Benutzer mit selektiertem Dokument)
  - Hersteller und Modell sind schreibgeschützt
  - Tags und Beschreibung können bearbeitet werden
  - Neue Files können hinzugefügt werden (mit Base64 data)
  - Bestehende Files können gelöscht werden (Soft-Delete)
  - Toggle-Button um gelöschte Files wiederherzustellen
  
- **Backend**: PATCH /api/v1/documents/{id} Endpoint
  - Permission Check: Admin oder Owner
  - Nur neue Files werden als Base64 in Blob gespeichert
  - Bestehende Files bleiben unverändert
  - Gelöschte Files: blob.DeleteByInfo() aufgerufen, dann aus Document entfernt
  - Timestamps aktualisiert (LastModifiedAt)
  - Only editable fields: subtitle, tags, description
  - Manufacturer und Model sind schreibgeschützt

- **Domain**: ContainerInfo um Deleted Feld erweitert
  - Konsistenz mit Blob-Metadaten

## 0.2.11 - 2026-05-13

- **Backend**: Deleted-Flag in INF-Dateien (Blob Metadata)
  - containerInfoEntry Struct um `Deleted bool` Feld erweitert (mit omitempty)
  - LoadContainerInfos lädt jetzt auch das Deleted Flag
  - Neue BlobService.DeleteByInfo() Methode:
    - Erhält ContainerInfo Parameter
    - Lädt INF-Datei des Containers
    - Findet Eintrag nach Offset und Length
    - Markiert Eintrag als deleted
    - Speichert aktualisierte INF-Datei
  - DELETE /api/v1/documents/{id}:
    - Ruft blob.DeleteByInfo() für alle Files auf
    - Markiert Files persistent in .inf Dateien
    - Ermöglicht späteren Restore (noch nicht implementiert)

## 0.2.10 - 2026-05-13

- **Backend**: Soft-Delete für Dokument-Files
  - ContainerInfo Model um `Deleted bool` Feld erweitert (mit `omitempty`)
  - Rückwärtskompatibilität: alte INF-Dateien ohne Deleted Feld funktionieren
  - DELETE `/api/v1/documents/{id}`:
    - Markiert alle Files als deleted (Deleted: true)
    - Speichert Metadaten mit gelöschtem Status
    - Löscht dann das Dokument aus der DB
  - GET `/api/v1/documents/{id}/files/{filename}`:
    - Überprüft ob File gelöscht ist
    - Antwortet mit 404 wenn File gelöscht

## 0.2.9 - 2026-05-13

- **Frontend**: Dokument-Löschung in SearchView
  - Lösch-Button (rot) neben Upload-Button
  - Nur sichtbar für angemeldete Benutzer
  - Bestätigungsdialog vor dem Löschen
  - Aktualisiert die Trefferliste nach erfolgreichem Löschen
- **Backend**: DELETE `/api/v1/documents/{id}` Endpoint
  - Authentifizierung erforderlich (Bearer Token)
  - Berechtigungen:
    - Admin: Kann alle Dokumente löschen
    - Benutzer: Kann nur seine eigenen Dokumente löschen (Owner Feld)
  - 403 Forbidden wenn User nicht Eigentümer oder Admin
  - 404 wenn Dokument nicht existiert
  - MongoDB DeleteByID Methode in Store implementiert

## 0.2.8 - 2026-05-13

- **Frontend**: Effekt-Upload nur für angemeldete Benutzer
  - "Effekt hinzufügen" Button wird nur angezeigt wenn `isLoggedIn` true
  - useAuth() composable Integration in EffectsView
- **Backend**: POST `/api/v1/effects` geschützt durch authMiddleware
  - 401 Unauthorized wenn kein Bearer Token
  - 401 Unauthorized wenn Token ungültig
  - Nur authentifizierte Benutzer können Effects erstellen

## 0.2.7 - 2026-05-12

- **Frontend UI**: Button-Styling vereinheitlicht
  - Suche-Button: Blau (primary), eckig mit gerundeten Kanten, Tooltip "Suchen"
  - Upload-Button (EffectsView): Grün (success), eckig mit gerundeten Kanten, Tooltip "Effekt hinzufügen"
  - Upload-Button (SearchView): Grün (success), eckig mit gerundeten Kanten, Tooltip "Upload"
  - Upload-Button (EffectUploadDialog): Grün (success), eckig mit gerundeten Kanten, Tooltip "Hochladen"
  - Abbrechen-Button: Grau (secondary), eckig mit gerundeten Kanten, Tooltip "Abbrechen"
  - Tooltips: Einheitlich `v-tooltip.bottom` statt `:title` Attribut
  - Entfernt `rounded` Attribut für konsistentes Aussehen mit ecking-gerundeten Kanten
- **Backend**: Sortierung für Effects Tabelle auf Model-Spalte korrigiert
  - MongoDB $sort duplicate key error behoben
  - Sekundäre Sortierung nach Model nur wenn nicht Hauptsortfeld

## 0.2.6 - 2026-05-12

- **Frontend & Backend**: Sortierung für Effects Tabelle implementiert
  - Sortierbar nach: Typ, Hersteller, Modell, Spannung, Strom
  - Frontend: DataTable mit `@sort` Event und `sortable` Attribute auf Columns
  - API: `/api/v1/effects/search` akzeptiert `sort` (Feldname) und `order` (asc/desc) Parameter
  - Backend: SearchEffects Repository Methode um Sorting unterstützt
  - Helper-Funktion `mapEffectSortField()` für Feldname-Mapping
  - Sekundäre Sortierung nach Modell für Konsistenz
- **Frontend**: Typ-Spalte zeigt jetzt i18n Übersetzungen (z.B. "Verzerrung" statt "Distortion")
  - Lokaler Lookup über `effectTypeMap`
  - auch im Detail-Modal
- **Bug Fixes**:
  - Manufacturer AutoComplete zeigt jetzt korrekt Vorschläge aus API

## 0.2.5 - 2026-05-12

- **Backend**: EffectTypes jetzt vollständig aus MongoDB geladen
  - `GetAllEffectTypes()` Datenbankfunktion aktiv (statt hardcodierte Liste)
  - Collection: `effecttypes`
  - Sortierung nach TypeName aufsteigend
  - Unterstützt i18n (de/en Übersetzungen)
- **Frontend**: EffectTypes-Dropdown funktioniert vollständig
  - Lädt Types von `GET /api/v1/effecttypes`
  - Zeigt deutschsprachige Namen an
  - Speichert TypeName als Wert
  - Integriert in EffectUploadDialog

## 0.2.4 - 2026-05-12

- **Frontend**: Effects Effects Upload Dialog Refactoring und Bug Fixes
  - Neue Komponente `EffectUploadDialog.vue` - Upload-Logik in separate, wiederverwendbare Komponente ausgelagert
  - Komponente nutzt v-model für visibility binding
  - Props: `visible` (Boolean), `effectTypes` (Array)
  - Events: `@update:visible`, `@effect-created`
  - Form-Validierung: Typ, Hersteller, Modell, Anschluss sind erforderlich
  - Manufacturer AutoComplete mit API-Vorschläge
  - Connector Dropdown mit 9 Optionen
  - File Upload mit Bildvalidation
  - Error Handling und User Feedback
- **Frontend**: EffectsView Cleanup
  - Entfernt alte Upload-Funktionen nach Refactoring
  - Vereinfachte State Management (nur noch `showUploadDialog` ref)
  - Integration mit neuer EffectUploadDialog Komponente
  - Fehlerbehandlung für null effectTypes robuster gemacht
- **Bug Fixes**:
  - Dialog close buttons (X-Button) funktionieren jetzt korrekt
  - Form wird bei Dialog-Hide korrekt zurückgesetzt
  - Manufacturer Autocomplete zeigt jetzt Vorschläge an
  - Connector Dropdown Wert-Binding korrigiert
  - Null-Safety Checks für alle Form-Felder
  - Props Validation vereinfacht und robuster

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