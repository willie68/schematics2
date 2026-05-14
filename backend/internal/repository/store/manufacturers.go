package store

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (s *MongoStore) SuggestManufacturers(ctx context.Context, prefix string, limit int) ([]string, error) {
	if s.manufCol == nil {
		return nil, errors.New("mongodb manufacturers collection not initialised")
	}

	if limit <= 0 {
		limit = 10
	}

	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}).SetLimit(int64(limit))
		cur, err := s.manufCol.Find(ctx, bson.D{}, opts)
		if err != nil {
			return nil, fmt.Errorf("suggest manufacturers: %w", err)
		}
		defer cur.Close(ctx)

		var docs []bson.M
		if err = cur.All(ctx, &docs); err != nil {
			return nil, fmt.Errorf("decode manufacturers: %w", err)
		}

		out := make([]string, 0, len(docs))
		for _, doc := range docs {
			if id, ok := doc["_id"].(string); ok {
				out = append(out, id)
			}
		}
		return out, nil
	}

	// Case-insensitive regex for manufacturer search (stored with case preservation)
	filter := bson.D{{Key: "_id", Value: bson.D{
		{Key: "$regex", Value: "^" + regexp.QuoteMeta(prefix)},
		{Key: "$options", Value: "i"},
	}}}
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}).SetLimit(int64(limit))
	cur, err := s.manufCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("suggest manufacturers: %w", err)
	}
	defer cur.Close(ctx)

	var docs []bson.M
	if err = cur.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode manufacturers: %w", err)
	}

	out := make([]string, 0, len(docs))
	for _, doc := range docs {
		if id, ok := doc["_id"].(string); ok {
			out = append(out, id)
		}
	}

	return out, nil
}

func (s *MongoStore) UpdateManufacturer(ctx context.Context, manufacturer string) error {
	if s.manufCol == nil || manufacturer == "" {
		return nil
	}

	// Just ensure the manufacturer exists in the collection (upsert)
	_, err := s.manufCol.UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: manufacturer}},
		bson.D{{Key: "$setOnInsert", Value: bson.D{{Key: "_id", Value: manufacturer}}}},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("update manufacturer: %w", err)
	}

	return nil
}

func (s *MongoStore) updateManufacturer(ctx context.Context, manufacturer string) error {
	return s.UpdateManufacturer(ctx, manufacturer)
}
