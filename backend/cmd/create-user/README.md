# Create User Tool

Command-line tool to create a new user in the schematics2 system.

## Usage

```bash
go run ./cmd/create-user/main.go -email user@example.com -password secret123 -firstName John -lastName Doe -street "123 Main St" -zipCode "12345" -city "Springfield"
```

## Flags

### Required Flags
- `-email string` - User email address (required)
- `-password string` - User password (required)

### Optional Flags
- `-firstName string` - User first name
- `-lastName string` - User last name
- `-street string` - Address street
- `-zipCode string` - Address zip code
- `-city string` - Address city
- `-dry-run` - Print user as JSON without saving to database

## Examples

### Create a minimal user (dry-run)
```bash
go run ./cmd/create-user/main.go -email test@example.com -password pass123 -dry-run
```

### Create a user with complete information
```bash
go run ./cmd/create-user/main.go \
  -email john.doe@example.com \
  -password securePassword123 \
  -firstName John \
  -lastName Doe \
  -street "456 Oak Avenue" \
  -zipCode "54321" \
  -city "Shelbyville"
```

### Create a user and save to database
Ensure MongoDB connection is configured via environment variables (see config.go).

```bash
go run ./cmd/create-user/main.go \
  -email admin@example.com \
  -password adminPass \
  -firstName Admin \
  -lastName User
```

## Fields

| Field | Required | Notes |
|-------|----------|-------|
| ID | No | Auto-generated UUID |
| Email | Yes | - |
| Password | Yes | - |
| FirstName | No | - |
| LastName | No | - |
| Address.Street | No | - |
| Address.ZipCode | No | - |
| Address.City | No | - |
| Created | No | Auto-set to current time |
| Updated | No | Auto-set to current time |
