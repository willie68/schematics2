package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/willie68/schematics2/backend/internal/domain/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (s *MongoStore) ListTags(ctx context.Context) ([]model.Tag, error) {
	if s.tagsCol == nil {
		return nil, errors.New("mongodb tags collection not initialised")
	}

	opts := options.Find().SetSort(bson.D{{Key: "counter", Value: -1}})
	cur, err := s.tagsCol.Find(ctx, bson.D{}, opts)
	if err != nil {
		s.logger.Error("list tags failed", "error", err)
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer cur.Close(ctx)

	var mongoTags []mongoTag
	if err = cur.All(ctx, &mongoTags); err != nil {
		s.logger.Error("decode tag list failed", "error", err)
		return nil, fmt.Errorf("decode tags: %w", err)
	}

	out := make([]model.Tag, 0, len(mongoTags))
	for _, t := range mongoTags {
		out = append(out, model.Tag{Name: t.Tag, Counter: t.Counter})
	}

	return out, nil
}

func (s *MongoStore) SuggestTags(ctx context.Context, prefix string, limit int) ([]model.Tag, error) {
	if s.tagsCol == nil {
		return nil, errors.New("mongodb tags collection not initialised")
	}

	if limit <= 0 {
		limit = 10
	}

	prefix = strings.ToLower(strings.TrimSpace(prefix))
	if prefix == "" {
		opts := options.Find().SetSort(bson.D{{Key: "counter", Value: -1}}).SetLimit(int64(limit))
		cur, err := s.tagsCol.Find(ctx, bson.D{}, opts)
		if err != nil {
			return nil, fmt.Errorf("suggest tags: %w", err)
		}
		defer cur.Close(ctx)

		var mongoTags []mongoTag
		if err = cur.All(ctx, &mongoTags); err != nil {
			return nil, fmt.Errorf("decode tags: %w", err)
		}

		out := make([]model.Tag, 0, len(mongoTags))
		for _, t := range mongoTags {
			out = append(out, model.Tag{Name: t.Tag, Counter: t.Counter})
		}
		return out, nil
	}

	filter := bson.D{{Key: "_id", Value: bson.D{{Key: "$regex", Value: "^" + prefix}}}}
	opts := options.Find().SetSort(bson.D{{Key: "counter", Value: -1}}).SetLimit(int64(limit))
	cur, err := s.tagsCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("suggest tags: %w", err)
	}
	defer cur.Close(ctx)

	var mongoTags []mongoTag
	if err = cur.All(ctx, &mongoTags); err != nil {
		return nil, fmt.Errorf("decode tags: %w", err)
	}

	out := make([]model.Tag, 0, len(mongoTags))
	for _, t := range mongoTags {
		out = append(out, model.Tag{Name: t.Tag, Counter: t.Counter})
	}

	return out, nil
}
