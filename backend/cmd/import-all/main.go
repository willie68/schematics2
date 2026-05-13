// Package main provides a unified import tool for migrating all data
// (manufacturers, tags, schematics, effect types, and effects) from the backup into the Schematic2 system.
//
// Directory structure (under base-dir):
//   - manufacturers/     – JSON files with manufacturer data
//   - tags/              – JSON files with tag data
//   - schematics/        – One folder per document with schematic.json and files
//   - effecttypes/       – One folder per effect type with effecttype.json and image file
//   - effects/           – One folder per effect with effect.json and image file
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/samber/do/v2"
	"github.com/willie68/schematic2/backend/internal/config"
	"github.com/willie68/schematic2/backend/internal/domain"
	"github.com/willie68/schematic2/backend/internal/logging"
	"github.com/willie68/schematic2/backend/internal/repository/blob"
	"github.com/willie68/schematic2/backend/internal/repository/store"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Manufacturer represents a manufacturer record
type Manufacturer struct {
	Name  string `json:"name" bson:"name"`
	Count int    `json:"count" bson:"count"`
}

// Tag represents a tag record
type Tag struct {
	Name  string `json:"name" bson:"name"`
	Count int    `json:"count" bson:"count"`
}

// backupSchematic is the JSON structure of the old server backup
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

// backupEffectType is the JSON structure for effect types
type backupEffectType struct {
	ID             string            `json:"id"`
	ForeignID      string            `json:"foreignId"`
	CreatedAt      string            `json:"createdAt"`
	LastModifiedAt string            `json:"lastModifiedAt"`
	TypeName       string            `json:"typeName"`
	Nls            map[string]string `json:"nls"`
	TypeImage      string            `json:"typeImage"`
}

// backupEffect is the JSON structure for effects
type backupEffect struct {
	ID             string   `json:"id"`
	ForeignID      string   `json:"foreignId"`
	CreatedAt      string   `json:"createdAt"`
	LastModifiedAt string   `json:"lastModifiedAt"`
	EffectType     string   `json:"effectType"`
	Manufacturer   string   `json:"manufacturer"`
	Model          string   `json:"model"`
	Tags           []string `json:"tags"`
	Comment        string   `json:"comment"`
	Image          string   `json:"image"`
	Connector      string   `json:"connector"`
	Voltage        string   `json:"voltage"`
	Current        string   `json:"current"`
}

