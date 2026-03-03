package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"backend/internal/auth/services"

	"github.com/gorilla/mux"
)

// AuthController handles HTTP requests for authentication
type AuthController struct {
	authService *services.AuthService
}

// NewAuthController creates a new AuthController
func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// Register handles POST /api/auth/register
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	// Decode request body
	var req services.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Perform registration
	resp, err := c.authService.Register(req)
	if err != nil {
		c.handleAuthError(w, err)
		return
	}

	// Success response
	c.writeJSON(w, http.StatusCreated, resp)
}

// Login handles POST /api/auth/login
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	// Decode request body
	var req services.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Perform login
	resp, err := c.authService.Login(req)
	if err != nil {
		c.handleAuthError(w, err)
		return
	}

	// Success response
	c.writeJSON(w, http.StatusOK, resp)
}

// GetMe handles GET /api/auth/me
func (c *AuthController) GetMe(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by JWT middleware)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		c.writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get user from database
	user, err := c.authService.GetUserByID(userID)
	if err != nil {
		c.writeError(w, http.StatusNotFound, "User not found")
		return
	}

	// Return user response (excludes password)
	c.writeJSON(w, http.StatusOK, user.ToResponse())
}

// handleAuthError maps authentication errors to HTTP status codes
func (c *AuthController) handleAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrUserExists):
		c.writeError(w, http.StatusConflict, "An account with this email already exists")
	case errors.Is(err, services.ErrInvalidCredentials):
		c.writeError(w, http.StatusUnauthorized, "Invalid email or password")
	case errors.Is(err, services.ErrWeakPassword):
		c.writeError(w, http.StatusBadRequest, "Password must be at least 8 characters")
	case errors.Is(err, services.ErrInvalidEmail):
		c.writeError(w, http.StatusBadRequest, "Invalid email format")
	case errors.Is(err, services.ErrUserInactive):
		c.writeError(w, http.StatusForbidden, "User account is inactive")
	default:
		c.writeError(w, http.StatusInternalServerError, "An error occurred")
	}
}

// writeJSON writes a JSON response
func (c *AuthController) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log error in production
	}
}

// writeError writes an error response
func (c *AuthController) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// RegisterRoutes registers auth routes on the router
func (c *AuthController) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/register", c.Register).Methods("POST")
	router.HandleFunc("/login", c.Login).Methods("POST")
	router.HandleFunc("/me", c.GetMe).Methods("GET")
}
