// Package main provides a tool for importing effects from the backup format.
//
// Directory structure (under base-dir):
//   - 5e8485f337ce217b378082a4/
//   - effect.json           – effect metadata
//   - <image-file>          – effect image (stored in blob repo)
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"regexp"

	"github.com/samber/do/v2"
	"github.com/willie68/schematic2/backend/internal/config"
	"github.com/willie68/schematic2/backend/internal/domain"
	"github.com/willie68/schematic2/backend/internal/services/blob"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// backupEffect is the JSON structure of the old server backup
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
	effectsDir := flag.String("effects-dir", "testdata/effects", "Directory containing effect subdirectories")
	dryRun := flag.Bool("dry-run", false, "Process and validate data but do not write to MongoDB or blob store")
	flag.Parse()

	cfg := config.LoadFromEnv()

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

	// Set up DI container for blob store
	inj := do.New()
	do.ProvideValue(inj, cfg)

	var blobSvc *blob.Service
	if !*dryRun {
		blobSvc = blob.New(inj)
		if err := blobSvc.Prepare(); err != nil {
			log.Fatalf("prepare blob store: %v", err)
		}
		do.ProvideValue(inj, blobSvc)
	}

	// Import effects
	imported, err := importEffects(ctx, db, *effectsDir, blobSvc, *dryRun)
	if err != nil {
		log.Fatalf("Failed to import effects: %v", err)
	}

	fmt.Printf("✓ Successfully imported %d effects\n", imported)
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

func importEffects(ctx context.Context, db *mongo.Database, dir string, blobSvc *blob.Service, dryRun bool) (int, error) {
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
		eff, imageFile, err := parseEffect(effectDir)
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
			count, _ := effectTypesCol.CountDocuments(ctx, filter)
			if count == 0 {
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

// parseEffect reads and parses an effect from a directory
// Returns the parsed Effect, the image filename, and any error
func parseEffect(dir string) (*domain.Effect, string, error) {
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

// mimeTypeForFile returns the MIME type based on the file extension
func mimeTypeForFile(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