func main() {
	baseDir := flag.String("base-dir", "testdata", "Base directory containing manufacturers/, tags/, schematics/, effecttypes/, and effects/ subdirectories")
	importMfg := flag.Bool("manufacturers", true, "Import manufacturers from manufacturers/ subdirectory")
	importTgs := flag.Bool("tags", true, "Import tags from tags/ subdirectory")
	importSch := flag.Bool("schematics", true, "Import schematics from schematics/ subdirectory")
	importEft := flag.Bool("effecttypes", true, "Import effect types from effecttypes/ subdirectory")
	importEff := flag.Bool("effects", true, "Import effects from effects/ subdirectory")
	dryRun := flag.Bool("dry-run", false, "Process and validate data but do not write to MongoDB or blob store")
	skipExisting := flag.Bool("skip-existing", true, "Skip documents that already exist in MongoDB")
	maxErrors := flag.Int("max-errors", 50, "Abort after this many import errors (0 = unlimited)")
	flag.Parse()

	cfg := config.LoadFromEnv()
	logging.Init(cfg.Logging)
	logger := logging.New("import-all")

	logger.Info("starting unified import",
		"base-dir", *baseDir,
		"manufacturers", *importMfg,
		"tags", *importTgs,
		"schematics", *importSch,
		"effecttypes", *importEft,
		"effects", *importEff,
		"dry-run", *dryRun,
		"skip-existing", *skipExisting,
	)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := connectMongo(ctx, cfg.MongoDB)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Warning: error disconnecting from MongoDB: %v", err)
		}
	}()

	db := client.Database(cfg.MongoDB.Database)

	// Set up DI container for schematics import
	inj := do.New()
	do.ProvideValue(inj, cfg)

	var blobSvc *blob.Service
	var docStore *store.MongoStore

	if !*dryRun && *importSch {
		blobSvc = blob.New(inj)
		if err := blobSvc.Prepare(); err != nil {
			log.Fatalf("prepare blob store: %v", err)
		}
		do.ProvideValue(inj, blobSvc)

		docStore = store.NewMongoStore(inj)
		if err := docStore.Prepare(); err != nil {
			log.Fatalf("prepare document store: %v", err)
		}
		do.ProvideValue(inj, docStore)
	}

	// Import manufacturers
	if *importMfg {
		manufDir := filepath.Join(*baseDir, "manufacturers")
		if _, err := os.Stat(manufDir); err == nil {
			logger.Info("importing manufacturers", "path", manufDir)
			manufCol := db.Collection("manufacturers")
			imported, err := importManufacturers(ctx, manufCol, manufDir)
			if err != nil {
				logger.Error("failed to import manufacturers", "err", err)
			} else {
				fmt.Printf("✓ Successfully imported %d manufacturers\n", imported)
			}
		} else {
			logger.Warn("manufacturers directory not found", "path", manufDir)
		}
	}

	// Import tags
	if *importTgs {
		tagsDir := filepath.Join(*baseDir, "tags")
		if _, err := os.Stat(tagsDir); err == nil {
			logger.Info("importing tags", "path", tagsDir)
			tagsCol := db.Collection("tags")
			imported, err := importTags(ctx, tagsCol, tagsDir)
			if err != nil {
				logger.Error("failed to import tags", "err", err)
			} else {
				fmt.Printf("✓ Successfully imported %d tags\n", imported)
			}
		} else {
			logger.Warn("tags directory not found", "path", tagsDir)
		}
	}

	// Import schematics
	if *importSch {
		schemsDir := filepath.Join(*baseDir, "schematics")
		if _, err := os.Stat(schemsDir); err == nil {
			logger.Info("importing schematics", "path", schemsDir)
			imported, skipped, failed := importSchematics(ctx, schemsDir, blobSvc, docStore, *skipExisting, *dryRun, *maxErrors, logger)
			fmt.Printf("\n=== Schematics Import complete ===\n")
			fmt.Printf("  Imported : %d\n", imported)
			fmt.Printf("  Skipped  : %d\n", skipped)
			fmt.Printf("  Failed   : %d\n", failed)
			if *dryRun {
				fmt.Println("  (dry-run – no data was written)")
			}
		} else {
			logger.Warn("schematics directory not found", "path", schemsDir)
		}
	}

	// Import effect types
	if *importEft {
		effectTypesDir := filepath.Join(*baseDir, "effecttypes")
		if _, err := os.Stat(effectTypesDir); err == nil {
			logger.Info("importing effect types", "path", effectTypesDir)
			effectTypesCol := db.Collection("effecttypes")
			imagesDir := "internal/repository/effecttypes"
			imported, err := importEffectTypes(ctx, effectTypesCol, effectTypesDir, imagesDir, *dryRun)
			if err != nil {
				logger.Error("failed to import effect types", "err", err)
			} else {
				fmt.Printf("✓ Successfully imported %d effect types\n", imported)
				if !*dryRun {
					fmt.Printf("✓ Images copied to %s\n", imagesDir)
				}
			}
		} else {
			logger.Warn("effect types directory not found", "path", effectTypesDir)
		}
	}

	// Import effects
	if *importEff {
		effectsDir := filepath.Join(*baseDir, "effects")
		if _, err := os.Stat(effectsDir); err == nil {
			logger.Info("importing effects", "path", effectsDir)
			imported, err := importEffectsData(ctx, db, effectsDir, blobSvc, *dryRun)
			if err != nil {
				logger.Error("failed to import effects", "err", err)
			} else {
				fmt.Printf("✓ Successfully imported %d effects\n", imported)
			}
		} else {
			logger.Warn("effects directory not found", "path", effectsDir)
		}
	}

	fmt.Println("\n✓ Import complete!")
}

func connectMongo(ctx context.Context, cfg config.MongoDB) (*mongo.Client, error) {
	uri := buildMongoURI(cfg)
	clientOpts := options.Client().ApplyURI(uri)
	if cfg.DirectConnection {
		clientOpts.SetDirect(true)
	}

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(context.Background())
		return nil, err
	}

	return client, nil
}

