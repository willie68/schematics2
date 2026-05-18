package store

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/samber/do/v2"
	"github.com/willie68/schematics2/backend/internal/config"
	"github.com/willie68/schematics2/backend/internal/domain/model"
	"github.com/willie68/schematics2/backend/internal/logging"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const documentsCollection = "documents"
const tagsCollection = "tags"
const manufacturersCollection = "manufacturers"
const usersCollection = "users"

// MongoStore stores model.Document in MongoDB.
type MongoStore struct {
	cfg            config.MongoDB
	client         *mongo.Client
	db             *mongo.Database
	col            *mongo.Collection
	tagsCol        *mongo.Collection
	manufCol       *mongo.Collection
	usersCol       *mongo.Collection
	effectsCol     *mongo.Collection
	effectTypesCol *mongo.Collection
	logger         *slog.Logger
}

type mongoDocument struct {
	ID             string               `bson:"_id"`
	CreatedAt      time.Time            `bson:"createdAt"`
	LastModifiedAt time.Time            `bson:"lastModifiedAt"`
	Manufacturer   string               `bson:"manufacturer"`
	Model          string               `bson:"model"`
	Subtitle       string               `bson:"subtitle"`
	Tags           []string             `bson:"tags"`
	Description    string               `bson:"description"`
	PrivateFile    bool                 `bson:"privateFile"`
	Owner          string               `bson:"owner"`
	Files          []model.DocumentFile `bson:"files"`
}

type mongoTag struct {
	Tag     string `bson:"_id"`
	Counter int64  `bson:"counter"`
}

func NewMongoStore(inj do.Injector) *MongoStore {
	cfg := do.MustInvoke[config.Config](inj)
	return &MongoStore{
		cfg:    cfg.MongoDB,
		logger: logging.New("mongo-store"),
	}
}

func (s *MongoStore) Prepare() error {
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
	s.usersCol = s.db.Collection(usersCollection)
	s.effectsCol = s.db.Collection(effectsCollection)
	s.effectTypesCol = s.db.Collection(effectTypesCollection)

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

func (s *MongoStore) ensureIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Indices for documents collection
	// Using B-tree indices for regex-based search (not text search)
	indexes := []mongo.IndexModel{
		// Search fields - used for prefix matching and term search with $regex
		{Keys: bson.D{{Key: "manufacturer", Value: 1}}},
		{Keys: bson.D{{Key: "model", Value: 1}}},
		{Keys: bson.D{{Key: "subtitle", Value: 1}}},
		{Keys: bson.D{{Key: "description", Value: 1}}},
		{Keys: bson.D{{Key: "tags", Value: 1}}},

		// Access control fields
		{Keys: bson.D{{Key: "privateFile", Value: 1}}},
		{Keys: bson.D{{Key: "owner", Value: 1}}},

		// Compound index for private file + owner queries
		{Keys: bson.D{{Key: "privateFile", Value: 1}, {Key: "owner", Value: 1}}},
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
	if err != nil {
		return err
	}

	userIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err = s.usersCol.Indexes().CreateMany(ctx, userIndexes)
	return err
}

func (s *MongoStore) Upsert(doc model.Document) error {
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
	if err = s.UpdateManufacturer(ctx, payload.Manufacturer); err != nil {
		return err
	}

	return nil
}

func (s *MongoStore) fetchDocumentTags(ctx context.Context, id string) ([]string, error) {
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

func (s *MongoStore) GetByID(ctx context.Context, id string) (model.Document, error) {
	if s.col == nil {
		return model.Document{}, errors.New("mongodb not initialised")
	}

	var mdoc mongoDocument
	err := s.col.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&mdoc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return model.Document{}, fmt.Errorf("document not found")
	}
	if err != nil {
		return model.Document{}, fmt.Errorf("find document: %w", err)
	}

	return model.Document{
		ID:             mdoc.ID,
		CreatedAt:      mdoc.CreatedAt,
		LastModifiedAt: mdoc.LastModifiedAt,
		Manufacturer:   mdoc.Manufacturer,
		Model:          mdoc.Model,
		Subtitle:       mdoc.Subtitle,
		Tags:           mdoc.Tags,
		Description:    mdoc.Description,
		PrivateFile:    mdoc.PrivateFile,
		Owner:          mdoc.Owner,
		Files:          mdoc.Files,
	}, nil
}

// Exists reports whether a document with the given ID already exists in the store.
func (s *MongoStore) Exists(ctx context.Context, id string) (bool, error) {
	if s.col == nil {
		return false, errors.New("mongodb not initialised")
	}

	count, err := s.col.CountDocuments(ctx, bson.D{{Key: "_id", Value: id}}, options.Count().SetLimit(1))
	if err != nil {
		return false, fmt.Errorf("count document: %w", err)
	}
	return count > 0, nil
}

// DeleteByID deletes a document from the store by its ID
func (s *MongoStore) DeleteByID(ctx context.Context, id string) error {
	if s.col == nil {
		return errors.New("mongodb not initialised")
	}

	result, err := s.col.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("document not found")
	}

	return nil
}

