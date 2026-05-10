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

type Manufacturer struct {
	Name  string `json:"name" bson:"name"`
	Count int    `json:"count" bson:"count"`
}

func main() {
	manufacturersDir := flag.String("manufacturers-dir", "testdata/manufacturers", "Directory containing manufacturer JSON files")
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
	manufCol := db.Collection("manufacturers")

	// Import manufacturers
	imported, err := importManufacturers(ctx, manufCol, *manufacturersDir)
	if err != nil {
		log.Fatalf("Failed to import manufacturers: %v", err)
	}

	fmt.Printf("Imported manufacturers from %d JSON files\n", countFiles(*manufacturersDir))
	fmt.Printf("✓ Successfully imported %d manufacturers\n", imported)
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

func countFiles(dir string) int {
	files, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			count++
		}
	}
	return count
}

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

		// Trim and preserve case (unlike tags which are normalized to lowercase)
		name := strings.TrimSpace(manufacturer.Name)
		if name == "" {
			continue
		}

		// Create BSON document with name as _id (case-preserved)
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
