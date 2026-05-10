// Package main provides a one-time import tool for migrating documents
// from the old WilliesSchematicsWorld backup format into the Schematic2 system.
//
// Backup format (per document folder):
//   - schematic.json   – metadata (id, manufacturer, model, tags, files map, …)
//   - *.pdf / *.png /… – the actual document files referenced by the files map
//
// The tool reads each folder, stores the physical files in the Schematic2
// blob store, and upserts the document metadata into MongoDB.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/samber/do/v2"
	"github.com/willie68/schematic2/backend/internal/config"
	"github.com/willie68/schematic2/backend/internal/domain"
	"github.com/willie68/schematic2/backend/internal/logging"
	"github.com/willie68/schematic2/backend/internal/repository/store"
	"github.com/willie68/schematic2/backend/internal/services/blob"
)

// backupSchematic is the JSON structure of the old server backup.
type backupSchematic struct {
	ID             string            `json:"id"`
	ForeignID      string            `json:"foreignId"`
	CreatedAt      time.Time         `json:"createdAt"`
	LastModifiedAt time.Time         `json:"lastModifiedAt"`
	Manufacturer   string            `json:"manufacturer"`
	Model          string            `json:"model"`
	Subtitle       string            `json:"subtitle"`
	Tags           []string          `json:"tags"`
	Description    string            `json:"description"`
	PrivateFile    bool              `json:"privateFile"`
	Owner          string            `json:"owner"`
	Files          map[string]string `json:"files"` // filename -> old-blob-id (ignored here)
}

func main() {
	backupDir := flag.String("backup-dir", "Y:/schematics/backup/schematics", "Root directory of the schematic backup (contains one folder per document)")
	dryRun := flag.Bool("dry-run", false, "Process and validate data but do not write to MongoDB or blob store")
	skipExisting := flag.Bool("skip-existing", true, "Skip documents that already exist in MongoDB")
	maxErrors := flag.Int("max-errors", 50, "Abort after this many import errors (0 = unlimited)")
	flag.Parse()

	cfg := config.LoadFromEnv()
	logging.Init(cfg.Logging)
	logger := logging.New("import-schematics")

	logger.Info("starting schematic backup import",
		"backup-dir", *backupDir,
		"dry-run", *dryRun,
		"skip-existing", *skipExisting,
	)

	// Set up DI container with blob store and document store
	inj := do.New()
	do.ProvideValue(inj, cfg)

	var blobSvc *blob.Service
	var docStore *store.MongoDocumentStore

	if !*dryRun {
		blobSvc = blob.New(inj)
		if err := blobSvc.Prepare(); err != nil {
			log.Fatalf("prepare blob store: %v", err)
		}
		do.ProvideValue(inj, blobSvc)

		docStore = store.NewMongoDocumentStore(inj)
		if err := docStore.Prepare(); err != nil {
			log.Fatalf("prepare document store: %v", err)
		}
		do.ProvideValue(inj, docStore)
	}

	// Collect all document directories
	entries, err := os.ReadDir(*backupDir)
	if err != nil {
		log.Fatalf("read backup dir %q: %v", *backupDir, err)
	}

	var dirs []os.DirEntry
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e)
		}
	}

	logger.Info("found document directories", "count", len(dirs))

	stats := importStats{}
	errCount := 0

	for i, d := range dirs {
		docDir := filepath.Join(*backupDir, d.Name())
		logCtx := logger.With("dir", d.Name(), "progress", fmt.Sprintf("%d/%d", i+1, len(dirs)))

		err := importDocument(docDir, blobSvc, docStore, *skipExisting, *dryRun, logCtx.Handler())
		switch {
		case errors.Is(err, errSkipped):
			stats.skipped++
		case err != nil:
			logCtx.Error("import failed", "err", err)
			stats.failed++
			errCount++
			if *maxErrors > 0 && errCount >= *maxErrors {
				logger.Error("too many errors, aborting", "max-errors", *maxErrors)
				break
			}
		default:
			stats.imported++
		}
	}

	fmt.Printf("\n=== Import complete ===\n")
	fmt.Printf("  Imported : %d\n", stats.imported)
	fmt.Printf("  Skipped  : %d\n", stats.skipped)
	fmt.Printf("  Failed   : %d\n", stats.failed)
	if *dryRun {
		fmt.Println("  (dry-run – no data was written)")
	}
}

