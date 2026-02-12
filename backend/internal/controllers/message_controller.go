package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/services"

	"github.com/gorilla/mux"
)

type MessageController struct {
	service *services.MessageService
}

func NewMessageController(service *services.MessageService) *MessageController {
	return &MessageController{service: service}
}

// CreateMessage handles POST /api/projects/:projectId/messages
func (c *MessageController) CreateMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, err := strconv.ParseInt(vars["projectId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var req struct {
		SenderName string `json:"sender_name"`
		Content    string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SenderName == "" || req.Content == "" {
		http.Error(w, "Sender name and content are required", http.StatusBadRequest)
		return
	}

	message, err := c.service.CreateMessage(projectID, req.SenderName, req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}

// GetMessagesByProject handles GET /api/projects/:projectId/messages
func (c *MessageController) GetMessagesByProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, err := strconv.ParseInt(vars["projectId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	messages, err := c.service.GetMessagesByProject(projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// GetRecentMessages handles GET /api/projects/:projectId/messages/recent
func (c *MessageController) GetRecentMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, err := strconv.ParseInt(vars["projectId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	messages, err := c.service.GetRecentMessages(projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// DeleteMessage handles DELETE /api/messages/:id
func (c *MessageController) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	if err := c.service.DeleteMessage(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}