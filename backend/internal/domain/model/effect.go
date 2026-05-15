package model

import (
	"time"
)

type Effect struct {
	ID             string         `json:"id,omitempty" bson:"_id,omitempty"`
	ForeignID      string         `json:"foreignId" bson:"foreignId,omitempty"`
	CreatedAt      time.Time      `json:"createdAt" bson:"createdAt,omitempty"`
	LastModifiedAt time.Time      `json:"lastModifiedAt" bson:"lastModifiedAt,omitempty"`
	EffectType     string         `json:"effectType" bson:"effectType,omitempty"`
	Manufacturer   string         `json:"manufacturer" bson:"manufacturer,omitempty"`
	Model          string         `json:"model" bson:"model,omitempty"`
	Tags           []string       `json:"tags" bson:"tags,omitempty"`
	Comment        string         `json:"comment" bson:"comment,omitempty"`
	Image          *ContainerInfo `json:"image,omitempty" bson:"image,omitempty"`
	Connector      string         `json:"connector" bson:"connector,omitempty"`
	Voltage        string         `json:"voltage" bson:"voltage,omitempty"`
	Current        string         `json:"current" bson:"current,omitempty"`
}
