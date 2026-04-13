package controllers

import (
	"net/http"
	"strconv"
	"time"

	"backend/internal/helpers"
	"backend/internal/repository"
	"backend/internal/services"

	"github.com/gorilla/mux"
)

// ActivityController handles activity log operations
type ActivityController struct {
	service *services.ActivityService
}

// NewActivityController creates a new ActivityController
func NewActivityController(service *services.ActivityService) *ActivityController {
	return &ActivityController{service: service}
}

// GetActivity handles GET /api/projects/:id/activity
func (c *ActivityController) GetActivity(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
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

	// Parse query parameters
	params := repository.ActivityLogParams{
		Page:  1,
		Limit: 20,
	}

	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			params.Page = p
		}
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			params.Limit = l
		}
	}

	if userID := r.URL.Query().Get("user_id"); userID != "" {
		params.UserID = userID
	}

	if from := r.URL.Query().Get("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			params.From = &t
		}
	}

	if to := r.URL.Query().Get("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			params.To = &t
		}
	}

	// Get activity logs
	logs, total, err := c.service.GetProjectActivity(projectID, requesterID, params)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	helpers.WritePaginated(w, http.StatusOK, logs, params.Page, params.Limit, total)
}

// GetRecentActivity handles GET /api/projects/:id/activity/recent
func (c *ActivityController) GetRecentActivity(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
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

	// Parse limit parameter
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get recent activity
	logs, err := c.service.GetRecentActivity(projectID, requesterID, limit)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	helpers.WriteSuccess(w, http.StatusOK, logs, "")
}