func buildMongoURI(cfg config.MongoDB) string {
	uri := "mongodb://"
	if cfg.Username != "" {
		uri += cfg.Username
		if cfg.Password != "" {
			uri += ":" + cfg.Password
		}
		uri += "@"
	}
	uri += strings.Join(cfg.GetHosts(), ",")
	if cfg.Database != "" {
		uri += "/" + cfg.Database
	}
	authDB := cfg.GetAuthDatabase()
	if authDB != "" {
		if cfg.Database == "" {
			uri += "/"
		}
		uri += "?authSource=" + authDB
	}
	return uri
}

// importManufacturers imports manufacturers from JSON files
func importManufacturers(ctx context.Context, col *mongo.Collection, dir string) (int, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("read directory: %w", err)
	}

	// Clear existing manufacturers
	_, err = col.DeleteMany(ctx, bson.D{})
	if err != nil {
		return 0, fmt.Errorf("delete existing manufacturers: %w", err)
	}

	count := 0
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(dir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Warning: failed to read %s: %v", file.Name(), err)
			continue
		}

		var manufacturer Manufacturer
		if err := json.Unmarshal(data, &manufacturer); err != nil {
			log.Printf("Warning: failed to unmarshal %s: %v", file.Name(), err)
			continue
		}

		// Trim and preserve case
		name := strings.TrimSpace(manufacturer.Name)
		if name == "" {
			continue
		}

		// Create BSON document with name as _id
		doc := bson.D{
			{Key: "_id", Value: name},
		}

		if _, err := col.InsertOne(ctx, doc); err != nil {
			log.Printf("Warning: failed to insert manufacturer %q: %v", name, err)
			continue
		}

		count++
	}

	return count, nil
}

// importTags imports tags from JSON files
func importTags(ctx context.Context, col *mongo.Collection, dir string) (int, error) {
	// Clear existing tags
	_, err := col.DeleteMany(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("failed to clear existing tags: %w", err)
	}

	// Read all JSON files from directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("failed to read tags directory: %w", err)
	}

	var tags []interface{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Warning: failed to read %s: %v\n", entry.Name(), err)
			continue
		}

		var tag Tag
		if err := json.Unmarshal(data, &tag); err != nil {
			fmt.Printf("Warning: failed to parse %s: %v\n", entry.Name(), err)
			continue
		}

		// Normalize tag name
		tag.Name = strings.ToLower(strings.TrimSpace(tag.Name))
		if tag.Name == "" {
			fmt.Printf("Warning: empty tag name in %s\n", entry.Name())
			continue
		}

		// Reset count to ensure fresh import
		tag.Count = 0

		tags = append(tags, bson.M{
			"_id":   tag.Name,
			"count": tag.Count,
		})
	}

	if len(tags) == 0 {
		fmt.Println("No tags found to import")
		return 0, nil
	}

	// Insert tags
	result, err := col.InsertMany(ctx, tags)
	if err != nil {
		return 0, fmt.Errorf("failed to insert tags: %w", err)
	}

	return len(result.InsertedIDs), nil
}

// importSchematics imports schematics from backup directories
func importSchematics(ctx context.Context, baseDir string, blobSvc *blob.Service, docStore *store.MongoStore, skipExisting, dryRun bool, maxErrors int, logger any) (imported, skipped, failed int) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		log.Fatalf("read backup dir %q: %v", baseDir, err)
	}

	var dirs []os.DirEntry
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e)
		}
	}

	errCount := 0
	for i, d := range dirs {
		docDir := filepath.Join(baseDir, d.Name())
		fmt.Printf("[%d/%d] Importing %s...\n", i+1, len(dirs), d.Name())

		err := importDocument(docDir, blobSvc, docStore, skipExisting, dryRun)
		switch {
		case errors.Is(err, errSkipped):
			skipped++
		case err != nil:
			log.Printf("Error importing %s: %v\n", d.Name(), err)
			failed++
			errCount++
			if maxErrors > 0 && errCount >= maxErrors {
				log.Printf("Too many errors (%d), aborting", maxErrors)
				break
			}
		default:
			imported++
		}
	}

	return imported, skipped, failed
}

var errSkipped = errors.New("skipped")