type importStats struct {
	imported int
	skipped  int
	failed   int
}

var errSkipped = errors.New("skipped")

func importDocument(
	docDir string,
	blobSvc *blob.Service,
	docStore *store.MongoDocumentStore,
	skipExisting bool,
	dryRun bool,
	_ any, // log handler – kept for future structured log attachment
) error {
	// Read JSON metadata
	jsonPath := filepath.Join(docDir, "schematic.json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("read schematic.json: %w", err)
	}

	var bk backupSchematic
	if err := json.Unmarshal(data, &bk); err != nil {
		return fmt.Errorf("unmarshal schematic.json: %w", err)
	}

	if strings.TrimSpace(bk.ID) == "" {
		return errors.New("document has empty id")
	}
	if strings.TrimSpace(bk.Manufacturer) == "" || strings.TrimSpace(bk.Model) == "" {
		return fmt.Errorf("document %q: manufacturer/model missing", bk.ID)
	}

	// Check if document exists already
	if skipExisting && !dryRun && docStore != nil {
		exists, err := docStore.Exists(context.Background(), bk.ID)
		if err != nil {
			return fmt.Errorf("check exists: %w", err)
		}
		if exists {
			return errSkipped
		}
	}

	// Build domain document files by reading the physical files
	var docFiles []domain.DocumentFile
	for filename := range bk.Files {
		filePath := filepath.Join(docDir, filename)

		if dryRun {
			// In dry-run mode only check existence, do not read file contents
			if _, err := os.Stat(filePath); err != nil {
				log.Printf("warning: document %q: file %q not found on disk: %v", bk.ID, filename, err)
			}
			docFiles = append(docFiles, domain.DocumentFile{
				Name:     filename,
				MIMEType: mimeTypeForFile(filename),
				Type:     "schematic",
			})
			continue
		}

		fileData, err := os.ReadFile(filePath)
		if err != nil {
			// File referenced in JSON but not present on disk – warn and skip
			log.Printf("warning: document %q: file %q not found on disk: %v", bk.ID, filename, err)
			continue
		}

		mimeType := mimeTypeForFile(filename)

		docFile := domain.DocumentFile{
			Name:     filename,
			MIMEType: mimeType,
			Type:     "schematic",
		}

		if !dryRun && blobSvc != nil {
			info, err := blobSvc.Save(fileData, mimeType)
			if err != nil {
				return fmt.Errorf("save blob %q: %w", filename, err)
			}
			docFile.Container = info
		}

		docFiles = append(docFiles, docFile)
	}

	doc := domain.Document{
		ID:             bk.ID,
		CreatedAt:      bk.CreatedAt,
		LastModifiedAt: bk.LastModifiedAt,
		Manufacturer:   strings.TrimSpace(bk.Manufacturer),
		Model:          strings.TrimSpace(bk.Model),
		Subtitle:       strings.TrimSpace(bk.Subtitle),
		Tags:           normalizeTags(bk.Tags),
		Description:    strings.TrimSpace(bk.Description),
		PrivateFile:    bk.PrivateFile,
		Owner:          strings.TrimSpace(bk.Owner),
		Files:          docFiles,
	}

	if !dryRun && docStore != nil {
		if err := docStore.Upsert(doc); err != nil {
			return fmt.Errorf("upsert document: %w", err)
		}
	}

	return nil
}

// normalizeTags trims whitespace and removes empty/duplicate tags.
func normalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	result := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		key := strings.ToLower(t)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, t)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// mimeTypeForFile returns the MIME type based on the file extension.
func mimeTypeForFile(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if t := mime.TypeByExtension(ext); t != "" {
		return t
	}
	// Fallback for common types not always registered by the OS
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".tif", ".tiff":
		return "image/tiff"
	case ".svg":
		return "image/svg+xml"
	case ".zip":
		return "application/zip"
	}
	return "application/octet-stream"
}
