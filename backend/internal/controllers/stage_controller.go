package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/helpers"
	"backend/internal/services"

	"github.com/gorilla/mux"
)

type StageController struct {
	service *services.StageService
}

func NewStageController(service *services.StageService) *StageController {
	return &StageController{service: service}
}

// getUserID extracts user_id from request context
func getUserID(r *http.Request) string {
	if userID, ok := r.Context().Value("user_id").(string); ok {
		return userID
	}
	return ""
}

// CreateStage handles POST /api/projects/:projectId/stages
func (c *StageController) CreateStage(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	projectID, err := strconv.ParseInt(vars["projectId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name     string `json:"name"`
		Position int    `json:"position"`
		IsFinal  *bool  `json:"is_final"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Stage name is required", http.StatusBadRequest)
		return
	}

	isFinal := 0
	if req.IsFinal != nil && *req.IsFinal {
		isFinal = 1
	}

	stage, err := c.service.CreateStage(userID, projectID, req.Name, req.Position, isFinal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(stage)
}

// GetStagesByProject handles GET /api/projects/:projectId/stages
func (c *StageController) GetStagesByProject(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	projectID, err := strconv.ParseInt(vars["projectId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	stages, err := c.service.GetStagesByProject(userID, projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stages)
}

// ReorderStages handles PUT /api/projects/:id/stages/reorder
func (c *StageController) ReorderStages(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		StageIDs []int64 `json:"stage_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	stages, err := c.service.ReorderStages(userID, projectID, req.StageIDs)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stages)
}

// GetStage handles GET /api/stages/:id
func (c *StageController) GetStage(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}

	stage, err := c.service.GetStageByID(userID, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if stage == nil {
		http.Error(w, "Stage not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stage)
}

// UpdateStage handles PUT /api/stages/:id
func (c *StageController) UpdateStage(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name     string `json:"name"`
		Position int    `json:"position"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	stage, err := c.service.UpdateStage(userID, id, req.Name, req.Position)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if stage == nil {
		http.Error(w, "Stage not found or access denied", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stage)
}

// DeleteStage handles DELETE /api/stages/:id
func (c *StageController) DeleteStage(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}

	if err := c.service.DeleteStage(userID, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
