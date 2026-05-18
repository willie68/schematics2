package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/willie68/schematics2/backend/internal/domain/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestBuildRegexFilterForTerm_PrefixMatch(t *testing.T) {
	store := &MongoStore{}

	term := model.QueryTerm{
		Value:    "fend",
		IsPrefix: true,
	}

	result := store.buildRegexFilterForTerm(term)

	// Verify structure: should be $or with multiple fields
	orValue, ok := result["$or"]
	assert.True(t, ok)

	orClauses, ok := orValue.([]bson.M)
	assert.True(t, ok)
	assert.Len(t, orClauses, 5) // manufacturer, model, subtitle, description, tags

	// Check that one clause contains the prefix pattern
	found := false
	for _, clause := range orClauses {
		if mfg, ok := clause["manufacturer"]; ok {
			regexFilter := mfg.(bson.M)
			if pattern, ok := regexFilter["$regex"]; ok {
				if pattern == "^fend" {
					found = true
					assert.Equal(t, "i", regexFilter["$options"])
				}
			}
		}
	}
	assert.True(t, found, "Expected to find ^fend pattern in manufacturer field")
}

func TestBuildRegexFilterForTerm_FullMatch(t *testing.T) {
	store := &MongoStore{}

	term := model.QueryTerm{
		Value:    "bassman",
		IsPrefix: false,
	}

	result := store.buildRegexFilterForTerm(term)

	// Verify structure: should be $or with multiple fields
	orValue, ok := result["$or"]
	assert.True(t, ok)

	orClauses, ok := orValue.([]bson.M)
	assert.True(t, ok)
	assert.Len(t, orClauses, 5)

	// Check that one clause contains the non-prefix pattern
	found := false
	for _, clause := range orClauses {
		if model, ok := clause["model"]; ok {
			regexFilter := model.(bson.M)
			if pattern, ok := regexFilter["$regex"]; ok {
				if pattern == "bassman" {
					found = true
					assert.Equal(t, "i", regexFilter["$options"])
				}
			}
		}
	}
	assert.True(t, found, "Expected to find bassman pattern in model field")
}

func TestBuildPrivateFileFilter_Guest(t *testing.T) {
	store := &MongoStore{}

	result := store.buildPrivateFileFilter(false, false, "")

	// Guests: only public documents
	expected := bson.M{"privateFile": bson.M{"$ne": true}}
	assert.Equal(t, expected, result)
}

func TestBuildPrivateFileFilter_AuthenticatedPrivateOnly(t *testing.T) {
	store := &MongoStore{}

	result := store.buildPrivateFileFilter(true, true, "testuser")

	// Should only allow user's private documents
	andClauses, ok := result["$and"].([]bson.M)
	assert.True(t, ok)
	assert.Len(t, andClauses, 2)

	// One clause checks privateFile=true, other checks owner
	hasPrivateTrue := false
	hasOwner := false

	for _, clause := range andClauses {
		if v, ok := clause["privateFile"]; ok && v == true {
			hasPrivateTrue = true
		}
		if _, ok := clause["owner"]; ok {
			hasOwner = true
			assert.Equal(t, "testuser", clause["owner"])
		}
	}

	assert.True(t, hasPrivateTrue)
	assert.True(t, hasOwner)
}

func TestBuildPrivateFileFilter_AuthenticatedPublicAndOwn(t *testing.T) {
	store := &MongoStore{}

	result := store.buildPrivateFileFilter(true, false, "testuser")

	// Should allow public documents OR user's private documents
	orClauses, ok := result["$or"].([]bson.M)
	assert.True(t, ok)
	assert.Len(t, orClauses, 2)

	// One clause checks public (privateFile != true), other checks user's private
	hasPublicClause := false
	hasPrivateClause := false

	for _, clause := range orClauses {
		if privateFileClause, ok := clause["privateFile"]; ok {
			if m, ok := privateFileClause.(bson.M); ok {
				if _, ok := m["$ne"]; ok {
					hasPublicClause = true
				}
			}
		}
		if andClauses, ok := clause["$and"].([]bson.M); ok {
			hasPrivateClause = true
			assert.Len(t, andClauses, 2)
		}
	}

	assert.True(t, hasPublicClause)
	assert.True(t, hasPrivateClause)
}

func TestBuildMongoFilterFromParsedQuery_EmptyQuery(t *testing.T) {
	store := &MongoStore{}

	query := model.Query{
		ParsedQuery: model.ParsedQuery{
			Required:   []model.QueryTerm{},
			Optional:   []model.QueryTerm{},
			Excluded:   []model.QueryTerm{},
			TagFilters: []string{},
		},
		IsAuthenticated: false,
	}

	result := store.buildMongoFilterFromParsedQuery(query)

	// Should only have private file filter for guests
	privateFileFilter := result["privateFile"]
	assert.NotNil(t, privateFileFilter)
}

