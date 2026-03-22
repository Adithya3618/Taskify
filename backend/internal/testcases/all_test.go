package testcases

import (
	"testing"
)

// AllTestCases runs all test functions in the testcases package
// This file serves as an aggregator for running all tests at once
func TestAllTestCases(t *testing.T) {
	// This test serves as documentation that all tests in this package
	// can be run together with: go test -v ./internal/testcases/...
	//
	// Individual test files:
	// - jwt_service_test.go: JWT token tests
	// - otp_service_test.go: OTP generation/verification tests
	// - auth_service_test.go: Auth service validation tests
	// - auth_controller_test.go: Auth controller HTTP tests
	// - middleware_test.go: JWT middleware tests
	// - models_test.go: Data model tests
	// - controllers_test.go: Controller authorization tests
	// - services_test.go: Service constructor tests
	// - user_repository_test.go: User repository tests
	// - database_test.go: Database connection tests

	// Run all sub-tests
	t.Run("JWTService", func(t *testing.T) {
		TestJWTService_GenerateToken(t)
	})

	t.Run("OTPService", func(t *testing.T) {
		TestOTPService_GenerateOTP(t)
	})

	t.Run("AuthService", func(t *testing.T) {
		TestIsValidEmail(t)
	})

	t.Run("Models", func(t *testing.T) {
		TestUser_ToResponse(t)
	})

	t.Run("Middleware", func(t *testing.T) {
		TestGetUserID(t)
	})

	t.Run("Controllers", func(t *testing.T) {
		TestProjectController_CreateProject_Unauthorized(t)
	})

	t.Run("Services", func(t *testing.T) {
		TestNewProjectService(t)
	})

	t.Run("UserRepository", func(t *testing.T) {
		TestNewUserRepository(t)
	})

	t.Run("Database", func(t *testing.T) {
		TestNewDB_Success(t)
	})
}
