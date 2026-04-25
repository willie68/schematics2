package index

import (
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/willie68/schematic2/backend/internal/store"
)

type SearchResult struct {
	Document store.Document `json:"document"`
	Score    int            `json:"score"`
}

type InMemoryIndex struct {
	mu       sync.RWMutex
	byToken  map[string]map[string]int
	docByID  map[string]store.Document
	tagsByID map[string]map[string]struct{}
}

func NewInMemoryIndex() *InMemoryIndex {
	return &InMemoryIndex{
		byToken:  make(map[string]map[string]int),
		docByID:  make(map[string]store.Document),
		tagsByID: make(map[string]map[string]struct{}),
	}
}

func (idx *InMemoryIndex) Upsert(doc store.Document) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.docByID[doc.ID] = doc
	tagSet := make(map[string]struct{}, len(doc.Tags))
	for _, t := range doc.Tags {
		tagSet[strings.ToLower(strings.TrimSpace(t))] = struct{}{}
	}
	idx.tagsByID[doc.ID] = tagSet

	for _, token := range tokenize(doc.Title + " " + doc.Text + " " + strings.Join(doc.Tags, " ")) {
		if _, ok := idx.byToken[token]; !ok {
			idx.byToken[token] = make(map[string]int)
		}
		idx.byToken[token][doc.ID]++
	}
}

func (idx *InMemoryIndex) Search(query string, tags []string) []SearchResult {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	needTags := make([]string, 0, len(tags))
	for _, t := range tags {
		trimmed := strings.ToLower(strings.TrimSpace(t))
		if trimmed != "" {
			needTags = append(needTags, trimmed)
		}
	}

	scores := make(map[string]int)
	for _, token := range tokenize(query) {
		for docID, score := range idx.byToken[token] {
			scores[docID] += score
		}
	}

	results := make([]SearchResult, 0, len(scores))
	for docID, score := range scores {
		doc, ok := idx.docByID[docID]
		if !ok {
			continue
		}
		if !hasAllTags(idx.tagsByID[docID], needTags) {
			continue
		}
		results = append(results, SearchResult{Document: doc, Score: score})
	}

	sort.Slice(results, func(i int, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].Document.ID < results[j].Document.ID
		}
		return results[i].Score > results[j].Score
	})
	return results
}

func hasAllTags(current map[string]struct{}, wanted []string) bool {
	if len(wanted) == 0 {
		return true
	}
	for _, tag := range wanted {
		if _, ok := current[tag]; !ok {
			return false
		}
	}
	return true
}

func tokenize(input string) []string {
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return unicode.ToLower(r)
		}
		return ' '
	}, input)

	parts := strings.Fields(cleaned)
	if len(parts) == 0 {
		return nil
	}
	return parts
}