func TestBuildMongoFilterFromParsedQuery_OptionalTermsOnly(t *testing.T) {
	store := &MongoStore{}

	query := model.Query{
		ParsedQuery: model.ParsedQuery{
			Required: []model.QueryTerm{},
			Optional: []model.QueryTerm{
				{Value: "fender", IsPrefix: false},
				{Value: "bassman", IsPrefix: false},
			},
			Excluded:   []model.QueryTerm{},
			TagFilters: []string{},
		},
		IsAuthenticated: false,
	}

	result := store.buildMongoFilterFromParsedQuery(query)

	// Should have $and combining optional terms OR with private filter
	andClauses, ok := result["$and"].([]bson.M)
	assert.True(t, ok)
	assert.Len(t, andClauses, 2) // optional OR clause + private file filter
}

func TestBuildMongoFilterFromParsedQuery_RequiredTermsOnly(t *testing.T) {
	store := &MongoStore{}

	query := model.Query{
		ParsedQuery: model.ParsedQuery{
			Required: []model.QueryTerm{
				{Value: "combo", IsPrefix: false},
			},
			Optional:   []model.QueryTerm{},
			Excluded:   []model.QueryTerm{},
			TagFilters: []string{},
		},
		IsAuthenticated: false,
	}

	result := store.buildMongoFilterFromParsedQuery(query)

	// Should have $and with required term AND private filter
	andClauses, ok := result["$and"].([]bson.M)
	assert.True(t, ok)
	assert.Len(t, andClauses, 2) // required term + private file filter
}

func TestBuildMongoFilterFromParsedQuery_ExcludedTermsOnly(t *testing.T) {
	store := &MongoStore{}

	query := model.Query{
		ParsedQuery: model.ParsedQuery{
			Required: []model.QueryTerm{},
			Optional: []model.QueryTerm{},
			Excluded: []model.QueryTerm{
				{Value: "broken", IsPrefix: false},
			},
			TagFilters: []string{},
		},
		IsAuthenticated: false,
	}

	result := store.buildMongoFilterFromParsedQuery(query)

	// Should have $and with $nor (excluded) AND private filter
	andClauses, ok := result["$and"].([]bson.M)
	assert.True(t, ok)
	assert.Len(t, andClauses, 2) // $nor clause + private file filter

	// Check for $nor
	found := false
	for _, clause := range andClauses {
		if _, ok := clause["$nor"]; ok {
			found = true
		}
	}
	assert.True(t, found, "Expected to find $nor clause")
}

func TestBuildMongoFilterFromParsedQuery_ComplexQuery(t *testing.T) {
	store := &MongoStore{}

	query := model.Query{
		ParsedQuery: model.ParsedQuery{
			Required: []model.QueryTerm{
				{Value: "combo", IsPrefix: false},
			},
			Optional: []model.QueryTerm{
				{Value: "fender", IsPrefix: true},
			},
			Excluded: []model.QueryTerm{
				{Value: "broken", IsPrefix: false},
			},
			TagFilters: []string{"amplifier", "vintage"},
		},
		IsAuthenticated: false,
	}

	result := store.buildMongoFilterFromParsedQuery(query)

	// Should have $and with multiple clauses
	andClauses, ok := result["$and"].([]bson.M)
	assert.True(t, ok)
	// Should have: required + optional + excluded + tags + private filter = 5
	assert.Len(t, andClauses, 5)
}

func TestBuildMongoFilterFromParsedQuery_TagFilters(t *testing.T) {
	store := &MongoStore{}

	query := model.Query{
		ParsedQuery: model.ParsedQuery{
			Required:   []model.QueryTerm{},
			Optional:   []model.QueryTerm{},
			Excluded:   []model.QueryTerm{},
			TagFilters: []string{"amplifier", "vintage"},
		},
		IsAuthenticated: false,
	}

	result := store.buildMongoFilterFromParsedQuery(query)

	// Should have $and with tag filter AND private filter
	andClauses, ok := result["$and"].([]bson.M)
	assert.True(t, ok)
	assert.Len(t, andClauses, 2) // tags filter + private filter

	// Check for tags $all
	found := false
	for _, clause := range andClauses {
		if tagsClause, ok := clause["tags"]; ok {
			if m, ok := tagsClause.(bson.M); ok {
				if allTags, ok := m["$all"].([]string); ok {
					if len(allTags) == 2 && allTags[0] == "amplifier" && allTags[1] == "vintage" {
						found = true
					}
				}
			}
		}
	}
	assert.True(t, found, "Expected to find tags $all clause")
}
