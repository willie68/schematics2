package store

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
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
const tagsCollection = "tags"
const manufacturersCollection = "manufacturers"

// mongoDocumentStore stores domain.Document in MongoDB.
type mongoDocumentStore struct {
	cfg       config.MongoDB
	client    *mongo.Client
	db        *mongo.Database
	col       *mongo.Collection
	tagsCol   *mongo.Collection
	manufCol  *mongo.Collection
	logger    *slog.Logger
}

type mongoDocument struct {
	ID             string                `bson:"_id"`
	CreatedAt      time.Time             `bson:"createdAt"`
	LastModifiedAt time.Time             `bson:"lastModifiedAt"`
	Manufacturer   string                `bson:"manufacturer"`
	Model          string                `bson:"model"`
	Subtitle       string                `bson:"subtitle"`
	Tags           []string              `bson:"tags"`
	Description    string                `bson:"description"`
	PrivateFile    bool                  `bson:"privateFile"`
	Owner          string                `bson:"owner"`
	Files          []domain.DocumentFile `bson:"files"`
}

type mongoTag struct {
	Tag     string `bson:"_id"`
	Counter int64  `bson:"counter"`
}

func NewMongoDocumentStore(inj do.Injector) *mongoDocumentStore {
	cfg := do.MustInvoke[config.Config](inj)
	return &mongoDocumentStore{
		cfg:    cfg.MongoDB,
		logger: logging.New("mongo-document-store"),
	}
}

func (s *mongoDocumentStore) Prepare() error {
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
	if s.cfg.DirectConnection {
		clientOpts.SetDirect(true)
	}
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
	s.tagsCol = s.db.Collection(tagsCollection)
	s.manufCol = s.db.Collection(manufacturersCollection)

	if err = s.ensureIndexes(); err != nil {
		_ = client.Disconnect(context.Background())
		return err
	}
	if err = s.rebuildTagCounters(); err != nil {
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

func (s *mongoDocumentStore) ensureIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "tags", Value: 1}}},
		{Keys: bson.D{{Key: "manufacturer", Value: "text"}, {Key: "model", Value: "text"}, {Key: "subtitle", Value: "text"}, {Key: "description", Value: "text"}, {Key: "files.name", Value: "text"}, {Key: "files.type", Value: "text"}}},
	}

	_, err := s.col.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}

	tagIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "counter", Value: -1}}},
	}

	_, err = s.tagsCol.Indexes().CreateMany(ctx, tagIndexes)
	if err != nil {
		return err
	}

	manufIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "_id", Value: 1}}},
	}

	_, err = s.manufCol.Indexes().CreateMany(ctx, manufIndexes)
	return err
}

func (s *mongoDocumentStore) Upsert(doc domain.Document) error {
	if s.col == nil {
		return errors.New("mongodb not initialised")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC()
	createdAt := doc.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}

	oldTags, err := s.fetchDocumentTags(ctx, doc.ID)
	if err != nil {
		return err
	}

	payload := mongoDocument{
		ID:             doc.ID,
		CreatedAt:      createdAt,
		LastModifiedAt: now,
		Manufacturer:   doc.Manufacturer,
		Model:          doc.Model,
		Subtitle:       doc.Subtitle,
		Tags:           normalizeTags(doc.Tags),
		Description:    doc.Description,
		PrivateFile:    doc.PrivateFile,
		Owner:          doc.Owner,
		Files:          doc.Files,
	}

	_, err = s.col.UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: doc.ID}},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "lastModifiedAt", Value: payload.LastModifiedAt},
				{Key: "manufacturer", Value: payload.Manufacturer},
				{Key: "model", Value: payload.Model},
				{Key: "subtitle", Value: payload.Subtitle},
				{Key: "tags", Value: payload.Tags},
				{Key: "description", Value: payload.Description},
				{Key: "privateFile", Value: payload.PrivateFile},
				{Key: "owner", Value: payload.Owner},
				{Key: "files", Value: payload.Files},
			}},
			{Key: "$setOnInsert", Value: bson.D{{Key: "createdAt", Value: payload.CreatedAt}}},
		},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("upsert document: %w", err)
	}
	if err = s.updateTagCounters(ctx, oldTags, payload.Tags); err != nil {
		return err
	}
	if err = s.updateManufacturer(ctx, payload.Manufacturer); err != nil {
		return err
	}

	return nil
}

