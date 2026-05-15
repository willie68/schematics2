package store

import (
	"context"
	"fmt"

	"github.com/willie68/schematic2/backend/internal/domain/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const effectTypesCollection = "effecttypes"

// GetAllEffectTypes retrieves all effect types from the database
func (m *MongoStore) GetAllEffectTypes(ctx context.Context) ([]model.EffectType, error) {
	filter := bson.M{}

	opts := options.Find().SetSort(bson.D{
		{Key: "typeName", Value: 1},
	})
	cursor, err := m.effectTypesCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("find effect types: %w", err)
	}
	defer cursor.Close(ctx)

	var effectTypes []model.EffectType
	if err = cursor.All(ctx, &effectTypes); err != nil {
		return nil, fmt.Errorf("decode effect types: %w", err)
	}

	return effectTypes, nil
}
