package users

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/willie68/schematics2/backend/internal/auth"
	"github.com/willie68/schematics2/backend/internal/domain/model"
)

// ServiceTestSuite tests the user service
type ServiceTestSuite struct {
	suite.Suite
	inj   do.Injector
	store *mockuserStore
	svc   *Service
}

func (s *ServiceTestSuite) SetupTest() {
	s.inj = do.New()
	s.store = newMockuserStore(s.T())
	do.ProvideValue(s.inj, s.store)
	s.svc = NewService(s.inj, 100*time.Millisecond) // Use small duration for tests
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (s *ServiceTestSuite) TestRegister_Success() {
	// GIVEN
	req := RegisterRequest{
		Email:     "user@example.com",
		Password:  "securePassword123",
		FirstName: "John",
		LastName:  "Doe",
		Street:    "123 Main St",
		ZipCode:   "12345",
		City:      "TestCity",
	}

	s.store.EXPECT().CreateUser(mock.Anything, mock.Anything).Return(nil)

	// WHEN
	user, err := s.svc.Register(context.Background(), req)

	// THEN
	s.Require().NoError(err, "registration should succeed")
	s.Assert().Equal("user@example.com", user.Email)
	s.Assert().Equal("John", user.FirstName)
	s.Assert().Equal("Doe", user.LastName)
	s.Assert().NotEmpty(user.ID)
	s.Assert().Empty(user.Password, "password should not be returned")
	s.Assert().NotZero(user.Created)
}

func (s *ServiceTestSuite) TestRegister_DuplicateEmail() {
	// GIVEN
	req := RegisterRequest{
		Email:     "user@example.com",
		Password:  "securePassword123",
		FirstName: "John",
		LastName:  "Doe",
		Street:    "123 Main St",
		ZipCode:   "12345",
		City:      "TestCity",
	}

	s.store.EXPECT().CreateUser(mock.Anything, mock.Anything).Return(errors.New("email already exists"))

	// WHEN - Try to register same email
	_, err := s.svc.Register(context.Background(), req)

	// THEN
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "email already exists")
}

func (s *ServiceTestSuite) TestRegister_MissingEmail() {
	// GIVEN
	req := RegisterRequest{
		Email:     "", // empty email
		Password:  "securePassword123",
		FirstName: "John",
		LastName:  "Doe",
		Street:    "123 Main St",
		ZipCode:   "12345",
		City:      "TestCity",
	}

	// WHEN
	_, err := s.svc.Register(context.Background(), req)

	// THEN
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "email is required")
}

func (s *ServiceTestSuite) TestRegister_PasswordTooShort() {
	// GIVEN
	req := RegisterRequest{
		Email:     "user@example.com",
		Password:  "short",
		FirstName: "John",
		LastName:  "Doe",
		Street:    "123 Main St",
		ZipCode:   "12345",
		City:      "TestCity",
	}

	// WHEN
	_, err := s.svc.Register(context.Background(), req)

	// THEN
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "at least 8 characters")
}

func (s *ServiceTestSuite) TestRegister_MissingFirstName() {
	// GIVEN
	req := RegisterRequest{
		Email:     "user@example.com",
		Password:  "securePassword123",
		FirstName: "", // empty first name
		LastName:  "Doe",
		Street:    "123 Main St",
		ZipCode:   "12345",
		City:      "TestCity",
	}

	// WHEN
	_, err := s.svc.Register(context.Background(), req)

	// THEN
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "firstName is required")
}

func (s *ServiceTestSuite) TestRegister_MissingLastName() {
	// GIVEN
	req := RegisterRequest{
		Email:     "user@example.com",
		Password:  "securePassword123",
		FirstName: "John",
		LastName:  "", // empty last name
		Street:    "123 Main St",
		ZipCode:   "12345",
		City:      "TestCity",
	}

	// WHEN
	_, err := s.svc.Register(context.Background(), req)

	// THEN
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "lastName is required")
}

