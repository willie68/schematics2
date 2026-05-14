package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/willie68/schematic2/backend/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const effectsCollection = "effects"

// SearchEffects searches for effects with pagination and sorting
func (m *MongoStore) SearchEffects(ctx context.Context, query string, skip, limit int64, sortField, sortOrder string) (domain.PagedEffects, error) {
	filter := bson.M{}

	// Build search filter if query is provided
	if strings.TrimSpace(query) != "" {
		q := strings.TrimSpace(query)
		filter = bson.M{
			"$or": []bson.M{
				{"effectType": bson.M{"$regex": q, "$options": "i"}},
				{"manufacturer": bson.M{"$regex": q, "$options": "i"}},
				{"model": bson.M{"$regex": q, "$options": "i"}},
				{"tags": bson.M{"$regex": q, "$options": "i"}},
				{"comment": bson.M{"$regex": q, "$options": "i"}},
			},
		}
	}

	// Get total count
	total, err := m.effectsCol.CountDocuments(ctx, filter)
	if err != nil {
		return domain.PagedEffects{}, fmt.Errorf("count documents: %w", err)
	}

	// Determine sort order: default 1 (ascending), use -1 for descending
	sortValue := int32(1)
	if sortOrder == "desc" {
		sortValue = -1
	}

	// Map frontend field names to MongoDB field names
	sortFieldMapped := mapEffectSortField(sortField)

	// Build sort specification, avoiding duplicate keys
	sortFields := bson.D{
		{Key: sortFieldMapped, Value: sortValue},
	}
	// Add secondary sort by model only if not already sorting by model
	if sortFieldMapped != "model" {
		sortFields = append(sortFields, bson.E{Key: "model", Value: 1})
	}

	// Query with pagination and sorting
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(sortFields)
	cursor, err := m.effectsCol.Find(ctx, filter, opts)
	if err != nil {
		return domain.PagedEffects{}, fmt.Errorf("find effects: %w", err)
	}
	defer cursor.Close(ctx)

	var effects []domain.Effect
	if err = cursor.All(ctx, &effects); err != nil {
		return domain.PagedEffects{}, fmt.Errorf("decode effects: %w", err)
	}

	return domain.PagedEffects{
		Items: effects,
		Total: total,
		Skip:  skip,
		Limit: limit,
	}, nil
}

// mapEffectSortField maps frontend field names to MongoDB field names
func mapEffectSortField(field string) string {
	switch field {
	case "effectType":
		return "effectType"
	case "manufacturer":
		return "manufacturer"
	case "model":
		return "model"
	case "voltage":
		return "voltage"
	case "current":
		return "current"
	default:
		return "manufacturer"
	}
}

// GetEffectByID retrieves a single effect by ID
func (m *MongoStore) GetEffectByID(ctx context.Context, id string) (*domain.Effect, error) {
	var effect domain.Effect
	err := m.effectsCol.FindOne(ctx, bson.M{"_id": id}).Decode(&effect)
	if err != nil {
		return nil, fmt.Errorf("get effect: %w", err)
	}
	return &effect, nil
}

// CreateEffect creates a new effect in the database
func (m *MongoStore) CreateEffect(ctx context.Context, effect *domain.Effect) error {
	if effect.ID == "" {
		effect.ID = fmt.Sprintf("effect_%d", time.Now().UnixNano())
	}

	_, err := m.effectsCol.InsertOne(ctx, effect)
	if err != nil {
		return fmt.Errorf("insert effect: %w", err)
	}
	return nil
}

// UpdateEffect updates an existing effect in the database
func (m *MongoStore) UpdateEffect(ctx context.Context, effect *domain.Effect) error {
	if effect.ID == "" {
		return fmt.Errorf("effect ID is required for update")
	}

	effect.LastModifiedAt = time.Now()

	result, err := m.effectsCol.ReplaceOne(ctx, bson.M{"_id": effect.ID}, effect)
	if err != nil {
		return fmt.Errorf("replace effect: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("effect not found")
	}

	return nil
}
