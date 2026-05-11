package domain

import (
	"time"
)

type EffectType struct {
	ID             string            `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt      time.Time         `json:"createdAt" bson:"createdAt,omitempty"`
	LastModifiedAt time.Time         `json:"lastModifiedAt" bson:"lastModifiedAt,omitempty"`
	TypeName       string            `json:"typeName" bson:"typeName,omitempty"`
	I18n           map[string]string `json:"i18n" bson:"i18n,omitempty"`
	TypeImage      string            `json:"typeImage" bson:"typeImage,omitempty"`
}
