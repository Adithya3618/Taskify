package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"backend/internal/helpers"
	"backend/internal/models"
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
		p, err := strconv.Atoi(page)
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "Invalid page parameter", helpers.ErrCodeBadRequest)
			return
		}
		params.Page = p
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "Invalid limit parameter", helpers.ErrCodeBadRequest)
			return
		}
		params.Limit = l
	}
	params.Page, params.Limit = normalizeActivityPagination(params.Page, params.Limit)

	if userID := r.URL.Query().Get("user_id"); userID != "" {
		params.UserID = userID
	}

	if from := r.URL.Query().Get("from"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "Invalid from date format", helpers.ErrCodeBadRequest)
			return
		}
		params.From = &t
	}

	if to := r.URL.Query().Get("to"); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "Invalid to date format", helpers.ErrCodeBadRequest)
			return
		}
		params.To = &t
	}

	// Get activity logs
	logs, total, err := c.service.GetProjectActivity(projectID, requesterID, params)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.ActivityFeedResponse{
		Logs:  toActivityFeedLogs(logs),
		Total: total,
		Page:  params.Page,
	})
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
		parsedLimit, err := strconv.Atoi(l)
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "Invalid limit parameter", helpers.ErrCodeBadRequest)
			return
		}
		limit = parsedLimit
	}
	_, limit = normalizeActivityPagination(1, limit)

	// Get recent activity
	logs, err := c.service.GetRecentActivity(projectID, requesterID, limit)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	helpers.WriteSuccess(w, http.StatusOK, logs, "")
}

func toActivityFeedLogs(logs []models.ActivityLogResponse) []models.ActivityFeedLog {
	feedLogs := make([]models.ActivityFeedLog, 0, len(logs))
	for _, log := range logs {
		feedLogs = append(feedLogs, models.ActivityFeedLog{
			ID:          log.ID,
			UserName:    log.UserName,
			Action:      log.Action,
			EntityType:  log.EntityType,
			EntityTitle: log.Description,
			CreatedAt:   log.CreatedAt,
		})
	}
	return feedLogs
}

func normalizeActivityPagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}
