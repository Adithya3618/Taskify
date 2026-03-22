# Test Cases - Backend Unit Tests

This folder contains all unit tests for the Taskify backend, organized by functionality.

## Test Files

| File | Description | Test Count |
|------|-------------|------------|
| `jwt_service_test.go` | JWT token generation and validation | 8 tests |
| `otp_service_test.go` | OTP generation and verification | 6 tests |
| `auth_service_test.go` | Email validation, normalization, auth structures | 12 tests |
| `auth_controller_test.go` | HTTP request handling for auth endpoints | 11 tests |
| `middleware_test.go` | JWT middleware, context helpers | 10 tests |
| `models_test.go` | User, Project, Stage, Task, Message models | 13 tests |
| `controllers_test.go` | Project, Stage, Task, Message controller authorization | 14 tests |
| `services_test.go` | Service constructor tests | 4 tests |
| `user_repository_test.go` | Database operations for users | 11 tests |
| `database_test.go` | Database connection and SQL operations | 6 tests |

**Total: ~95 unit tests**

## Running Tests

### Using Shell Scripts (Recommended)
```bash
# Make scripts executable (if not already)
chmod +x alltestcase.sh run_all_tests.sh

# Run all tests with coverage
./alltestcase.sh
# or
./run_all_tests.sh
```

### Using Go Commands
```bash
# Run all tests in the testcases package
go test -v ./internal/testcases/...

# Run tests with coverage
go test -cover ./internal/testcases/...

# Run specific test file
go test -v ./internal/testcases/... -run "TestJWT"
```

## Test Categories

### Authentication Tests
- JWT token generation and validation
- OTP service (generation, verification, reset tokens)
- Email validation and normalization
- Auth service errors

### Controller Tests
- Request validation (invalid JSON, missing fields)
- Authorization checks (401 for missing user context)
- Error response handling

### Model Tests
- Structure validation
- JSON tag verification
- Zero value handling

### Repository Tests
- Database CRUD operations
- User lookup by email/ID
- Password updates

### Database Tests
- Connection handling
- SQL operations (CREATE, INSERT, SELECT)
- Transaction support
