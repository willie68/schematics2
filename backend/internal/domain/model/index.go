package model

type SearchResult struct {
	Document Document `json:"document"`
}

type PagedSearchResult struct {
	Results []SearchResult `json:"results"`
	Total   int64          `json:"total"`
	Skip    int64          `json:"skip"`
	Limit   int64          `json:"limit"`
}

type PagedEffects struct {
	Items []Effect `json:"items"`
	Total int64    `json:"total"`
	Skip  int64    `json:"skip"`
	Limit int64    `json:"limit"`
}

// SearchFilter represents the internal search model for MongoDB queries
type SearchFilter struct {
	Query           string   // Full-text query (can be empty)
	Tags            []string // Tags to filter by (all required, AND logic)
	Skip            int64    // Number of results to skip (pagination)
	Limit           int64    // Maximum number of results to return (0 = no limit)
	SortField       string   // Field to sort by (e.g. "manufacturer", "model"); empty = default (_id)
	SortOrder       int      // 1 = ascending, -1 = descending
	PrivateOnly     bool     // If true AND IsAuthenticated, only return documents with privateFile=true
	Username        string   // Current authenticated user, for owner-based filtering
	IsAuthenticated bool     // Whether the request is authenticated (determines private file visibility)
}
