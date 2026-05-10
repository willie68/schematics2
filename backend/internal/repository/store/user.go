package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/willie68/schematic2/backend/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// CreateUser adds a new user to the database
func (s *MongoStore) CreateUser(ctx context.Context, user domain.User) error {
	if s.usersCol == nil {
		return errors.New("mongodb users collection not initialised")
	}

	// Only set Created/Updated if not already set by caller
	if user.Created == 0 {
		user.Created = time.Now().UTC().Unix()
	}
	if user.Updated == 0 {
		user.Updated = user.Created
	}

	_, err := s.usersCol.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("user with this email already exists")
		}
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

// GetUserByEmail retrieves a user by email
func (s *MongoStore) GetUserByEmail(ctx context.Context, email string) (domain.User, bool) {
	if s.usersCol == nil {
		return domain.User{}, false
	}

	var user domain.User
	err := s.usersCol.FindOne(ctx, bson.D{{Key: "email", Value: email}}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.User{}, false
	}
	if err != nil {
		s.logger.Error("get user failed", "error", err, "email", email)
		return domain.User{}, false
	}

	return user, true
}
