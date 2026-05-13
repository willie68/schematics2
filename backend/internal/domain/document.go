package domain

import "time"

type Tag struct {
	Name    string `json:"name"`
	Counter int64  `json:"counter"`
}

type ContainerInfo struct {
	ContainerNumber int    `json:"containerNumber" bson:"containerNumber"`
	Offset          int64  `json:"offset" bson:"offset"`
	Length          int64  `json:"length" bson:"length"`
	OriginalLength  int64  `json:"originalLength" bson:"originalLength"` // Length before compression
	MIMEType        string `json:"mimetype,omitempty" bson:"mimetype,omitempty"`
	Compressed      string `json:"compressed,omitempty" bson:"compressed,omitempty"` // "none", "gzip", "zstd"
	Deleted         bool   `json:"deleted,omitempty" bson:"deleted,omitempty"`       // Soft delete flag for files
}

type DocumentFile struct {
	Name      string         `json:"name" bson:"name"`
	Page      int            `json:"page" bson:"page"`
	MIMEType  string         `json:"mimetype" bson:"mimetype"`
	Type      string         `json:"type" bson:"type"`
	Container *ContainerInfo `json:"container,omitempty" bson:"container,omitempty"`
	Data      string         `json:"data,omitempty" bson:"-"`
}

type Document struct {
	ID             string         `json:"id"`
	CreatedAt      time.Time      `json:"createdAt"`
	LastModifiedAt time.Time      `json:"lastModifiedAt"`
	Manufacturer   string         `json:"manufacturer"`
	Model          string         `json:"model"`
	Subtitle       string         `json:"subtitle"`
	Tags           []string       `json:"tags"`
	Description    string         `json:"description"`
	PrivateFile    bool           `json:"privateFile"`
	Owner          string         `json:"owner"`
	Files          []DocumentFile `json:"files"`
}
