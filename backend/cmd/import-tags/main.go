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

	"github.com/willie68/schematic2/backend/internal/config"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Tag struct {
	Name  string `json:"name" bson:"name"`
	Count int    `json:"count" bson:"count"`
}

func main() {
	tagsDir := flag.String("tags-dir", "testdata/tags", "Directory containing tag JSON files")
	flag.Parse()

	// Load config
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
	tagsCol := db.Collection("tags")

	// Import tags
	imported, err := importTags(ctx, tagsCol, *tagsDir)
	if err != nil {
		log.Fatalf("Failed to import tags: %v", err)
	}

	fmt.Printf("✓ Successfully imported %d tags\n", imported)
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

func importTags(ctx context.Context, tagsCol *mongo.Collection, tagsDir string) (int, error) {
	// Clear existing tags first
	_, err := tagsCol.DeleteMany(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("failed to clear existing tags: %w", err)
	}

	// Read all JSON files from directory
	entries, err := os.ReadDir(tagsDir)
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

		filePath := filepath.Join(tagsDir, entry.Name())
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
	result, err := tagsCol.InsertMany(ctx, tags)
	if err != nil {
		return 0, fmt.Errorf("failed to insert tags: %w", err)
	}

	fmt.Printf("Imported tags from %d JSON files\n", len(result.InsertedIDs))
	return len(result.InsertedIDs), nil
}
