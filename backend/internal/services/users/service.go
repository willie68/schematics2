package users

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"github.com/willie68/schematic2/backend/internal/auth"
	"github.com/willie68/schematic2/backend/internal/domain"
)

// userStore interface for user persistence
type userStore interface {
	CreateUser(ctx context.Context, user domain.User) error
	GetUserByEmail(ctx context.Context, email string) (domain.User, bool)
}

// Service manages user operations with flood protection
type Service struct {
	store       userStore
	mu          sync.Mutex
	minDuration time.Duration
}

// NewService creates a new user service with flood protection
// minDuration is the minimum time each operation should take (e.g., 10 seconds to prevent flooding)
func NewService(inj do.Injector, minDuration time.Duration) *Service {
	if minDuration < 0 {
		minDuration = 0
	}
	store := do.MustInvokeAs[userStore](inj)
	return &Service{
		store:       store,
		minDuration: minDuration,
	}
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Street    string `json:"street"`
	ZipCode   string `json:"zipCode"`
	City      string `json:"city"`
}

// Register creates a new user with flood protection (serialized + min duration)
func (s *Service) Register(ctx context.Context, req RegisterRequest) (domain.User, error) {
	start := time.Now()

	// Lock: only one registration at a time
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate input
	if err := validateRegisterRequest(req); err != nil {
		time.Sleep(time.Until(start.Add(s.minDuration)))
		return domain.User{}, err
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		time.Sleep(time.Until(start.Add(s.minDuration)))
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	// Create user
	user := domain.User{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Password:  hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Address: domain.Address{
			Street:  req.Street,
			ZipCode: req.ZipCode,
			City:    req.City,
		},
		Created: time.Now().UTC().Unix(),
		Updated: time.Now().UTC().Unix(),
	}

	// Save to database
	if err := s.store.CreateUser(ctx, user); err != nil {
		time.Sleep(time.Until(start.Add(s.minDuration)))
		return domain.User{}, err
	}

	// Ensure minimum duration
	time.Sleep(time.Until(start.Add(s.minDuration)))

	// Return user without password
	user.Password = ""
	return user, nil
}

// Authenticate validates user credentials and returns the user if valid
func (s *Service) Authenticate(ctx context.Context, email, password string) (domain.User, error) {
	user, exists := s.store.GetUserByEmail(ctx, email)
	if !exists {
		return domain.User{}, errors.New("user not found")
	}

	if err := auth.CheckPassword(user.Password, password); err != nil {
		return domain.User{}, errors.New("invalid password")
	}

	user.Password = ""
	return user, nil
}

func validateRegisterRequest(req RegisterRequest) error {
	if req.Email == "" {
		return errors.New("email is required")
	}
	if req.Password == "" {
		return errors.New("password is required")
	}
	if len(req.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	if req.FirstName == "" {
		return errors.New("firstName is required")
	}
	if req.LastName == "" {
		return errors.New("lastName is required")
	}
	if req.Street == "" {
		return errors.New("street is required")
	}
	if req.ZipCode == "" {
		return errors.New("zipCode is required")
	}
	if req.City == "" {
		return errors.New("city is required")
	}
	return nil
}
