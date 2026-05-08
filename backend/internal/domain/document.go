package domain

type Document struct {
	ID    string   `json:"id"`
	Title string   `json:"title"`
	Path  string   `json:"path"`
	Tags  []string `json:"tags"`
	Text  string   `json:"text"`
}
