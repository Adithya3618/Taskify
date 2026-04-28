package controller

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"

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

// GoogleLoginWithIDToken handles POST /api/auth/google/id-token
func (c *AuthController) GoogleLoginWithIDToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDToken string `json:"id_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.IDToken == "" {
		c.writeError(w, http.StatusBadRequest, "id_token is required")
		return
	}

	resp, err := c.authService.GoogleLoginWithIDToken(r.Context(), req.IDToken)
	if err != nil {
		c.handleAuthError(w, err)
		return
	}

	c.writeJSON(w, http.StatusOK, resp)
}

// GoogleLoginRedirect handles GET /api/auth/google/login
func (c *AuthController) GoogleLoginRedirect(w http.ResponseWriter, r *http.Request) {
	url, err := c.authService.GetGoogleAuthURL()
	if err != nil {
		c.handleAuthError(w, err)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback handles GET /api/auth/google/callback
func (c *AuthController) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	if state == "" || code == "" {
		c.writeError(w, http.StatusBadRequest, "state and code are required")
		return
	}

	resp, err := c.authService.GoogleLoginWithCode(r.Context(), state, code)
	if err != nil {
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:4200"
		}
		q := url.Values{}
		q.Set("error", "google_auth_failed")
		http.Redirect(w, r, frontendURL+"/auth/google/callback?"+q.Encode(), http.StatusTemporaryRedirect)
		return
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:4200"
	}
	q := url.Values{}
	q.Set("token", resp.Token)
	q.Set("name", resp.User.Name)
	q.Set("email", resp.User.Email)
	q.Set("id", resp.User.ID)
	http.Redirect(w, r, frontendURL+"/auth/google/callback?"+q.Encode(), http.StatusTemporaryRedirect)
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

// UpdateMe handles PUT /api/auth/me
func (c *AuthController) UpdateMe(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by JWT middleware)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		c.writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update user name
	user, err := c.authService.UpdateUserName(userID, req.Name)
	if err != nil {
		c.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Return updated user
	c.writeJSON(w, http.StatusOK, user)
}

// ForgotPassword handles POST /api/auth/forgot-password
func (c *AuthController) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" {
		c.writeError(w, http.StatusBadRequest, "Email is required")
		return
	}

	if err := c.authService.ForgotPassword(req.Email); err != nil {
		log.Printf("ForgotPassword error: %v", err)
		// Still return success to avoid email enumeration
	}

	c.writeJSON(w, http.StatusOK, map[string]string{
		"message": "If an account exists with this email, a verification code has been sent.",
	})
}

// VerifyOTP handles POST /api/auth/verify-otp
func (c *AuthController) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Code == "" {
		c.writeError(w, http.StatusBadRequest, "Email and code are required")
		return
	}

	resetToken, err := c.authService.VerifyResetOTP(req.Email, req.Code)
	if err != nil {
		c.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	c.writeJSON(w, http.StatusOK, map[string]string{
		"reset_token": resetToken,
	})
}

// ResetPassword handles POST /api/auth/reset-password
func (c *AuthController) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ResetToken  string `json:"reset_token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ResetToken == "" || req.NewPassword == "" {
		c.writeError(w, http.StatusBadRequest, "Reset token and new password are required")
		return
	}

	if err := c.authService.ResetPassword(req.ResetToken, req.NewPassword); err != nil {
		if errors.Is(err, services.ErrWeakPassword) {
			c.writeError(w, http.StatusBadRequest, "Password must be at least 8 characters")
			return
		}
		c.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	c.writeJSON(w, http.StatusOK, map[string]string{
		"message": "Password has been reset successfully.",
	})
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
	case errors.Is(err, services.ErrGoogleNotConfigured):
		c.writeError(w, http.StatusServiceUnavailable, "Google auth is not configured")
	case errors.Is(err, services.ErrInvalidGoogleToken):
		c.writeError(w, http.StatusUnauthorized, "Invalid Google token")
	case errors.Is(err, services.ErrGoogleEmailNotVerified):
		c.writeError(w, http.StatusUnauthorized, "Google email is not verified")
	case errors.Is(err, services.ErrInvalidOAuthState):
		c.writeError(w, http.StatusBadRequest, "Invalid OAuth state")
	case errors.Is(err, services.ErrGoogleCodeExchange):
		c.writeError(w, http.StatusBadGateway, "Failed to exchange Google auth code")
	case errors.Is(err, services.ErrGoogleProfileFetch):
		c.writeError(w, http.StatusBadGateway, "Failed to fetch Google profile")
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
	router.HandleFunc("/google/id-token", c.GoogleLoginWithIDToken).Methods("POST")
	router.HandleFunc("/google/login", c.GoogleLoginRedirect).Methods("GET")
	router.HandleFunc("/google/callback", c.GoogleCallback).Methods("GET")
	router.HandleFunc("/me", c.GetMe).Methods("GET")
	router.HandleFunc("/me", c.UpdateMe).Methods("PUT")
	router.HandleFunc("/forgot-password", c.ForgotPassword).Methods("POST")
	router.HandleFunc("/verify-otp", c.VerifyOTP).Methods("POST")
	router.HandleFunc("/reset-password", c.ResetPassword).Methods("POST")
}