func (s *mongoDocumentStore) fetchDocumentTags(ctx context.Context, id string) ([]string, error) {
	if s.col == nil {
		return nil, errors.New("mongodb not initialised")
	}

	var prev struct {
		Tags []string `bson:"tags"`
	}
	err := s.col.FindOne(ctx, bson.D{{Key: "_id", Value: id}}, options.FindOne().SetProjection(bson.D{{Key: "tags", Value: 1}})).Decode(&prev)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read existing tags: %w", err)
	}

	return normalizeTags(prev.Tags), nil
}

func (s *mongoDocumentStore) updateTagCounters(ctx context.Context, oldTags []string, newTags []string) error {
	if s.tagsCol == nil {
		return errors.New("mongodb tags collection not initialised")
	}

	oldSet := toTagSet(oldTags)
	newSet := toTagSet(newTags)

	for tag := range newSet {
		if _, exists := oldSet[tag]; exists {
			continue
		}
		_, err := s.tagsCol.UpdateOne(
			ctx,
			bson.D{{Key: "_id", Value: tag}},
			bson.D{{Key: "$inc", Value: bson.D{{Key: "counter", Value: 1}}}},
			options.UpdateOne().SetUpsert(true),
		)
		if err != nil {
			return fmt.Errorf("increment tag counter for %q: %w", tag, err)
		}
	}

	for tag := range oldSet {
		if _, exists := newSet[tag]; exists {
			continue
		}
		_, err := s.tagsCol.UpdateOne(
			ctx,
			bson.D{{Key: "_id", Value: tag}},
			bson.D{{Key: "$inc", Value: bson.D{{Key: "counter", Value: -1}}}},
		)
		if err != nil {
			return fmt.Errorf("decrement tag counter for %q: %w", tag, err)
		}
	}

	if _, err := s.tagsCol.DeleteMany(ctx, bson.D{{Key: "counter", Value: bson.D{{Key: "$lte", Value: 0}}}}); err != nil {
		return fmt.Errorf("cleanup tag counters: %w", err)
	}

	return nil
}

func (s *mongoDocumentStore) rebuildTagCounters() error {
	if s.col == nil || s.tagsCol == nil {
		return errors.New("mongodb not initialised")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cur, err := s.col.Find(ctx, bson.D{}, options.Find().SetProjection(bson.D{{Key: "tags", Value: 1}}))
	if err != nil {
		return fmt.Errorf("load document tags for rebuild: %w", err)
	}
	defer cur.Close(ctx)

	var docs []struct {
		Tags []string `bson:"tags"`
	}
	if err = cur.All(ctx, &docs); err != nil {
		return fmt.Errorf("decode document tags for rebuild: %w", err)
	}

	counters := make(map[string]int64)
	for _, d := range docs {
		for tag := range toTagSet(d.Tags) {
			counters[tag]++
		}
	}

	if _, err = s.tagsCol.DeleteMany(ctx, bson.D{}); err != nil {
		return fmt.Errorf("clear tag counters before rebuild: %w", err)
	}

	if len(counters) == 0 {
		return nil
	}

	docsToInsert := make([]any, 0, len(counters))
	for tag, counter := range counters {
		docsToInsert = append(docsToInsert, mongoTag{Tag: tag, Counter: counter})
	}

	if _, err = s.tagsCol.InsertMany(ctx, docsToInsert); err != nil {
		return fmt.Errorf("insert rebuilt tag counters: %w", err)
	}

	return nil
}

func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}

	set := toTagSet(tags)
	if len(set) == 0 {
		return nil
	}

	out := make([]string, 0, len(set))
	for tag := range set {
		out = append(out, tag)
	}
	return out
}

func toTagSet(tags []string) map[string]struct{} {
	set := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		t := strings.ToLower(strings.TrimSpace(tag))
		if t == "" {
			continue
		}
		set[t] = struct{}{}
	}
	return set
}