func (s *MongoStore) updateTagCounters(ctx context.Context, oldTags []string, newTags []string) error {
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

func (s *MongoStore) rebuildTagCounters() error {
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

func (s *MongoStore) Get(id string) (model.Document, bool) {
	if s.col == nil {
		return model.Document{}, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var doc mongoDocument
	err := s.col.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return model.Document{}, false
	}
	if err != nil {
		s.logger.Error("get document failed", "error", err, "id", id)
		return model.Document{}, false
	}

	return model.Document{
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

func (s *MongoStore) List() []model.Document {
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

	out := make([]model.Document, 0, len(docs))
	for _, d := range docs {
		out = append(out, model.Document{
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

// Search executes a search using parsed query terms with regex and tag-based filtering.
// Supports prefix matching, AND/OR/NOT operators, and tag filters.
func (s *MongoStore) SearchStore(filter model.Query) model.PagedSearchResult {
	if s.col == nil {
		return model.PagedSearchResult{Results: []model.SearchResult{}}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build MongoDB filter from parsed query terms
	mongoFilterM := s.buildMongoFilterFromParsedQuery(filter)

	// Count total matching documents
	total, err := s.col.CountDocuments(ctx, mongoFilterM)
	if err != nil {
		s.logger.Error("count search results failed", "error", err)
		return model.PagedSearchResult{Results: []model.SearchResult{}}
	}

	sortField := "_id"
	if filter.Sort.Field != "" {
		sortField = filter.Sort.Field
	}
	sortOrder := int32(1)
	if filter.Sort.Order == -1 {
		sortOrder = -1
	}
	opts := options.Find().SetSort(bson.D{{Key: sortField, Value: sortOrder}})
	if filter.Pagination.Skip > 0 {
		opts.SetSkip(filter.Pagination.Skip)
	}
	if filter.Pagination.Limit > 0 {
		opts.SetLimit(filter.Pagination.Limit)
	}

	cur, err := s.col.Find(ctx, mongoFilterM, opts)
	if err != nil {
		s.logger.Error("search documents failed", "error", err)
		return model.PagedSearchResult{Results: []model.SearchResult{}}
	}
	defer cur.Close(ctx)

	var docs []mongoDocument
	if err = cur.All(ctx, &docs); err != nil {
		s.logger.Error("decode search results failed", "error", err)
		return model.PagedSearchResult{Results: []model.SearchResult{}}
	}

	out := make([]model.SearchResult, 0, len(docs))
	for _, d := range docs {
		out = append(out, model.SearchResult{
			Document: model.Document{
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

	return model.PagedSearchResult{
		Results:    out,
		Total:      total,
		Pagination: filter.Pagination,
	}
}

// buildMongoFilterFromParsedQuery constructs a MongoDB filter from the parsed query.
// It combines required terms (AND), optional terms (OR), excluded terms (NOT), and tag filters.
func (s *MongoStore) buildMongoFilterFromParsedQuery(filter model.Query) bson.M {
	var conditions []bson.M

	parsed := filter.ParsedQuery

	// Build filter for required terms (ALL must match - AND logic)
	if len(parsed.Required) > 0 {
		for _, term := range parsed.Required {
			conditions = append(conditions, s.buildRegexFilterForTerm(term))
		}
	}

	// Build filter for optional terms (AT LEAST ONE must match - OR logic)
	if len(parsed.Optional) > 0 {
		var optionalConditions []bson.M
		for _, term := range parsed.Optional {
			optionalConditions = append(optionalConditions, s.buildRegexFilterForTerm(term))
		}
		if len(optionalConditions) > 0 {
			conditions = append(conditions, bson.M{"$or": optionalConditions})
		}
	}

	// Build filter for excluded terms (NONE must match - NOT logic)
	if len(parsed.Excluded) > 0 {
		for _, term := range parsed.Excluded {
			exclusionFilter := s.buildRegexFilterForTerm(term)
			// Negate the filter using $nor
			conditions = append(conditions, bson.M{
				"$nor": []bson.M{exclusionFilter},
			})
		}
	}

	// Build filter for tag filters (ALL tags must be present - AND logic)
	if len(parsed.TagFilters) > 0 {
		conditions = append(conditions, bson.M{
			"tags": bson.M{"$all": parsed.TagFilters},
		})
	}

	// Add privateFile filter based on authentication and privateOnly flag
	privateFileFilter := s.buildPrivateFileFilter(filter.IsAuthenticated, filter.PrivateOnly, filter.Username)
	if len(privateFileFilter) > 0 {
		conditions = append(conditions, privateFileFilter)
	}

	// Combine all conditions with $and
	if len(conditions) == 0 {
		return bson.M{}
	}
	if len(conditions) == 1 {
		return conditions[0]
	}
	return bson.M{"$and": conditions}
}

// buildRegexFilterForTerm builds a regex filter that matches the term in any searchable field.
// Searches in: manufacturer, model, subtitle, description, and tags.
// Prefix match if term.IsPrefix is true (uses ^), otherwise full-word match.
func (s *MongoStore) buildRegexFilterForTerm(term model.QueryTerm) bson.M {
	pattern := term.Value
	if term.IsPrefix {
		pattern = "^" + term.Value
	}

	regexFilter := bson.M{"$regex": pattern, "$options": "i"} // case-insensitive

	// Match in any of these fields (OR logic)
	return bson.M{
		"$or": []bson.M{
			{"manufacturer": regexFilter},
			{"model": regexFilter},
			{"subtitle": regexFilter},
			{"description": regexFilter},
			{"tags": regexFilter},
		},
	}
}

// buildPrivateFileFilter constructs the filter for private/public document access.
// Guests: only public documents (privateFile != true)
// Authenticated with privateOnly=false: public documents + own private documents
// Authenticated with privateOnly=true: only own private documents
func (s *MongoStore) buildPrivateFileFilter(isAuthenticated, privateOnly bool, username string) bson.M {
	if !isAuthenticated {
		// Guests: only public documents
		return bson.M{"privateFile": bson.M{"$ne": true}}
	}
	if privateOnly {
		// Authenticated with private filter ON: only own private documents
		return bson.M{
			"$and": []bson.M{
				{"privateFile": true},
				{"owner": username},
			},
		}
	}
	// Authenticated with private filter OFF: public documents + own private documents
	return bson.M{
		"$or": []bson.M{
			{"privateFile": bson.M{"$ne": true}},
			{
				"$and": []bson.M{
					{"privateFile": true},
					{"owner": username},
				},
			},
		},
	}
}

func (s *MongoStore) Close(ctx context.Context) error {
	if s.client == nil {
		return nil
	}
	return s.client.Disconnect(ctx)
}
