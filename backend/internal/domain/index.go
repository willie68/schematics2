package domain

type SearchResult struct {
	Document Document `json:"document"`
}

// SearchFilter represents the internal search model for MongoDB queries
type SearchFilter struct {
	Query string   // Full-text query (can be empty)
	Tags  []string // Tags to filter by (all required, AND logic)
}
