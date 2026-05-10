package index

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/willie68/schematic2/backend/internal/domain"
)

// MockDocumentStore implements docStoreInterface for testing
type MockDocumentStore struct {
	docs []domain.Document
}

func (m *MockDocumentStore) Search(filter domain.SearchFilter) []domain.SearchResult {
	results := make([]domain.SearchResult, 0)

	// Normalize tags for comparison
	requiredTags := make(map[string]struct{})
	for _, t := range filter.Tags {
		requiredTags[strings.ToLower(strings.TrimSpace(t))] = struct{}{}
	}

	// Filter documents
	for _, doc := range m.docs {
		// Check tag filter
		if len(requiredTags) > 0 {
			// All tags must match
			docTagSet := make(map[string]struct{})
			for _, tag := range doc.Tags {
				docTagSet[strings.ToLower(strings.TrimSpace(tag))] = struct{}{}
			}

			match := true
			for reqTag := range requiredTags {
				if _, exists := docTagSet[reqTag]; !exists {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}

		// Check text filter
		if filter.Query != "" {
			searchText := strings.ToLower(strings.Join([]string{
				doc.Manufacturer,
				doc.Model,
				doc.Subtitle,
				doc.Description,
				strings.Join(doc.Tags, " "),
			}, " "))

			if !strings.Contains(searchText, strings.ToLower(filter.Query)) {
				continue
			}
		}

		results = append(results, domain.SearchResult{Document: doc})
	}

	return results
}

func NewMongoIndexWithMockStore(docs []domain.Document) *MongoIndex {
	return &MongoIndex{
		docStore: &MockDocumentStore{docs: docs},
		logger:   nil,
	}
}

func TestMongoIndex_SearchByTagsOnly(t *testing.T) {
	// Setup
	docs := []domain.Document{
		{
			ID:           "doc1",
			Manufacturer: "Siemens",
			Model:        "S7-1200",
			Tags:         []string{"PLC", "automation"},
		},
		{
			ID:           "doc2",
			Manufacturer: "Philips",
			Model:        "Oscilloscope",
			Tags:         []string{"test", "measurement"},
		},
		{
			ID:           "doc3",
			Manufacturer: "Bosch",
			Model:        "Drill",
			Tags:         []string{"PLC", "tool"},
		},
	}

	idx := NewMongoIndexWithMockStore(docs)

	// Test 1: Search by single tag
	results := idx.Search("", []string{"PLC"})
	require.NotNil(t, results)
	assert.Len(t, results, 2)
	assert.Equal(t, "doc1", results[0].Document.ID)
	assert.Equal(t, "doc3", results[1].Document.ID)

	// Test 2: Search by multiple tags (all required)
	results = idx.Search("", []string{"PLC", "automation"})
	require.NotNil(t, results)
	assert.Len(t, results, 1)
	assert.Equal(t, "doc1", results[0].Document.ID)

	// Test 3: No match on second tag
	results = idx.Search("", []string{"PLC", "measurement"})
	require.NotNil(t, results)
	assert.Len(t, results, 0)

	// Test 4: Case insensitive tag matching
	results = idx.Search("", []string{"plc"})
	require.NotNil(t, results)
	assert.Len(t, results, 2)
}

func TestMongoIndex_SearchByQueryOnly(t *testing.T) {
	// Setup
	docs := []domain.Document{
		{
			ID:           "doc1",
			Manufacturer: "Siemens",
			Model:        "S7-1200",
			Subtitle:     "PLC System",
			Description:  "Programmable Logic Controller for automation",
		},
		{
			ID:           "doc2",
			Manufacturer: "Philips",
			Model:        "Oscilloscope",
			Subtitle:     "Test Equipment",
			Description:  "Electronic measurement tool",
		},
		{
			ID:           "doc3",
			Manufacturer: "Bosch",
			Model:        "Power Drill",
			Subtitle:     "Tool",
			Description:  "Electrical drilling tool for construction",
		},
	}

	idx := NewMongoIndexWithMockStore(docs)

	// Test 1: Search for "PLC"
	results := idx.Search("PLC", nil)
	require.NotNil(t, results)
	assert.Len(t, results, 1)
	assert.Equal(t, "doc1", results[0].Document.ID)

	// Test 2: Search for "tool" (should match doc2 and doc3)
	results = idx.Search("tool", nil)
	require.NotNil(t, results)
	assert.Len(t, results, 2)
	// Results should be sorted by score
	assert.Contains(t, []string{"doc2", "doc3"}, results[0].Document.ID)
	assert.Contains(t, []string{"doc2", "doc3"}, results[1].Document.ID)

	// Test 3: Search for non-existent term
	results = idx.Search("nonexistent", nil)
	require.NotNil(t, results)
	assert.Len(t, results, 0)
}

func TestMongoIndex_SearchByQueryAndTags(t *testing.T) {
	// Setup
	docs := []domain.Document{
		{
			ID:           "doc1",
			Manufacturer: "Siemens",
			Model:        "S7-1200",
			Subtitle:     "PLC System",
			Tags:         []string{"PLC", "automation", "industrial"},
		},
		{
			ID:           "doc2",
			Manufacturer: "Siemens",
			Model:        "Logo",
			Subtitle:     "Compact PLC",
			Tags:         []string{"PLC", "small"},
		},
		{
			ID:           "doc3",
			Manufacturer: "Philips",
			Model:        "Oscilloscope",
			Tags:         []string{"test", "measurement"},
		},
	}

	idx := NewMongoIndexWithMockStore(docs)

	// Test 1: Search "PLC" with tag "automation"
	// Should only return doc1 (has "PLC" in description AND "automation" tag)
	results := idx.Search("PLC", []string{"automation"})
	require.NotNil(t, results)
	assert.Len(t, results, 1)
	assert.Equal(t, "doc1", results[0].Document.ID)

	// Test 2: Search "PLC" with tag "PLC" (redundant but valid)
	results = idx.Search("PLC", []string{"PLC"})
	require.NotNil(t, results)
	assert.Len(t, results, 2)
	// Should match doc1 and doc2 (both have "PLC" tag and "PLC" in text)
	assert.Contains(t, []string{results[0].Document.ID, results[1].Document.ID}, "doc1")
	assert.Contains(t, []string{results[0].Document.ID, results[1].Document.ID}, "doc2")

	// Test 3: Search "PLC" with tag "measurement" (should have no match)
	results = idx.Search("PLC", []string{"measurement"})
	require.NotNil(t, results)
	assert.Len(t, results, 0)
}

func TestMongoIndex_SearchAll(t *testing.T) {
	// Setup
	docs := []domain.Document{
		{
			ID:           "doc1",
			Manufacturer: "Siemens",
			Model:        "S7-1200",
		},
		{
			ID:           "doc2",
			Manufacturer: "Philips",
			Model:        "Oscilloscope",
		},
	}

	idx := NewMongoIndexWithMockStore(docs)

	// Test: Empty query and no tags should return all documents
	results := idx.Search("", nil)
	require.NotNil(t, results)
	assert.Len(t, results, 2)
}