func (s *mongoDocumentStore) Get(id string) (domain.Document, bool) {
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

	return domain.Document{
		ID:             doc.ID,
		CreatedAt:      doc.CreatedAt,
		LastModifiedAt: doc.LastModifiedAt,
		Manufacturer:   doc.Manufacturer,
		Model:          doc.Model,
		Subtitle:       doc.Subtitle,
		Tags:           doc.Tags,
		Description:    doc.Description,
		PrivateFile:    doc.PrivateFile,
		Owner:          doc.Owner,
		Files:          doc.Files,
	}, true
}

func (s *mongoDocumentStore) List() []domain.Document {
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
		out = append(out, domain.Document{
			ID:             d.ID,
			CreatedAt:      d.CreatedAt,
			LastModifiedAt: d.LastModifiedAt,
			Manufacturer:   d.Manufacturer,
			Model:          d.Model,
			Subtitle:       d.Subtitle,
			Tags:           d.Tags,
			Description:    d.Description,
			PrivateFile:    d.PrivateFile,
			Owner:          d.Owner,
			Files:          d.Files,
		})
	}

	return out
}

// Search executes a full-text and tag-based search using MongoDB queries
func (s *mongoDocumentStore) Search(filter domain.SearchFilter) []domain.SearchResult {
	if s.col == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build MongoDB filter
	var mongoFilter bson.D

	if filter.Query == "" && len(filter.Tags) == 0 {
		// No filters - return all documents
		mongoFilter = bson.D{}
	} else if filter.Query == "" && len(filter.Tags) > 0 {
		// Tag-only search: all tags must match
		mongoFilter = bson.D{
			{Key: "tags", Value: bson.D{
				{Key: "$all", Value: filter.Tags},
			}},
		}
	} else if filter.Query != "" && len(filter.Tags) == 0 {
		// Full-text search only
		mongoFilter = bson.D{
			{Key: "$text", Value: bson.D{
				{Key: "$search", Value: filter.Query},
			}},
		}
	} else {
		// Both query and tags: combine with AND
		mongoFilter = bson.D{
			{Key: "$and", Value: []bson.D{
				{
					{Key: "$text", Value: bson.D{
						{Key: "$search", Value: filter.Query},
					}},
				},
				{
					{Key: "tags", Value: bson.D{
						{Key: "$all", Value: filter.Tags},
					}},
				},
			}},
		}
	}

	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}})
	cur, err := s.col.Find(ctx, mongoFilter, opts)
	if err != nil {
		s.logger.Error("search documents failed", "error", err)
		return nil
	}
	defer cur.Close(ctx)

	var docs []mongoDocument
	if err = cur.All(ctx, &docs); err != nil {
		s.logger.Error("decode search results failed", "error", err)
		return nil
	}

	out := make([]domain.SearchResult, 0, len(docs))
	for _, d := range docs {
		out = append(out, domain.SearchResult{
			Document: domain.Document{
				ID:             d.ID,
				CreatedAt:      d.CreatedAt,
				LastModifiedAt: d.LastModifiedAt,
				Manufacturer:   d.Manufacturer,
				Model:          d.Model,
				Subtitle:       d.Subtitle,
				Tags:           d.Tags,
				Description:    d.Description,
				PrivateFile:    d.PrivateFile,
				Owner:          d.Owner,
				Files:          d.Files,
			},
		})
	}

	return out
}

