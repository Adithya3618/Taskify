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

// NewProjectMemberController initializes controller
func NewProjectMemberController(service *services.ProjectMemberService) *ProjectMemberController {
	return &ProjectMemberController{service: service}
}

// CreateInvite handles POST /api/projects/:id/invites
func (c *ProjectMemberController) CreateInvite(w http.ResponseWriter, r *http.Request) {
	currentUserID := getUserIDFromContext(r)
	if currentUserID == "" {
		helpers.WriteError(w, http.StatusUnauthorized, "Authentication required", helpers.ErrCodeUnauthorized)
		return
	}

	projectID, err := getProjectID(r)
	if err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "Invalid project ID", helpers.ErrCodeBadRequest)
		return
	}

	var req struct {
		ExpiresInHours int `json:"expires_in_hours"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	// Default expiry: 7 days
	if req.ExpiresInHours <= 0 {
		req.ExpiresInHours = 168 // 7 days
	}

	invite, err := c.service.CreateInvite(projectID, currentUserID, req.ExpiresInHours)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	helpers.WriteSuccess(w, http.StatusCreated, invite, "Invite created successfully")
}

// AcceptInvite handles POST /api/invites/:id/accept
func (c *ProjectMemberController) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	currentUserID := getUserIDFromContext(r)
	if currentUserID == "" {
		helpers.WriteError(w, http.StatusUnauthorized, "Authentication required", helpers.ErrCodeUnauthorized)
		return
	}

	vars := mux.Vars(r)
	inviteID := vars["id"]
	if inviteID == "" {
		helpers.WriteError(w, http.StatusBadRequest, "Invite ID required", helpers.ErrCodeBadRequest)
		return
	}

	invite, err := c.service.AcceptInviteByID(inviteID, currentUserID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	helpers.WriteSuccess(w, http.StatusOK, invite, "Invite accepted successfully")
}

// GetInvite handles GET /api/invites/:id
func (c *ProjectMemberController) GetInvite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	inviteID := vars["id"]
	if inviteID == "" {
		helpers.WriteError(w, http.StatusBadRequest, "Invite ID required", helpers.ErrCodeBadRequest)
		return
	}

	invite, err := c.service.GetInvite(inviteID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	helpers.WriteSuccess(w, http.StatusOK, invite, "")
}

// AddMember handles POST /api/projects/:id/members
func (c *ProjectMemberController) AddMember(w http.ResponseWriter, r *http.Request) {
	currentUserID := getUserIDFromContext(r)
	if currentUserID == "" {
		helpers.WriteError(w, http.StatusUnauthorized, "Authentication required", helpers.ErrCodeUnauthorized)
		return
	}

	projectID, err := getProjectID(r)
	if err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "Invalid project ID", helpers.ErrCodeBadRequest)
		return
	}

	var req struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "Invalid request body", helpers.ErrCodeBadRequest)
		return
	}

	if req.UserID == "" && req.Email == "" {
		helpers.WriteError(w, http.StatusBadRequest, "user_id or email is required", helpers.ErrCodeValidationFailed)
		return
	}

	member, err := c.service.AddMember(projectID, req.UserID, currentUserID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	helpers.WriteSuccess(w, http.StatusCreated, member, "Member added successfully")
}

// RemoveMember handles DELETE /api/projects/:id/members/:userId
func (c *ProjectMemberController) RemoveMember(w http.ResponseWriter, r *http.Request) {
	currentUserID := getUserIDFromContext(r)
	if currentUserID == "" {
		helpers.WriteError(w, http.StatusUnauthorized, "Authentication required", helpers.ErrCodeUnauthorized)
		return
	}

	projectID, err := getProjectID(r)
	if err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "Invalid project ID", helpers.ErrCodeBadRequest)
		return
	}

	targetUserID := mux.Vars(r)["userId"]
	if targetUserID == "" {
		helpers.WriteError(w, http.StatusBadRequest, "User ID is required", helpers.ErrCodeBadRequest)
		return
	}

	if err := c.service.RemoveMember(projectID, targetUserID, currentUserID); err != nil {
		handleServiceError(w, err)
		return
	}

	helpers.WriteSuccess(w, http.StatusOK, nil, "Member removed successfully")
}

// GetMembers handles GET /api/projects/:id/members
func (c *ProjectMemberController) GetMembers(w http.ResponseWriter, r *http.Request) {
	currentUserID := getUserIDFromContext(r)
	if currentUserID == "" {
		helpers.WriteError(w, http.StatusUnauthorized, "Authentication required", helpers.ErrCodeUnauthorized)
		return
	}

	projectID, err := getProjectID(r)
	if err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "Invalid project ID", helpers.ErrCodeBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	result, err := c.service.GetMembers(projectID, currentUserID, page, limit)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	helpers.WritePaginated(w, http.StatusOK, result.Members, result.Page, result.Limit, result.Total)
}

// -------------------- Helpers --------------------

// getProjectID extracts and validates project ID
func getProjectID(r *http.Request) (int64, error) {
	vars := mux.Vars(r)
	return strconv.ParseInt(vars["id"], 10, 64)
}

// handleServiceError maps service errors to HTTP responses
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

	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "already a member"):
		helpers.WriteError(w, http.StatusConflict, errMsg, helpers.ErrCodeConflict)
	case strings.Contains(errMsg, "not found"):
		helpers.WriteError(w, http.StatusNotFound, errMsg, helpers.ErrCodeNotFound)
	case strings.Contains(errMsg, "cannot remove"):
		helpers.WriteError(w, http.StatusForbidden, errMsg, helpers.ErrCodeForbidden)
	default:
		helpers.WriteError(w, http.StatusInternalServerError, "Internal server error", helpers.ErrCodeInternalError)
	}
}

// getUserIDFromContext retrieves user ID from request context
func getUserIDFromContext(r *http.Request) string {
	if userID, ok := r.Context().Value("user_id").(string); ok {
		return userID
	}
	return ""
}
