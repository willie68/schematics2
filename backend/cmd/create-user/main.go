package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/samber/do/v2"
	"github.com/willie68/schematic2/backend/internal"
	"github.com/willie68/schematic2/backend/internal/config"
	"github.com/willie68/schematic2/backend/internal/services/users"
)

func main() {
	// Define command-line flags
	email := flag.String("email", "", "User email address (required)")
	password := flag.String("password", "", "User password (required)")
	firstName := flag.String("firstName", "", "User first name")
	lastName := flag.String("lastName", "", "User last name")
	street := flag.String("street", "", "Address street")
	zipCode := flag.String("zipCode", "", "Address zip code")
	city := flag.String("city", "", "Address city")
	dryRun := flag.Bool("dry-run", false, "Print registration request without saving to database")

	flag.Parse()

	// Validate required fields
	if *email == "" {
		log.Fatal("Error: -email flag is required")
	}
	if *password == "" {
		log.Fatal("Error: -password flag is required")
	}

	// Create registration request
	req := users.RegisterRequest{
		Email:     *email,
		Password:  *password,
		FirstName: *firstName,
		LastName:  *lastName,
		Street:    *street,
		ZipCode:   *zipCode,
		City:      *city,
	}

	// Validate the request early (will catch password length < 8, etc.)
	if err := validateRegisterRequest(req); err != nil {
		log.Fatalf("Validation error: %v", err)
	}

	log.Printf("Registering user: %s (%s %s)", req.Email, req.FirstName, req.LastName)

	// Exit early if dry-run
	if *dryRun {
		log.Println("Dry-run mode: User not saved to database")
		return
	}

	// Load config
	cfg := config.LoadFromEnv()

	// Initialize DI container with all services
	inj := do.New()
	err := internal.InitServices(inj, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize services: %v", err)
	}
	defer internal.ShutdownServices(inj)

	// Get user service (handles password hashing automatically)
	userSvc := do.MustInvoke[*users.Service](inj)

	// Register user with automatic password hashing and validation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := userSvc.Register(ctx, req)
	if err != nil {
		log.Fatalf("Failed to register user: %v", err)
	}

	log.Printf("User created successfully!")
	log.Printf("ID: %s", user.ID)
	log.Printf("Email: %s", user.Email)
	log.Printf("Name: %s %s", user.FirstName, user.LastName)
	log.Printf("Address: %s, %s %s", user.Address.Street, user.Address.ZipCode, user.Address.City)
	log.Printf("Created: %v", time.Unix(user.Created, 0))
}

func validateRegisterRequest(req users.RegisterRequest) error {
	// This mirrors the validation in users/service.go to catch errors early
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	if req.FirstName == "" {
		return fmt.Errorf("firstName is required")
	}
	if req.LastName == "" {
		return fmt.Errorf("lastName is required")
	}
	if req.Street == "" {
		return fmt.Errorf("street is required")
	}
	if req.ZipCode == "" {
		return fmt.Errorf("zipCode is required")
	}
	if req.City == "" {
		return fmt.Errorf("city is required")
	}
	return nil
}
