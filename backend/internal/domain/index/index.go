package index

import (
	"log/slog"
	"strings"

	"github.com/samber/do/v2"
	"github.com/willie68/schematics2/backend/internal/domain/model"
	"github.com/willie68/schematics2/backend/internal/logging"
)

// docStoreInterface defines the interface for the document store
type docStoreInterface interface {
	SearchStore(filter model.Query) model.PagedSearchResult
}

type searchIndex struct {
	docStore docStoreInterface
	log      *slog.Logger
}

func New(inj do.Injector) *searchIndex {
	return &searchIndex{
		docStore: do.MustInvokeAs[docStoreInterface](inj),
		log:      logging.New("domain-index"),
	}
}

func (i *searchIndex) Search(query model.Query) model.PagedSearchResult {
	// Parse the user query into structured terms
	query.ParsedQuery = i.parseQuery(query.Query)
	query.ParsedQuery.TagFilters = query.Tags

	// Pass the enhanced query to the store
	return i.docStore.SearchStore(query)
}

// parseQuery parses the user input query string and returns a structured ParsedQuery.
//
// Supported syntax:
// - Simple terms: "fender" or "fend*" (prefix match)
// - AND operator: "+term" (e.g., "fender +bassman" finds docs with both)
// - OR operator: implicit (space-separated terms are OR'd)
// - NOT operator: "-term" (e.g., "fender -broken" excludes docs with "broken")
// - Combinations: "fend* +bassm*" finds docs matching both prefix patterns
//
// Examples:
// "fender" → Optional: [{Value: "fender", IsPrefix: false}]
// "fend*" → Optional: [{Value: "fend", IsPrefix: true}]
// "fender +bassman" → Optional: [{Value: "fender"}], Required: [{Value: "bassman"}]
// "fend* +bassm*" → Optional: [{Value: "fend", IsPrefix: true}], Required: [{Value: "bassm", IsPrefix: true}]
// "-broken" → Excluded: [{Value: "broken"}]
func (i *searchIndex) parseQuery(queryStr string) model.ParsedQuery {
	result := model.ParsedQuery{
		Required:   []model.QueryTerm{},
		Optional:   []model.QueryTerm{},
		Excluded:   []model.QueryTerm{},
		TagFilters: []string{},
	}

	if strings.TrimSpace(queryStr) == "" {
		return result
	}

	// Split by spaces and process each term
	terms := strings.Fields(queryStr)
	for _, term := range terms {
		if term == "" {
			continue
		}

		// Check for modifiers: +, -, or no modifier
		if strings.HasPrefix(term, "+") {
			// Required term (AND logic)
			value := strings.TrimPrefix(term, "+")
			isPrefix := strings.HasSuffix(value, "*")
			if isPrefix {
				value = strings.TrimSuffix(value, "*")
			}
			result.Required = append(result.Required, model.QueryTerm{
				Value:    value,
				IsPrefix: isPrefix,
			})
		} else if strings.HasPrefix(term, "-") {
			// Excluded term (NOT logic)
			value := strings.TrimPrefix(term, "-")
			isPrefix := strings.HasSuffix(value, "*")
			if isPrefix {
				value = strings.TrimSuffix(value, "*")
			}
			result.Excluded = append(result.Excluded, model.QueryTerm{
				Value:    value,
				IsPrefix: isPrefix,
			})
		} else {
			// Optional term (OR logic)
			value := term
			isPrefix := strings.HasSuffix(value, "*")
			if isPrefix {
				value = strings.TrimSuffix(value, "*")
			}
			result.Optional = append(result.Optional, model.QueryTerm{
				Value:    value,
				IsPrefix: isPrefix,
			})
		}
	}

	return result
}