func importDocument(
	docDir string,
	blobSvc *blob.Service,
	docStore *store.MongoStore,
	skipExisting bool,
	dryRun bool,
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
			// In dry-run mode only check existence
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

// normalizeTags trims whitespace and removes empty/duplicate tags
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

// mimeTypeForFile returns the MIME type based on the file extension
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

// importEffectTypes imports effect types from backup directories
func importEffectTypes(ctx context.Context, col *mongo.Collection, dir, imagesOutDir string, dryRun bool) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("read directory: %w", err)
	}

	// Clear existing effect types
	if !dryRun {
		_, err = col.DeleteMany(ctx, bson.D{})
		if err != nil {
			return 0, fmt.Errorf("delete existing effect types: %w", err)
		}
	}

	// Create output directory if it doesn't exist
	if !dryRun {
		if err := os.MkdirAll(imagesOutDir, 0755); err != nil {
			return 0, fmt.Errorf("create images directory: %w", err)
		}
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		effectTypeDir := filepath.Join(dir, entry.Name())
		et, imageFile, err := parseEffectType(effectTypeDir)
		if err != nil {
			log.Printf("Warning: failed to parse effect type %s: %v", entry.Name(), err)
			continue
		}

		if et == nil {
			continue
		}

		// Copy image file if provided
		if imageFile != "" && !dryRun {
			srcImage := filepath.Join(effectTypeDir, imageFile)
			dstImage := filepath.Join(imagesOutDir, imageFile)

			if err := copyFile(srcImage, dstImage); err != nil {
				log.Printf("Warning: failed to copy image %s: %v", imageFile, err)
				// Continue anyway - the effect type is still valid
			}
		}

		// Insert into MongoDB
		if !dryRun {
			doc := bson.M{
				"_id":            et.ID,
				"createdAt":      et.CreatedAt,
				"lastModifiedAt": et.LastModifiedAt,
				"typeName":       et.TypeName,
				"i18n":           et.I18n,
				"typeImage":      et.TypeImage,
			}

			if _, err := col.InsertOne(ctx, doc); err != nil {
				log.Printf("Warning: failed to insert effect type %q: %v", et.ID, err)
				continue
			}
		}

		count++
	}

	return count, nil
}

// parseEffectType reads and parses an effect type from a directory
// Returns the parsed EffectType, the image filename, and any error
func parseEffectType(dir string) (*domain.EffectType, string, error) {
	jsonFile := filepath.Join(dir, "effecttype.json")
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return nil, "", fmt.Errorf("read effecttype.json: %w", err)
	}

	var backup backupEffectType
	if err := json.Unmarshal(data, &backup); err != nil {
		return nil, "", fmt.Errorf("unmarshal effecttype.json: %w", err)
	}

	if strings.TrimSpace(backup.ID) == "" {
		return nil, "", fmt.Errorf("effect type has empty id")
	}

	if strings.TrimSpace(backup.TypeName) == "" {
		return nil, "", fmt.Errorf("effect type has empty typeName")
	}

	// Parse timestamps
	createdAt, err := time.Parse(time.RFC3339, backup.CreatedAt)
	if err != nil {
		createdAt = time.Now().UTC()
	}

	lastModifiedAt, err := time.Parse(time.RFC3339, backup.LastModifiedAt)
	if err != nil {
		lastModifiedAt = time.Now().UTC()
	}

	et := &domain.EffectType{
		ID:             backup.ID,
		CreatedAt:      createdAt,
		LastModifiedAt: lastModifiedAt,
		TypeName:       strings.TrimSpace(backup.TypeName),
		I18n:           backup.Nls, // Map nls to i18n
		TypeImage:      strings.TrimSpace(backup.TypeImage),
	}

	return et, backup.TypeImage, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}

	return dstFile.Sync()
}

