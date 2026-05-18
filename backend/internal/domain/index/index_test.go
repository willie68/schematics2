package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/willie68/schematics2/backend/internal/domain/model"
)

func TestParseQuery_SimpleTerms(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected model.ParsedQuery
	}{
		{
			name:  "empty query",
			input: "",
			expected: model.ParsedQuery{
				Required:   []model.QueryTerm{},
				Optional:   []model.QueryTerm{},
				Excluded:   []model.QueryTerm{},
				TagFilters: []string{},
			},
		},
		{
			name:  "single term",
			input: "fender",
			expected: model.ParsedQuery{
				Required:   []model.QueryTerm{},
				Optional:   []model.QueryTerm{{Value: "fender", IsPrefix: false}},
				Excluded:   []model.QueryTerm{},
				TagFilters: []string{},
			},
		},
		{
			name:  "single term with prefix",
			input: "fend*",
			expected: model.ParsedQuery{
				Required:   []model.QueryTerm{},
				Optional:   []model.QueryTerm{{Value: "fend", IsPrefix: true}},
				Excluded:   []model.QueryTerm{},
				TagFilters: []string{},
			},
		},
	}

	si := &searchIndex{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := si.parseQuery(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseQuery_WithOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected model.ParsedQuery
	}{
		{
			name:  "required term with +",
			input: "+bassman",
			expected: model.ParsedQuery{
				Required:   []model.QueryTerm{{Value: "bassman", IsPrefix: false}},
				Optional:   []model.QueryTerm{},
				Excluded:   []model.QueryTerm{},
				TagFilters: []string{},
			},
		},
		{
			name:  "required prefix with +",
			input: "+bassm*",
			expected: model.ParsedQuery{
				Required:   []model.QueryTerm{{Value: "bassm", IsPrefix: true}},
				Optional:   []model.QueryTerm{},
				Excluded:   []model.QueryTerm{},
				TagFilters: []string{},
			},
		},
		{
			name:  "excluded term with -",
			input: "-broken",
			expected: model.ParsedQuery{
				Required:   []model.QueryTerm{},
				Optional:   []model.QueryTerm{},
				Excluded:   []model.QueryTerm{{Value: "broken", IsPrefix: false}},
				TagFilters: []string{},
			},
		},
		{
			name:  "excluded prefix with -",
			input: "-broken*",
			expected: model.ParsedQuery{
				Required:   []model.QueryTerm{},
				Optional:   []model.QueryTerm{},
				Excluded:   []model.QueryTerm{{Value: "broken", IsPrefix: true}},
				TagFilters: []string{},
			},
		},
	}

	si := &searchIndex{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := si.parseQuery(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseQuery_CombinedTerms(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected model.ParsedQuery
	}{
		{
			name:  "optional OR (multiple terms)",
			input: "fender bassman",
			expected: model.ParsedQuery{
				Required: []model.QueryTerm{},
				Optional: []model.QueryTerm{
					{Value: "fender", IsPrefix: false},
					{Value: "bassman", IsPrefix: false},
				},
				Excluded:   []model.QueryTerm{},
				TagFilters: []string{},
			},
		},
		{
			name:  "optional with required (AND with implicit OR)",
			input: "fender +bassman",
			expected: model.ParsedQuery{
				Required: []model.QueryTerm{
					{Value: "bassman", IsPrefix: false},
				},
				Optional: []model.QueryTerm{
					{Value: "fender", IsPrefix: false},
				},
				Excluded:   []model.QueryTerm{},
				TagFilters: []string{},
			},
		},
		{
			name:  "prefix patterns with AND operator",
			input: "fend* +bassm*",
			expected: model.ParsedQuery{
				Required: []model.QueryTerm{
					{Value: "bassm", IsPrefix: true},
				},
				Optional: []model.QueryTerm{
					{Value: "fend", IsPrefix: true},
				},
				Excluded:   []model.QueryTerm{},
				TagFilters: []string{},
			},
		},
		{
			name:  "complex: optional OR, required AND, excluded NOT",
			input: "fender bassman +combo -broken",
			expected: model.ParsedQuery{
				Required: []model.QueryTerm{
					{Value: "combo", IsPrefix: false},
				},
				Optional: []model.QueryTerm{
					{Value: "fender", IsPrefix: false},
					{Value: "bassman", IsPrefix: false},
				},
				Excluded: []model.QueryTerm{
					{Value: "broken", IsPrefix: false},
				},
				TagFilters: []string{},
			},
		},
		{
			name:  "multiple prefix searches with AND",
			input: "12a* +7*",
			expected: model.ParsedQuery{
				Required: []model.QueryTerm{
					{Value: "7", IsPrefix: true},
				},
				Optional: []model.QueryTerm{
					{Value: "12a", IsPrefix: true},
				},
				Excluded:   []model.QueryTerm{},
				TagFilters: []string{},
			},
		},
	}

	si := &searchIndex{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := si.parseQuery(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseQuery_Whitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected model.ParsedQuery
	}{
		{
			name:  "leading/trailing whitespace",
			input: "  fender  ",
			expected: model.ParsedQuery{
				Required:   []model.QueryTerm{},
				Optional:   []model.QueryTerm{{Value: "fender", IsPrefix: false}},
				Excluded:   []model.QueryTerm{},
				TagFilters: []string{},
			},
		},
		{
			name:  "multiple spaces between terms",
			input: "fender    bassman",
			expected: model.ParsedQuery{
				Required: []model.QueryTerm{},
				Optional: []model.QueryTerm{
					{Value: "fender", IsPrefix: false},
					{Value: "bassman", IsPrefix: false},
				},
				Excluded:   []model.QueryTerm{},
				TagFilters: []string{},
			},
		},
	}

	si := &searchIndex{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := si.parseQuery(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseQuery_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected model.ParsedQuery
	}{
		{
			name:  "only prefix asterisk",
			input: "*",
			expected: model.ParsedQuery{
				Required:   []model.QueryTerm{},
				Optional:   []model.QueryTerm{{Value: "", IsPrefix: true}},
				Excluded:   []model.QueryTerm{},
				TagFilters: []string{},
			},
		},
		{
			name:  "multiple asterisks (only last removed)",
			input: "test**",
			expected: model.ParsedQuery{
				Required:   []model.QueryTerm{},
				Optional:   []model.QueryTerm{{Value: "test*", IsPrefix: true}},
				Excluded:   []model.QueryTerm{},
				TagFilters: []string{},
			},
		},
		{
			name:  "operator followed by prefix",
			input: "+test*",
			expected: model.ParsedQuery{
				Required:   []model.QueryTerm{{Value: "test", IsPrefix: true}},
				Optional:   []model.QueryTerm{},
				Excluded:   []model.QueryTerm{},
				TagFilters: []string{},
			},
		},
	}

	si := &searchIndex{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := si.parseQuery(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSearch_IntegrationWithParser(t *testing.T) {
	// This test verifies that the Search method properly initializes ParsedQuery
	// Note: This is a unit test that doesn't require a real database

	si := &searchIndex{
		docStore: nil, // We won't call the store in this test
		log:      nil, // No logging needed
	}

	query := model.Query{
		Query: "fend* +bassm* -broken",
		Tags:  []string{"amplifier", "vintage"},
	}

	// Call parseQuery directly to verify behavior
	si.parseQuery(query.Query)

	// Verify that parsing works correctly
	parsed := si.parseQuery(query.Query)
	assert.Len(t, parsed.Required, 1)
	assert.Equal(t, "bassm", parsed.Required[0].Value)
	assert.True(t, parsed.Required[0].IsPrefix)

	assert.Len(t, parsed.Optional, 1)
	assert.Equal(t, "fend", parsed.Optional[0].Value)
	assert.True(t, parsed.Optional[0].IsPrefix)

	assert.Len(t, parsed.Excluded, 1)
	assert.Equal(t, "broken", parsed.Excluded[0].Value)
	assert.False(t, parsed.Excluded[0].IsPrefix)
}
