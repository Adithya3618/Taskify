package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/helpers"
	"backend/internal/services"

	"github.com/gorilla/mux"
)

// LabelController handles HTTP requests for labels
type LabelController struct {
	labelService *services.LabelService
}

// NewLabelController creates a new LabelController
func NewLabelController(labelService *services.LabelService) *LabelController {
	return &LabelController{labelService: labelService}
}

// createLabelRequest represents the request body for creating a label
type createLabelRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// CreateLabel handles POST /api/projects/:id/labels
func (c *LabelController) CreateLabel(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	projectID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var req createLabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Label name is required", http.StatusBadRequest)
		return
	}

	// Use userID as userName for activity logs (can be enhanced later to fetch actual name)
	userName := userID

	label, err := c.labelService.CreateLabel(projectID, req.Name, req.Color, userID, userName)
	if err != nil {
		if svcErr, ok := err.(*services.ServiceError); ok {
			http.Error(w, svcErr.Message, http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    label,
	})
}

// GetLabels handles GET /api/projects/:id/labels
func (c *LabelController) GetLabels(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	projectID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	labels, err := c.labelService.GetProjectLabels(projectID, userID)
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

// DeleteLabel handles DELETE /api/labels/:id
func (c *LabelController) DeleteLabel(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	labelID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid label ID", http.StatusBadRequest)
		return
	}

	userName := userID

	err = c.labelService.DeleteLabel(labelID, userID, userName)
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
		"message": "Label deleted successfully",
	})
}
