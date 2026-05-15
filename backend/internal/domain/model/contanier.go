package model

type ContainerInfo struct {
	ID              string `json:"id,omitempty" bson:"_id,omitempty"`
	ContainerNumber int    `json:"containerNumber" bson:"containerNumber"`
	Offset          int64  `json:"offset" bson:"offset"`
	Name            string `json:"name,omitempty" bson:"name,omitempty"` // Optional name for easier identification
	Length          int64  `json:"length" bson:"length"`
	OriginalLength  int64  `json:"originalLength" bson:"originalLength"` // Length before compression
	MIMEType        string `json:"mimetype,omitempty" bson:"mimetype,omitempty"`
	Compressed      string `json:"compressed,omitempty" bson:"compressed,omitempty"` // "none", "gzip", "zstd"
	Deleted         bool   `json:"deleted,omitempty" bson:"deleted,omitempty"`
}
