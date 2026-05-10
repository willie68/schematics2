package index

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/samber/do/v2"
	"github.com/willie68/schematic2/backend/internal/domain"
	"github.com/willie68/schematic2/backend/internal/logging"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// MongoIndex implements documentIndex using MongoDB for search operations
// docStoreInterface defines the interface for the document store
type docStoreInterface interface {
	List() []domain.Document
}

type MongoIndex struct {
	docStore docStoreInterface
	logger   *slog.Logger
}

func NewMongoIndex(inj do.Injector) *MongoIndex {
	// Invoke the document store - it will be whatever was registered (mongoDocumentStore)
	// We use a generic interface to avoid direct type dependency
	docStore, err := do.Invoke[interface {
		List() []domain.Document
	}](inj)
	if err != nil {
		// If we can't get a proper store, create a dummy one that will fail
		panic(fmt.Sprintf("failed to get document store: %v", err))
	}

	return &MongoIndex{
		docStore: docStore,
		logger:   logging.New("mongo-index"),
	}
}

// Upsert updates the index with a document (no-op for MongoDB as it's stored directly in DB)
func (idx *MongoIndex) Upsert(doc domain.Document) {
	// In MongoDB-based index, documents are already persisted in the document store.
	// This method is kept for interface compatibility but performs no additional indexing.
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

	// Build MongoDB filter
	var filter bson.M

	if query == "" && len(normTags) == 0 {
		// No filters - return all documents
		filter = bson.M{}
	} else if query == "" && len(normTags) > 0 {
		// Tag-only search: all tags must match
		filter = bson.M{
			"tags": bson.M{
				"$all": normTags,
			},
		}
	} else if query != "" && len(normTags) == 0 {
		// Full-text search only
		filter = bson.M{
			"$text": bson.M{
				"$search": query,
			},
		}
	} else {
		// Both query and tags: combine with AND
		filter = bson.M{
			"$and": []bson.M{
				{
					"$text": bson.M{
						"$search": query,
					},
				},
				{
					"tags": bson.M{
						"$all": normTags,
					},
				},
			},
		}
	}

	// Execute search using document store
	results, err := idx.searchDocuments(filter, query)
	if err != nil {
		idx.logger.Error("search failed", "error", err)
		return nil
	}

	return results
}

// searchDocuments executes the MongoDB search and returns results
func (idx *MongoIndex) searchDocuments(filter bson.M, query string) ([]domain.SearchResult, error) {
	// Get all documents and filter in-memory
	// This is a workaround since MongoDB text search scoring isn't directly accessible from the Go driver
	docs := idx.docStore.List()
	if len(docs) == 0 {
		return nil, nil
	}

	results := make([]domain.SearchResult, 0, len(docs))

	for _, doc := range docs {
		// Check if document matches filter
		if idx.matchesFilter(doc, filter) {
			// Calculate score based on text match
			score := idx.calculateScore(doc, query)
			results = append(results, domain.SearchResult{
				Document: doc,
				Score:    score,
			})
		}
	}

	// Sort by score (descending), then by ID (ascending)
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].Document.ID < results[j].Document.ID
		}
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// matchesFilter checks if a document matches the MongoDB filter
func (idx *MongoIndex) matchesFilter(doc domain.Document, filter bson.M) bool {
	if len(filter) == 0 {
		// Empty filter matches all
		return true
	}

	// Handle text search
	if textSearch, ok := filter["$text"]; ok {
		textM := textSearch.(bson.M)
		if search, ok := textM["$search"]; ok {
			searchStr := search.(string)
			if !idx.matchesText(doc, searchStr) {
				return false
			}
		}
	}

	// Handle tag filtering
	if tagsFilter, ok := filter["tags"]; ok {
		tagsM := tagsFilter.(bson.M)
		if allTags, ok := tagsM["$all"]; ok {
			allTagsList := allTags.([]string)
			if !idx.hasAllTags(doc.Tags, allTagsList) {
				return false
			}
		}
	}

	// Handle AND logic
	if andConds, ok := filter["$and"]; ok {
		andList := andConds.([]bson.M)
		for _, cond := range andList {
			if !idx.matchesFilter(doc, cond) {
				return false
			}
		}
	}

	return true
}

// matchesText checks if document contains query terms
func (idx *MongoIndex) matchesText(doc domain.Document, query string) bool {
	searchText := strings.ToLower(strings.Join([]string{
		doc.Manufacturer,
		doc.Model,
		doc.Subtitle,
		doc.Description,
		strings.Join(doc.Tags, " "),
		idx.getFileNames(doc.Files),
	}, " "))

	// Tokenize query and search
	for _, token := range tokenize(query) {
		if !strings.Contains(searchText, token) {
			return false
		}
	}
	return true
}

// getFileNames extracts file names from document files
func (idx *MongoIndex) getFileNames(files []domain.DocumentFile) string {
	names := make([]string, 0, len(files)*2)
	for _, f := range files {
		if f.Name != "" {
			names = append(names, f.Name)
		}
		if f.Type != "" {
			names = append(names, f.Type)
		}
	}
	return strings.Join(names, " ")
}

// hasAllTags checks if document has all required tags
func (idx *MongoIndex) hasAllTags(docTags []string, requiredTags []string) bool {
	if len(requiredTags) == 0 {
		return true
	}

	docTagSet := make(map[string]struct{}, len(docTags))
	for _, tag := range docTags {
		docTagSet[strings.ToLower(strings.TrimSpace(tag))] = struct{}{}
	}

	for _, reqTag := range requiredTags {
		if _, exists := docTagSet[strings.ToLower(strings.TrimSpace(reqTag))]; !exists {
			return false
		}
	}
	return true
}

// calculateScore computes relevance score for a document
func (idx *MongoIndex) calculateScore(doc domain.Document, query string) int {
	if query == "" {
		return 0 // No query = no score
	}

	score := 0
	queryTokens := tokenize(query)
	searchText := strings.ToLower(strings.Join([]string{
		doc.Manufacturer,
		doc.Model,
		doc.Subtitle,
		doc.Description,
		strings.Join(doc.Tags, " "),
	}, " "))

	for _, token := range queryTokens {
		// Count occurrences for scoring
		count := strings.Count(searchText, token)
		score += count
	}

	return score
}
