package store

import (
	"errors"
	"sort"
	"sync"
)

var ErrDocumentExists = errors.New("document already exists")

type Document struct {
	ID    string   `json:"id"`
	Title string   `json:"title"`
	Path  string   `json:"path"`
	Tags  []string `json:"tags"`
	Text  string   `json:"text"`
}

type InMemoryDocumentStore struct {
	mu        sync.RWMutex
	documents map[string]Document
}

func NewInMemoryDocumentStore() *InMemoryDocumentStore {
	return &InMemoryDocumentStore{documents: make(map[string]Document)}
}

func (s *InMemoryDocumentStore) Upsert(doc Document) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.documents[doc.ID] = doc
	return nil
}

func (s *InMemoryDocumentStore) Get(id string) (Document, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	doc, ok := s.documents[id]
	return doc, ok
}

func (s *InMemoryDocumentStore) List() []Document {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Document, 0, len(s.documents))
	for _, doc := range s.documents {
		out = append(out, doc)
	}
	sort.Slice(out, func(i int, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}
