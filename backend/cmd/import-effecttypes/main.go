// Package main provides a tool for importing effect types from the backup format.
//
// Directory structure (under base-dir):
//   - 5e8485f337ce217b37808267/
//   - effecttype.json       – effect type metadata
//   - <image-file>          – type image (copied to internal/repository/effecttypes)
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/willie68/schematic2/backend/internal/config"
	"github.com/willie68/schematic2/backend/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// backupEffectType is the JSON structure of the old server backup
type backupEffectType struct {
	ID             string            `json:"id"`
	ForeignID      string            `json:"foreignId"`
	CreatedAt      string            `json:"createdAt"`
	LastModifiedAt string            `json:"lastModifiedAt"`
	TypeName       string            `json:"typeName"`
	Nls            map[string]string `json:"nls"`
	TypeImage      string            `json:"typeImage"`
}

func main() {
	effectTypesDir := flag.String("effecttypes-dir", "testdata/effecttypes", "Directory containing effect type subdirectories")
	imagesOutDir := flag.String("images-out-dir", "internal/repository/effecttypes", "Output directory for effect type images")
	dryRun := flag.Bool("dry-run", false, "Process and validate data but do not write to MongoDB or copy images")
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
	effectTypesCol := db.Collection("effecttypes")

	// Import effect types
	imported, err := importEffectTypes(ctx, effectTypesCol, *effectTypesDir, *imagesOutDir, *dryRun)
	if err != nil {
		log.Fatalf("Failed to import effect types: %v", err)
	}

	fmt.Printf("✓ Successfully imported %d effect types\n", imported)
	if !*dryRun {
		fmt.Printf("✓ Images copied to %s\n", *imagesOutDir)
	}
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
