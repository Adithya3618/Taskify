package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"backend/internal/helpers"
	"backend/internal/services"

	"github.com/gorilla/mux"
)

// ProjectMemberController handles project member operations
type ProjectMemberController struct {
	service *services.ProjectMemberService
}

// NewProjectMemberController creates a new ProjectMemberController
func NewProjectMemberController(service *services.ProjectMemberService) *ProjectMemberController {
	return &ProjectMemberController{service: service}
}

// AddMember handles POST /api/projects/:id/members
func (c *ProjectMemberController) AddMember(w http.ResponseWriter, r *http.Request) {
	// Get current user from context (set by JWT middleware)
	invitedBy := getUserIDFromContext(r)
	if invitedBy == "" {
		helpers.WriteError(w, http.StatusUnauthorized, "Authentication required", helpers.ErrCodeUnauthorized)
		return
	}

	// Get project ID from URL
	vars := mux.Vars(r)
	projectID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "Invalid project ID", helpers.ErrCodeBadRequest)
		return
	}

	// Parse request body
	var req struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "Invalid request body", helpers.ErrCodeBadRequest)
		return
	}

	// Validate user_id or email is provided
	if req.UserID == "" && req.Email == "" {
		helpers.WriteError(w, http.StatusBadRequest, "user_id or email is required", helpers.ErrCodeValidationFailed)
		return
	}

	// Add member
	member, err := c.service.AddMember(projectID, req.UserID, invitedBy)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	helpers.WriteSuccess(w, http.StatusCreated, member, "Member added successfully")
}

// RemoveMember handles DELETE /api/projects/:id/members/:userId
func (c *ProjectMemberController) RemoveMember(w http.ResponseWriter, r *http.Request) {
	// Get current user from context (set by JWT middleware)
	removedBy := getUserIDFromContext(r)
	if removedBy == "" {
		helpers.WriteError(w, http.StatusUnauthorized, "Authentication required", helpers.ErrCodeUnauthorized)
		return
	}

	// Get project ID from URL
	vars := mux.Vars(r)
	projectID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "Invalid project ID", helpers.ErrCodeBadRequest)
		return
	}

	// Get target user ID from URL
	targetUserID := vars["userId"]
	if targetUserID == "" {
		helpers.WriteError(w, http.StatusBadRequest, "User ID is required", helpers.ErrCodeBadRequest)
		return
	}

	// Remove member
	err = c.service.RemoveMember(projectID, targetUserID, removedBy)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	helpers.WriteSuccess(w, http.StatusOK, nil, "Member removed successfully")
}

// GetMembers handles GET /api/projects/:id/members
func (c *ProjectMemberController) GetMembers(w http.ResponseWriter, r *http.Request) {
	// Get current user from context (set by JWT middleware)
	requesterID := getUserIDFromContext(r)
	if requesterID == "" {
		helpers.WriteError(w, http.StatusUnauthorized, "Authentication required", helpers.ErrCodeUnauthorized)
		return
	}

	// Get project ID from URL
	vars := mux.Vars(r)
	projectID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "Invalid project ID", helpers.ErrCodeBadRequest)
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	// Get members
	result, err := c.service.GetMembers(projectID, requesterID, page, limit)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	helpers.WritePaginated(w, http.StatusOK, result.Members, result.Page, result.Limit, result.Total)
}

// handleServiceError converts service errors to HTTP responses
func handleServiceError(w http.ResponseWriter, err error) {
	if se, ok := services.IsServiceError(err); ok {
		switch se.Code {
		case "PROJECT_NOT_FOUND", "USER_NOT_FOUND":
			helpers.WriteError(w, http.StatusNotFound, se.Message, helpers.ErrCodeNotFound)
		case "ACCESS_DENIED":
			helpers.WriteError(w, http.StatusForbidden, se.Message, helpers.ErrCodeForbidden)
		case "INVALID_REQUEST":
			helpers.WriteError(w, http.StatusBadRequest, se.Message, helpers.ErrCodeBadRequest)
		default:
			helpers.WriteError(w, http.StatusInternalServerError, se.Message, helpers.ErrCodeInternalError)
		}
		return
	}

	// Handle other errors
	errMsg := err.Error()
	if strings.Contains(errMsg, "already a member") {
		helpers.WriteError(w, http.StatusConflict, errMsg, helpers.ErrCodeConflict)
		return
	}
	if strings.Contains(errMsg, "not found") {
		helpers.WriteError(w, http.StatusNotFound, errMsg, helpers.ErrCodeNotFound)
		return
	}
	if strings.Contains(errMsg, "cannot remove") {
		helpers.WriteError(w, http.StatusForbidden, errMsg, helpers.ErrCodeForbidden)
		return
	}

	helpers.WriteError(w, http.StatusInternalServerError, "Internal server error", helpers.ErrCodeInternalError)
}

// getUserIDFromContext retrieves the user ID from the request context
func getUserIDFromContext(r *http.Request) string {
	if userID, ok := r.Context().Value("user_id").(string); ok {
		return userID
	}
	return ""
}
