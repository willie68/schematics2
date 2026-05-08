package domain

type SearchResult struct {
	Document Document `json:"document"`
	Score    int      `json:"score"`
}
