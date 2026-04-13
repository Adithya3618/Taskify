package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/helpers"
	"backend/internal/services"

	"github.com/gorilla/mux"
)

// TaskLabelController handles HTTP requests for task-label associations
type TaskLabelController struct {
	taskLabelService *services.TaskLabelService
}

// NewTaskLabelController creates a new TaskLabelController
func NewTaskLabelController(taskLabelService *services.TaskLabelService) *TaskLabelController {
	return &TaskLabelController{taskLabelService: taskLabelService}
}

// assignLabelRequest represents the request body for assigning a label
type assignLabelRequest struct {
	LabelID int64 `json:"label_id"`
}

// AssignLabel handles POST /api/tasks/:id/labels
func (c *TaskLabelController) AssignLabel(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	taskID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var req assignLabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.LabelID == 0 {
		http.Error(w, "Label ID is required", http.StatusBadRequest)
		return
	}

	userName := userID

	err = c.taskLabelService.AssignLabel(taskID, req.LabelID, userID, userName)
	if err != nil {
		if svcErr, ok := err.(*services.ServiceError); ok {
			http.Error(w, svcErr.Message, http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Label assigned successfully",
	})
}

// RemoveLabel handles DELETE /api/tasks/:id/labels/:labelId
func (c *TaskLabelController) RemoveLabel(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	taskID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	labelID, err := strconv.ParseInt(vars["labelId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid label ID", http.StatusBadRequest)
		return
	}

	userName := userID

	err = c.taskLabelService.RemoveLabel(taskID, labelID, userID, userName)
	if err != nil {
		if svcErr, ok := err.(*services.ServiceError); ok {
			http.Error(w, svcErr.Message, http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Label removed successfully",
	})
}

// GetTaskLabels handles GET /api/tasks/:id/labels
func (c *TaskLabelController) GetTaskLabels(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	taskID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	labels, err := c.taskLabelService.GetTaskLabels(taskID, userID)
	if err != nil {
		if svcErr, ok := err.(*services.ServiceError); ok {
			http.Error(w, svcErr.Message, http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    labels,
	})
}
