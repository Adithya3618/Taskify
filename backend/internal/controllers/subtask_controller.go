package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"backend/internal/helpers"
	"backend/internal/services"
)

type SubtaskController struct {
	service *services.SubtaskService
}

func NewSubtaskController(service *services.SubtaskService) *SubtaskController {
	return &SubtaskController{service: service}
}

func (c *SubtaskController) CreateSubtask(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	taskID, err := parseInt64RouteParam(r, "id")
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Title    string `json:"title"`
		Position *int   `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	subtask, err := c.service.CreateSubtask(userID, taskID, req.Title, req.Position)
	if err != nil {
		c.handleSubtaskError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(subtask)
}

func (c *SubtaskController) GetSubtasksByTask(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	taskID, err := parseInt64RouteParam(r, "id")
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	subtasks, err := c.service.GetSubtasksByTask(userID, taskID)
	if err != nil {
		c.handleSubtaskError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subtasks)
}

func (c *SubtaskController) UpdateSubtask(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	subtaskID, err := parseInt64RouteParam(r, "id")
	if err != nil {
		http.Error(w, "Invalid subtask ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Title       *string `json:"title"`
		IsCompleted *bool   `json:"is_completed"`
		Position    *int    `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	subtask, err := c.service.UpdateSubtask(userID, subtaskID, services.SubtaskPatch{
		Title:       req.Title,
		IsCompleted: req.IsCompleted,
		Position:    req.Position,
	})
	if err != nil {
		c.handleSubtaskError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subtask)
}

func (c *SubtaskController) DeleteSubtask(w http.ResponseWriter, r *http.Request) {
	userID := helpers.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	subtaskID, err := parseInt64RouteParam(r, "id")
	if err != nil {
		http.Error(w, "Invalid subtask ID", http.StatusBadRequest)
		return
	}

	if err := c.service.DeleteSubtask(userID, subtaskID); err != nil {
		c.handleSubtaskError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *SubtaskController) handleSubtaskError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrSubtaskTitleRequired),
		errors.Is(err, services.ErrInvalidSubtaskPosition):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, services.ErrTaskNotFoundOrAccessDenied),
		errors.Is(err, services.ErrSubtaskNotFoundOrAccessDenied):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
