package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/willie68/schematic2/backend/internal/domain"
)

func TestInMemoryIndex_SearchByTagsOnly(t *testing.T) {
	// Setup
	idx := NewInMemoryIndex(nil)

	doc1 := domain.Document{
		ID:           "doc1",
		Manufacturer: "Siemens",
		Model:        "S7-1200",
		Tags:         []string{"PLC", "automation"},
	}

	doc2 := domain.Document{
		ID:           "doc2",
		Manufacturer: "Philips",
		Model:        "Oscilloscope",
		Tags:         []string{"test", "measurement"},
	}

	doc3 := domain.Document{
		ID:           "doc3",
		Manufacturer: "Bosch",
		Model:        "Drill",
		Tags:         []string{"PLC", "tool"},
	}

	idx.Upsert(doc1)
	idx.Upsert(doc2)
	idx.Upsert(doc3)

	// Test 1: Search by single tag
	results := idx.Search("", []string{"PLC"})
	require.Len(t, results, 2)
	assert.ElementsMatch(t, []string{"doc1", "doc3"}, []string{results[0].Document.ID, results[1].Document.ID})

	// Test 2: Search by multiple tags (AND logic)
	results = idx.Search("", []string{"PLC", "automation"})
	require.Len(t, results, 1)
	assert.Equal(t, "doc1", results[0].Document.ID)

	// Test 3: Search by non-matching tag
	results = idx.Search("", []string{"nonexistent"})
	require.Len(t, results, 0)

	// Test 4: Empty query and empty tags - return all
	results = idx.Search("", []string{})
	require.Len(t, results, 3)
}

func TestInMemoryIndex_SearchByQueryAndTags(t *testing.T) {
	// Setup
	idx := NewInMemoryIndex(nil)

	doc1 := domain.Document{
		ID:           "doc1",
		Manufacturer: "Siemens",
		Model:        "S7-1200",
		Tags:         []string{"PLC"},
	}

	doc2 := domain.Document{
		ID:           "doc2",
		Manufacturer: "Siemens",
		Model:        "Oscilloscope",
		Tags:         []string{"measurement"},
	}

	idx.Upsert(doc1)
	idx.Upsert(doc2)

	// Test: Search by query AND tag filter
	results := idx.Search("Siemens", []string{"PLC"})
	require.Len(t, results, 1)
	assert.Equal(t, "doc1", results[0].Document.ID)

	// Test: Search by query but no matching tags
	results = idx.Search("Siemens", []string{"nonexistent"})
	require.Len(t, results, 0)
}

func TestInMemoryIndex_SearchByQueryOnly(t *testing.T) {
	// Setup
	idx := NewInMemoryIndex(nil)

	doc1 := domain.Document{
		ID:           "doc1",
		Manufacturer: "Siemens",
		Model:        "S7-1200",
		Tags:         []string{"PLC"},
	}

	doc2 := domain.Document{
		ID:           "doc2",
		Manufacturer: "Philips",
		Model:        "Oscilloscope",
		Tags:         []string{"measurement"},
	}

	idx.Upsert(doc1)
	idx.Upsert(doc2)

	// Test: Query search only
	results := idx.Search("Siemens", []string{})
	require.Len(t, results, 1)
	assert.Equal(t, "doc1", results[0].Document.ID)
}
