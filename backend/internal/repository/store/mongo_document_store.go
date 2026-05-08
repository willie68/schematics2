package store

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/samber/do/v2"
	"github.com/willie68/schematic2/backend/internal/config"
	"github.com/willie68/schematic2/backend/internal/domain"
	"github.com/willie68/schematic2/backend/internal/logging"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const documentsCollection = "documents"

// MongoDocumentStore stores domain.Document in MongoDB.
type MongoDocumentStore struct {
	cfg    config.MongoDB
	client *mongo.Client
	db     *mongo.Database
	col    *mongo.Collection
	logger *slog.Logger
}

type mongoDocument struct {
	ID    string   `bson:"_id"`
	Title string   `bson:"title"`
	Path  string   `bson:"path"`
	Tags  []string `bson:"tags"`
	Text  string   `bson:"text"`
}

func NewMongoDocumentStore(inj do.Injector) *MongoDocumentStore {
	cfg := do.MustInvoke[config.Config](inj)
	return &MongoDocumentStore{
		cfg:    cfg.MongoDB,
		logger: logging.New("mongo-document-store"),
	}
}

func (s *MongoDocumentStore) Prepare() error {
	hosts := s.cfg.GetHosts()
	if len(hosts) == 0 {
		return errors.New("no mongo hosts provided")
	}
	if s.cfg.Database == "" {
		return errors.New("mongo database is empty")
	}

	uri := buildMongoURI(s.cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return err
	}

	if err = client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return err
	}

	s.client = client
	s.db = client.Database(s.cfg.Database)
	s.col = s.db.Collection(documentsCollection)

	if err = s.ensureIndexes(); err != nil {
		_ = client.Disconnect(context.Background())
		return err
	}

	s.logger.Info("connected to mongodb", "hosts", strings.Join(hosts, ","), "database", s.cfg.Database)
	return nil
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

func (s *MongoDocumentStore) ensureIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "path", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "tags", Value: 1}}},
		{Keys: bson.D{{Key: "title", Value: "text"}, {Key: "text", Value: "text"}}},
	}

	_, err := s.col.Indexes().CreateMany(ctx, indexes)
	return err
}

func (s *MongoDocumentStore) Upsert(doc domain.Document) error {
	if s.col == nil {
		return errors.New("mongodb not initialised")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	payload := mongoDocument{
		ID:    doc.ID,
		Title: doc.Title,
		Path:  doc.Path,
		Tags:  doc.Tags,
		Text:  doc.Text,
	}

	_, err := s.col.UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: doc.ID}},
		bson.D{{Key: "$set", Value: payload}},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("upsert document: %w", err)
	}

	return nil
}

func (s *MongoDocumentStore) Get(id string) (domain.Document, bool) {
	if s.col == nil {
		return domain.Document{}, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var doc mongoDocument
	err := s.col.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.Document{}, false
	}
	if err != nil {
		s.logger.Error("get document failed", "error", err, "id", id)
		return domain.Document{}, false
	}

	return domain.Document{ID: doc.ID, Title: doc.Title, Path: doc.Path, Tags: doc.Tags, Text: doc.Text}, true
}

func (s *MongoDocumentStore) List() []domain.Document {
	if s.col == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cur, err := s.col.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		s.logger.Error("list documents failed", "error", err)
		return nil
	}
	defer cur.Close(ctx)

	var docs []mongoDocument
	if err = cur.All(ctx, &docs); err != nil {
		s.logger.Error("decode document list failed", "error", err)
		return nil
	}

	out := make([]domain.Document, 0, len(docs))
	for _, d := range docs {
		out = append(out, domain.Document{ID: d.ID, Title: d.Title, Path: d.Path, Tags: d.Tags, Text: d.Text})
	}

	return out
}

func (s *MongoDocumentStore) Close(ctx context.Context) error {
	if s.client == nil {
		return nil
	}
	return s.client.Disconnect(ctx)
}
