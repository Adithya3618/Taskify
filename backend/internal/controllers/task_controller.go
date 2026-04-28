package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"backend/internal/helpers"
	"backend/internal/models"
	"backend/internal/services"

	"github.com/gorilla/mux"
)

type TaskController struct {
	service *services.TaskService
}

type taskRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Position    int        `json:"position"`
	StartDate   *time.Time `json:"start_date"`
	Deadline    *time.Time `json:"deadline"`
	Priority    *string    `json:"priority"`
	AssignedTo  *string    `json:"assigned_to"`
}

type taskUpdateRequest struct {
	taskRequest
	StartDateProvided  bool
	DeadlineProvided   bool
	PriorityProvided   bool
	AssignedToProvided bool
}

func NewTaskController(service *services.TaskService) *TaskController {
	return &TaskController{service: service}
}

// CreateTask handles POST /api/stages/:stageId/tasks
func (c *TaskController) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	stageID, err := strconv.ParseInt(vars["stageId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}

	req, err := decodeTaskRequest(r)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Task title is required", http.StatusBadRequest)
		return
	}

	task, err := c.service.CreateTask(userID, stageID, req.Title, req.Description, req.Position, taskAttributesFromRequest(req))
	if err != nil {
		if errors.Is(err, services.ErrInvalidTaskPriority) || errors.Is(err, services.ErrInvalidDateRange) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// GetTasksByStage handles GET /api/stages/:stageId/tasks
func (c *TaskController) GetTasksByStage(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	stageID, err := strconv.ParseInt(vars["stageId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}

	tasks, err := c.service.GetTasksByStage(userID, stageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// GetTask handles GET /api/tasks/:id
func (c *TaskController) GetTask(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := c.service.GetTaskByID(userID, id)
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

// GetProjectTimeline handles GET /api/projects/:id/timeline
func (c *TaskController) GetProjectTimeline(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	projectID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil || projectID <= 0 {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	timeline, err := c.service.GetProjectTimeline(userID, projectID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(timeline)
}

// SearchProjectTasks handles GET /api/projects/:id/tasks/search?q=
func (c *TaskController) SearchProjectTasks(w http.ResponseWriter, r *http.Request) {
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

	results, err := c.service.SearchProjectTasks(userID, projectID, r.URL.Query().Get("q"))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// UpdateTask handles PUT /api/tasks/:id
func (c *TaskController) UpdateTask(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	req, err := decodeTaskUpdateRequest(r)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	existingTask, err := c.service.GetTaskByID(userID, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existingTask == nil {
		http.Error(w, "Task not found or access denied", http.StatusNotFound)
		return
	}

	attrs := mergeTaskAttributes(existingTask, req)

	task, err := c.service.UpdateTask(userID, id, req.Title, req.Description, req.Position, attrs)
	if err != nil {
		if errors.Is(err, services.ErrInvalidTaskPriority) || errors.Is(err, services.ErrInvalidDateRange) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if task == nil {
		http.Error(w, "Task not found or access denied", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// MoveTask handles PUT /api/tasks/:id/move
func (c *TaskController) MoveTask(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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

	task, err := c.service.MoveTask(userID, id, int64(req.NewStageID), req.NewPos)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if task == nil {
		http.Error(w, "Task not found or access denied", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// DeleteTask handles DELETE /api/tasks/:id
func (c *TaskController) DeleteTask(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	if err := c.service.DeleteTask(userID, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func taskAttributesFromRequest(req taskRequest) services.TaskAttributes {
	return services.TaskAttributes{
		StartDate:  req.StartDate,
		Deadline:   req.Deadline,
		Priority:   req.Priority,
		AssignedTo: req.AssignedTo,
	}
}

func mergeTaskAttributes(existing *models.Task, req taskUpdateRequest) services.TaskAttributes {
	attrs := services.TaskAttributes{
		StartDate:  existing.StartDate,
		Deadline:   existing.Deadline,
		Priority:   existing.Priority,
		AssignedTo: existing.AssignedTo,
	}

	if req.StartDateProvided {
		attrs.StartDate = req.StartDate
	}
	if req.DeadlineProvided {
		attrs.Deadline = req.Deadline
	}
	if req.PriorityProvided {
		attrs.Priority = req.Priority
	}
	if req.AssignedToProvided {
		attrs.AssignedTo = req.AssignedTo
	}

	return attrs
}

func decodeTaskRequest(r *http.Request) (taskRequest, error) {
	var req taskRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeTaskUpdateRequest(r *http.Request) (taskUpdateRequest, error) {
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		return taskUpdateRequest{}, err
	}

	payload, err := json.Marshal(raw)
	if err != nil {
		return taskUpdateRequest{}, err
	}

	var req taskUpdateRequest
	if err := json.Unmarshal(payload, &req.taskRequest); err != nil {
		return taskUpdateRequest{}, err
	}

	_, req.StartDateProvided = raw["start_date"]
	_, req.DeadlineProvided = raw["deadline"]
	_, req.PriorityProvided = raw["priority"]
	_, req.AssignedToProvided = raw["assigned_to"]

	return req, nil
}
