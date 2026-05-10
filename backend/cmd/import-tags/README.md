# Tag Import Tool

Importiert Tags aus JSON-Dateien in die MongoDB `tags` Collection.

## Verwendung

```bash
cd backend
go run cmd/import-tags/main.go [-tags-dir path/to/tags]
```

## Parameter

- `-tags-dir`: Verzeichnis mit den Tag-JSON-Dateien (Standard: `testdata/tags`)

## Eingabeformat

Jede JSON-Datei im Tags-Verzeichnis muss folgendes Format haben:

```json
{
  "name": "effects",
  "count": 0
}
```

## Was das Tool macht

1. Verbindung zur MongoDB aufbauen (aus `configs/service.yaml`)
2. Alle bestehenden Tags aus der `tags` Collection löschen
3. Alle JSON-Dateien aus dem Tags-Verzeichnis lesen
4. Tags normalisieren (Kleinbuchstaben, Whitespace trimmen)
5. Das "name" Feld wird als `_id` in MongoDB gespeichert
6. Tags in die MongoDB importieren

**Beispiel MongoDB Document:**
```json
{
  "_id": "effects",
  "count": 0
}
```

## Beispiel

```bash
$ go run cmd/import-tags/main.go
Imported tags from 169 JSON files
✓ Successfully imported 169 tags
```

## Konfiguration

Das Tool liest die MongoDB-Konfiguration automatisch aus `configs/service.yaml`.
Keine zusätzliche Konfiguration erforderlich.
