package index

import (
	"log/slog"
	"strings"

	"github.com/samber/do/v2"
	"github.com/willie68/schematic2/backend/internal/domain/model"
	"github.com/willie68/schematic2/backend/internal/logging"
)

// docStoreInterface defines the interface for the document store
type docStoreInterface interface {
	Search(filter model.SearchFilter) model.PagedSearchResult
}

type MongoIndex struct {
	docStore docStoreInterface
	logger   *slog.Logger
}

func NewMongoIndex(inj do.Injector) *MongoIndex {
	return &MongoIndex{
		docStore: do.MustInvokeAs[docStoreInterface](inj),
		logger:   logging.New("mongo-index"),
	}
}

// Search performs full-text and tag-based search using MongoDB queries
func (idx *MongoIndex) Search(query string, tags []string, skip, limit int64, sortField string, sortOrder int, privateOnly, isAuthenticated bool, username string) model.PagedSearchResult {
	normTags := make([]string, 0, len(tags))
	for _, t := range tags {
		trimmed := strings.ToLower(strings.TrimSpace(t))
		if trimmed != "" {
			normTags = append(normTags, trimmed)
		}
	}

	filter := model.SearchFilter{
		Query:           query,
		Tags:            normTags,
		Skip:            skip,
		Limit:           limit,
		SortField:       sortField,
		SortOrder:       sortOrder,
		PrivateOnly:     privateOnly,
		IsAuthenticated: isAuthenticated,
		Username:        username,
	}

	return idx.docStore.Search(filter)
}