// importEffectsData imports effects from backup directories
func importEffectsData(ctx context.Context, db *mongo.Database, dir string, blobSvc *blob.Service, dryRun bool) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("read directory: %w", err)
	}

	effectsCol := db.Collection("effects")
	effectTypesCol := db.Collection("effecttypes")
	manufacturersCol := db.Collection("manufacturers")

	// Clear existing effects
	if !dryRun {
		_, err = effectsCol.DeleteMany(ctx, bson.D{})
		if err != nil {
			return 0, fmt.Errorf("delete existing effects: %w", err)
		}
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		effectDir := filepath.Join(dir, entry.Name())
		eff, imageFile, err := parseEffectData(effectDir)
		if err != nil {
			log.Printf("Warning: failed to parse effect %s: %v", entry.Name(), err)
			continue
		}

		if eff == nil {
			continue
		}

		// Validate effectType exists in effecttypes collection (case-insensitive)
		if eff.EffectType != "" {
			// Use regex for case-insensitive search on typeName field
			filter := bson.M{
				"typeName": bson.M{
					"$regex":   "^" + regexp.QuoteMeta(eff.EffectType) + "$",
					"$options": "i", // case-insensitive
				},
			}
			cnt, _ := effectTypesCol.CountDocuments(ctx, filter)
			if cnt == 0 {
				log.Printf("Warning: effect %q: effect type %q not found in effecttypes collection", eff.ID, eff.EffectType)
				continue
			}
		}

		// Validate/create manufacturer in manufacturers collection
		if eff.Manufacturer != "" && !dryRun {
			manufCount, _ := manufacturersCol.CountDocuments(ctx, bson.M{"_id": eff.Manufacturer})
			if manufCount == 0 {
				// Create manufacturer if it doesn't exist
				_, err := manufacturersCol.InsertOne(ctx, bson.M{
					"_id": eff.Manufacturer,
				})
				if err != nil {
					log.Printf("Warning: failed to create manufacturer %q: %v", eff.Manufacturer, err)
					// Continue anyway
				}
			}
		}

		// Handle image file
		if imageFile != "" && !dryRun {
			srcImage := filepath.Join(effectDir, imageFile)
			fileData, err := os.ReadFile(srcImage)
			if err != nil {
				log.Printf("Warning: failed to read image %s: %v", imageFile, err)
				continue
			}

			mimeType := mimeTypeForFile(imageFile)
			containerInfo, err := blobSvc.Save(fileData, mimeType)
			if err != nil {
				log.Printf("Warning: failed to save image to blob store: %v", err)
				continue
			}

			eff.Images = []*domain.ContainerInfo{containerInfo}
		}

		// Insert into MongoDB
		if !dryRun {
			doc := bson.M{
				"_id":            eff.ID,
				"createdAt":      eff.CreatedAt,
				"lastModifiedAt": eff.LastModifiedAt,
				"effectType":     eff.EffectType,
				"manufacturer":   eff.Manufacturer,
				"model":          eff.Model,
				"tags":           eff.Tags,
				"comment":        eff.Comment,
				"images":         eff.Images,
				"connector":      eff.Connector,
				"voltage":        eff.Voltage,
				"current":        eff.Current,
			}

			if _, err := effectsCol.InsertOne(ctx, doc); err != nil {
				log.Printf("Warning: failed to insert effect %q: %v", eff.ID, err)
				continue
			}
		}

		count++
	}

	return count, nil
}

// parseEffectData reads and parses an effect from a directory
// Returns the parsed Effect, the image filename, and any error
func parseEffectData(dir string) (*domain.Effect, string, error) {
	jsonFile := filepath.Join(dir, "effect.json")
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return nil, "", fmt.Errorf("read effect.json: %w", err)
	}

	var backup backupEffect
	if err := json.Unmarshal(data, &backup); err != nil {
		return nil, "", fmt.Errorf("unmarshal effect.json: %w", err)
	}

	if strings.TrimSpace(backup.ID) == "" {
		return nil, "", fmt.Errorf("effect has empty id")
	}

	// Parse timestamps
	createdAt, err := time.Parse(time.RFC3339, backup.CreatedAt)
	if err != nil {
		createdAt = time.Now().UTC()
	}

	lastModifiedAt, err := time.Parse(time.RFC3339, backup.LastModifiedAt)
	if err != nil {
		lastModifiedAt = time.Now().UTC()
	}

	eff := &domain.Effect{
		ID:             backup.ID,
		CreatedAt:      createdAt,
		LastModifiedAt: lastModifiedAt,
		EffectType:     strings.TrimSpace(backup.EffectType),
		Manufacturer:   strings.TrimSpace(backup.Manufacturer),
		Model:          strings.TrimSpace(backup.Model),
		Tags:           backup.Tags,
		Comment:        strings.TrimSpace(backup.Comment),
		Connector:      strings.TrimSpace(backup.Connector),
		Voltage:        strings.TrimSpace(backup.Voltage),
		Current:        strings.TrimSpace(backup.Current),
	}

	return eff, backup.Image, nil
}
