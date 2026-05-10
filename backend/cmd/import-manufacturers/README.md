# Manufacturers Import Tool

Ein einmaliges Import-Tool zum Laden von Hersteller-Daten aus JSON-Dateien in die MongoDB.

## Verwendung

```bash
cd backend
go run cmd/import-manufacturers/main.go [-manufacturers-dir testdata/manufacturers]
```

**Optionen:**
- `-manufacturers-dir`: Verzeichnis mit Hersteller-JSON-Dateien (Standard: `testdata/manufacturers`)

## Dateiformat

Jede JSON-Datei enthält einen Hersteller:

```json
{"name":"Valco","count":0}
```

**Felder:**
- `name` (string): Der Name des Herstellers (wird als MongoDB `_id` gespeichert)
- `count` (integer): Wird ignoriert, nur für Struktur vorhanden

## Verarbeitung

1. **Normalisierung**: Der Name wird trimmt, die **Originalschreibweise bleibt erhalten** (Case-Preserved)
   - ❌ Keine Kleinbuchstaben-Konvertierung (anders als Tags!)
   - ✅ `"Samsung"` bleibt `"Samsung"`, nicht `"samsung"`

2. **MongoDB-Speichern**: Der Name wird als `_id` des Dokuments gespeichert
   ```bson
   {
     "_id": "Samsung"
   }
   ```

3. **Clearing**: Alle existierenden Hersteller werden vor dem Import gelöscht

## Ausgabe

```
Imported manufacturers from 402 JSON files
✓ Successfully imported 402 manufacturers
```

## MongoDB-Struktur

**Collection:** `manufacturers`

**Dokument-Format:**
```bson
{
  "_id": "Samsung"      // Case-preserved name
}
```

## Unterschiede zu Tag-Import

| Aspekt | Tags | Manufacturers |
|--------|------|---|
| **Collection** | `tags` | `manufacturers` |
| **Normalisierung** | Kleinbuchstaben | Case-Preserved |
| **Feld-Name** | `_id` | `_id` |
| **Beispiel** | `{"_id": "effects"}` | `{"_id": "Samsung"}` |

## Fehlerbehandlung

- Ungültige JSON-Dateien werden geloggt und übersprungen
- Dateien ohne `.json`-Extension werden ignoriert
- Leere Hersteller-Namen werden übersprungen
- Alle Fehler werden auf stderr ausgegeben
