package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/services"

	"github.com/gorilla/mux"
)

type StageController struct {
	service *services.StageService
}

func NewStageController(service *services.StageService) *StageController {
	return &StageController{service: service}
}

// CreateStage handles POST /api/projects/:projectId/stages
func (c *StageController) CreateStage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, err := strconv.ParseInt(vars["projectId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
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

	if req.Name == "" {
		http.Error(w, "Stage name is required", http.StatusBadRequest)
		return
	}

	stage, err := c.service.CreateStage(projectID, req.Name, req.Position)
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
	vars := mux.Vars(r)
	projectID, err := strconv.ParseInt(vars["projectId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	stages, err := c.service.GetStagesByProject(projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stages)
}

// GetStage handles GET /api/stages/:id
func (c *StageController) GetStage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}

	stage, err := c.service.GetStageByID(id)
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

	stage, err := c.service.UpdateStage(id, req.Name, req.Position)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stage)
}

// DeleteStage handles DELETE /api/stages/:id
func (c *StageController) DeleteStage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}

	if err := c.service.DeleteStage(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}