func (s *mongoDocumentStore) ListTags(ctx context.Context) ([]domain.Tag, error) {
	if s.tagsCol == nil {
		return nil, errors.New("mongodb tags collection not initialised")
	}

	opts := options.Find().SetSort(bson.D{{Key: "counter", Value: -1}})
	cur, err := s.tagsCol.Find(ctx, bson.D{}, opts)
	if err != nil {
		s.logger.Error("list tags failed", "error", err)
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer cur.Close(ctx)

	var mongoTags []mongoTag
	if err = cur.All(ctx, &mongoTags); err != nil {
		s.logger.Error("decode tag list failed", "error", err)
		return nil, fmt.Errorf("decode tags: %w", err)
	}

	out := make([]domain.Tag, 0, len(mongoTags))
	for _, t := range mongoTags {
		out = append(out, domain.Tag{Name: t.Tag, Counter: t.Counter})
	}

	return out, nil
}

func (s *mongoDocumentStore) SuggestTags(ctx context.Context, prefix string, limit int) ([]domain.Tag, error) {
	if s.tagsCol == nil {
		return nil, errors.New("mongodb tags collection not initialised")
	}

	if limit <= 0 {
		limit = 10
	}

	prefix = strings.ToLower(strings.TrimSpace(prefix))
	if prefix == "" {
		opts := options.Find().SetSort(bson.D{{Key: "counter", Value: -1}}).SetLimit(int64(limit))
		cur, err := s.tagsCol.Find(ctx, bson.D{}, opts)
		if err != nil {
			return nil, fmt.Errorf("suggest tags: %w", err)
		}
		defer cur.Close(ctx)

		var mongoTags []mongoTag
		if err = cur.All(ctx, &mongoTags); err != nil {
			return nil, fmt.Errorf("decode tags: %w", err)
		}

		out := make([]domain.Tag, 0, len(mongoTags))
		for _, t := range mongoTags {
			out = append(out, domain.Tag{Name: t.Tag, Counter: t.Counter})
		}
		return out, nil
	}

	filter := bson.D{{Key: "_id", Value: bson.D{{Key: "$regex", Value: "^" + prefix}}}}
	opts := options.Find().SetSort(bson.D{{Key: "counter", Value: -1}}).SetLimit(int64(limit))
	cur, err := s.tagsCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("suggest tags: %w", err)
	}
	defer cur.Close(ctx)

	var mongoTags []mongoTag
	if err = cur.All(ctx, &mongoTags); err != nil {
		return nil, fmt.Errorf("decode tags: %w", err)
	}

	out := make([]domain.Tag, 0, len(mongoTags))
	for _, t := range mongoTags {
		out = append(out, domain.Tag{Name: t.Tag, Counter: t.Counter})
	}

	return out, nil
}

func (s *mongoDocumentStore) SuggestManufacturers(ctx context.Context, prefix string, limit int) ([]string, error) {
	if s.manufCol == nil {
		return nil, errors.New("mongodb manufacturers collection not initialised")
	}

	if limit <= 0 {
		limit = 10
	}

	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}).SetLimit(int64(limit))
		cur, err := s.manufCol.Find(ctx, bson.D{}, opts)
		if err != nil {
			return nil, fmt.Errorf("suggest manufacturers: %w", err)
		}
		defer cur.Close(ctx)

		var docs []bson.M
		if err = cur.All(ctx, &docs); err != nil {
			return nil, fmt.Errorf("decode manufacturers: %w", err)
		}

		out := make([]string, 0, len(docs))
		for _, doc := range docs {
			if id, ok := doc["_id"].(string); ok {
				out = append(out, id)
			}
		}
		return out, nil
	}

	// Case-sensitive regex for manufacturers (unlike tags which are case-insensitive)
	filter := bson.D{{Key: "_id", Value: bson.D{{Key: "$regex", Value: "^" + regexp.QuoteMeta(prefix)}}}}
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}).SetLimit(int64(limit))
	cur, err := s.manufCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("suggest manufacturers: %w", err)
	}
	defer cur.Close(ctx)

	var docs []bson.M
	if err = cur.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode manufacturers: %w", err)
	}

	out := make([]string, 0, len(docs))
	for _, doc := range docs {
		if id, ok := doc["_id"].(string); ok {
			out = append(out, id)
		}
	}

	return out, nil
}

func (s *mongoDocumentStore) updateManufacturer(ctx context.Context, manufacturer string) error {
	if s.manufCol == nil || manufacturer == "" {
		return nil
	}

	// Just ensure the manufacturer exists in the collection (upsert)
	_, err := s.manufCol.UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: manufacturer}},
		bson.D{{Key: "$setOnInsert", Value: bson.D{{Key: "_id", Value: manufacturer}}}},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("update manufacturer: %w", err)
	}

	return nil
}

func (s *mongoDocumentStore) Close(ctx context.Context) error {
	if s.client == nil {
		return nil
	}
	return s.client.Disconnect(ctx)
}
