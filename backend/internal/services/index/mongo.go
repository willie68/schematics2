package index

import (
	"log/slog"
	"strings"

	"github.com/samber/do/v2"
	"github.com/willie68/schematic2/backend/internal/domain"
	"github.com/willie68/schematic2/backend/internal/logging"
)

// docStoreInterface defines the interface for the document store
type docStoreInterface interface {
	Search(filter domain.SearchFilter) []domain.SearchResult
}

type MongoIndex struct {
	docStore docStoreInterface
	logger   *slog.Logger
}

func NewMongoIndex(inj do.Injector) *MongoIndex {
	// Invoke the document store - it will be whatever was registered (mongoDocumentStore)
	// We use a generic interface to avoid direct type dependency
	return &MongoIndex{
		docStore: do.MustInvokeAs[docStoreInterface](inj),
		logger:   logging.New("mongo-index"),
	}
}

// Search performs full-text and tag-based search using MongoDB queries
func (idx *MongoIndex) Search(query string, tags []string) []domain.SearchResult {
	// Normalize tags
	normTags := make([]string, 0, len(tags))
	for _, t := range tags {
		trimmed := strings.ToLower(strings.TrimSpace(t))
		if trimmed != "" {
			normTags = append(normTags, trimmed)
		}
	}

	// Build search filter
	filter := domain.SearchFilter{
		Query: query,
		Tags:  normTags,
	}

	// Delegate to document store for MongoDB query execution
	results := idx.docStore.Search(filter)
	if results == nil {
		return []domain.SearchResult{}
	}

	return results
}
