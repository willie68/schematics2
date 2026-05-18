package model

// Query represents the search query received from the API, which is then transformed into a SearchFilter for internal use.
type Query struct {
	Query           string     `json:"query"` // Full-text query (can be empty)
	Tags            []string   `json:"tags"`  // Tags to filter by (all required, AND logic)
	Pagination      Pagination `json:"pagination"`
	Sort            Sort       `json:"sort"`
	PrivateOnly     bool       `json:"privateOnly"`
	IsAuthenticated bool       `json:"isAuthenticated"`
	Username        string     `json:"username"` // Optional username for user-specific filtering (e.g. private schematics)
	// ParsedQuery is populated by the domain service to provide structured search terms
	ParsedQuery ParsedQuery `json:"-"` // Not serialized from API
}

type Pagination struct {
	Skip  int64 `json:"skip"`
	Limit int64 `json:"limit"`
}

type Sort struct {
	Field string `json:"field"`
	Order int    `json:"order"`
}

type SearchResult struct {
	Document Document `json:"document"`
}

type PagedSearchResult struct {
	Results    []SearchResult `json:"results"`
	Total      int64          `json:"total"`
	Pagination Pagination     `json:"pagination"`
}

type PagedEffects struct {
	Items      []Effect   `json:"items"`
	Total      int64      `json:"total"`
	Pagination Pagination `json:"pagination"`
}

// QueryTerm represents a single search term with modifiers
type QueryTerm struct {
	Value    string // The search value (e.g., "fend" or "bassman")
	IsPrefix bool   // true if the term should match as prefix (input: "fend*")
}

// ParsedQuery represents a parsed search query with different operator types
type ParsedQuery struct {
	Required   []QueryTerm // Terms with + operator (AND logic)
	Optional   []QueryTerm // Terms without operator (OR logic)
	Excluded   []QueryTerm // Terms with - operator (NOT logic)
	TagFilters []string    // Tag filters (from Tags field in Query)
}