func (s *ServiceTestSuite) TestRegister_MissingStreet() {
	// GIVEN
	req := RegisterRequest{
		Email:     "user@example.com",
		Password:  "securePassword123",
		FirstName: "John",
		LastName:  "Doe",
		Street:    "", // empty street
		ZipCode:   "12345",
		City:      "TestCity",
	}

	// WHEN
	_, err := s.svc.Register(context.Background(), req)

	// THEN
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "street is required")
}

func (s *ServiceTestSuite) TestRegister_MissingZipCode() {
	// GIVEN
	req := RegisterRequest{
		Email:     "user@example.com",
		Password:  "securePassword123",
		FirstName: "John",
		LastName:  "Doe",
		Street:    "123 Main St",
		ZipCode:   "", // empty zip code
		City:      "TestCity",
	}

	// WHEN
	_, err := s.svc.Register(context.Background(), req)

	// THEN
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "zipCode is required")
}

func (s *ServiceTestSuite) TestRegister_MissingCity() {
	// GIVEN
	req := RegisterRequest{
		Email:     "user@example.com",
		Password:  "securePassword123",
		FirstName: "John",
		LastName:  "Doe",
		Street:    "123 Main St",
		ZipCode:   "12345",
		City:      "", // empty city
	}

	// WHEN
	_, err := s.svc.Register(context.Background(), req)

	// THEN
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "city is required")
}

func (s *ServiceTestSuite) TestRegister_MinimumDuration() {
	// GIVEN
	req := RegisterRequest{
		Email:     "user@example.com",
		Password:  "securePassword123",
		FirstName: "John",
		LastName:  "Doe",
		Street:    "123 Main St",
		ZipCode:   "12345",
		City:      "TestCity",
	}

	s.store.EXPECT().CreateUser(mock.Anything, mock.Anything).Return(nil)
	// WHEN
	start := time.Now()
	_, err := s.svc.Register(context.Background(), req)
	elapsed := time.Since(start)

	// THEN
	s.Require().NoError(err)
	// Should have taken at least 100ms (the minDuration we set in SetupTest)
	s.Assert().GreaterOrEqual(elapsed, 100*time.Millisecond, "registration should take at least minDuration")
}

func (s *ServiceTestSuite) TestAuthenticate_Success() {
	// GIVEN
	pwd, err := auth.HashPassword("securePassword123")
	s.NoError(err)
	s.store.EXPECT().GetUserByEmail(mock.Anything, "user@example.com").Return(model.User{
		Email:    "user@example.com",
		Password: pwd,
	}, true)

	// WHEN
	user, err := s.svc.Authenticate(context.Background(), "user@example.com", "securePassword123")

	// THEN
	s.Require().NoError(err)
	s.Assert().Equal("user@example.com", user.Email)
	s.Assert().Empty(user.Password, "password should not be returned")
}

func (s *ServiceTestSuite) TestAuthenticate_InvalidPassword() {
	// GIVEN
	s.store.EXPECT().GetUserByEmail(mock.Anything, "user@example.com").Return(model.User{
		Email:    "user@example.com",
		Password: "securePassword123",
	}, true)

	// WHEN
	_, err := s.svc.Authenticate(context.Background(), "user@example.com", "wrongPassword")

	// THEN
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "invalid password")
}

func (s *ServiceTestSuite) TestAuthenticate_UserNotFound() {
	s.store.EXPECT().GetUserByEmail(mock.Anything, "nonexistent@example.com").Return(model.User{}, false)

	// WHEN
	_, err := s.svc.Authenticate(context.Background(), "nonexistent@example.com", "password")

	// THEN
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "user not found")
}

// TestPasswordHashing verifies password hashing works correctly
func TestPasswordHashing(t *testing.T) {
	// GIVEN
	plainPassword := "mySecurePassword123"

	// WHEN
	hash, err := auth.HashPassword(plainPassword)

	// THEN
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, plainPassword, hash)

	// Verify password matches hash
	err = auth.CheckPassword(hash, plainPassword)
	assert.NoError(t, err)

	// Verify wrong password fails
	err = auth.CheckPassword(hash, "wrongPassword")
	assert.Error(t, err)
}
