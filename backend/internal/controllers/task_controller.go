package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/services"

	"github.com/gorilla/mux"
)

type TaskController struct {
	service *services.TaskService
}

func NewTaskController(service *services.TaskService) *TaskController {
	return &TaskController{service: service}
}

// CreateTask handles POST /api/stages/:stageId/tasks
func (c *TaskController) CreateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stageID, err := strconv.ParseInt(vars["stageId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Position    int    `json:"position"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Task title is required", http.StatusBadRequest)
		return
	}

	task, err := c.service.CreateTask(stageID, req.Title, req.Description, req.Position)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// GetTasksByStage handles GET /api/stages/:stageId/tasks
func (c *TaskController) GetTasksByStage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stageID, err := strconv.ParseInt(vars["stageId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}

	tasks, err := c.service.GetTasksByStage(stageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// GetTask handles GET /api/tasks/:id
func (c *TaskController) GetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := c.service.GetTaskByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if task == nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// UpdateTask handles PUT /api/tasks/:id
func (c *TaskController) UpdateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Position    int    `json:"position"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	task, err := c.service.UpdateTask(id, req.Title, req.Description, req.Position)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// MoveTask handles PUT /api/tasks/:id/move
func (c *TaskController) MoveTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var req struct {
		NewStageID int `json:"newStageId"`
		NewPos     int `json:"newPos"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	task, err := c.service.MoveTask(id, int64(req.NewStageID), req.NewPos)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// DeleteTask handles DELETE /api/tasks/:id
func (c *TaskController) DeleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	if err := c.service.DeleteTask(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}