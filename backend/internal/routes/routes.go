package routes

import (
	"net/http"

	"backend/internal/controllers"
	"backend/internal/database"
	"backend/internal/services"

	"github.com/gorilla/mux"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *mux.Router, db *database.DB) {
	// Initialize services
	projectService := services.NewProjectService(db.DB)
	stageService := services.NewStageService(db.DB)
	taskService := services.NewTaskService(db.DB)
	messageService := services.NewMessageService(db.DB)

	// Initialize controllers
	projectController := controllers.NewProjectController(projectService)
	stageController := controllers.NewStageController(stageService)
	taskController := controllers.NewTaskController(taskService)
	messageController := controllers.NewMessageController(messageService)

	// API Routes
	api := router.PathPrefix("/api").Subrouter()

	// Project routes
	api.HandleFunc("/projects", projectController.CreateProject).Methods("POST")
	api.HandleFunc("/projects", projectController.GetAllProjects).Methods("GET")
	api.HandleFunc("/projects/{id}", projectController.GetProject).Methods("GET")
	api.HandleFunc("/projects/{id}", projectController.UpdateProject).Methods("PUT")
	api.HandleFunc("/projects/{id}", projectController.DeleteProject).Methods("DELETE")

	// Stage routes
	api.HandleFunc("/projects/{projectId}/stages", stageController.CreateStage).Methods("POST")
	api.HandleFunc("/projects/{projectId}/stages", stageController.GetStagesByProject).Methods("GET")
	api.HandleFunc("/stages/{id}", stageController.GetStage).Methods("GET")
	api.HandleFunc("/stages/{id}", stageController.UpdateStage).Methods("PUT")
	api.HandleFunc("/stages/{id}", stageController.DeleteStage).Methods("DELETE")

	// Task routes
	api.HandleFunc("/projects/{projectId}/stages/{stageId}/tasks", taskController.CreateTask).Methods("POST")
	api.HandleFunc("/projects/{projectId}/stages/{stageId}/tasks", taskController.GetTasksByStage).Methods("GET")
	api.HandleFunc("/tasks/{id}", taskController.GetTask).Methods("GET")
	api.HandleFunc("/tasks/{id}", taskController.UpdateTask).Methods("PUT")
	api.HandleFunc("/tasks/{id}/move", taskController.MoveTask).Methods("PUT")
	api.HandleFunc("/tasks/{id}", taskController.DeleteTask).Methods("DELETE")

	// Message routes
	api.HandleFunc("/projects/{projectId}/messages", messageController.CreateMessage).Methods("POST")
	api.HandleFunc("/projects/{projectId}/messages", messageController.GetMessagesByProject).Methods("GET")
	api.HandleFunc("/projects/{projectId}/messages/recent", messageController.GetRecentMessages).Methods("GET")
	api.HandleFunc("/messages/{id}", messageController.DeleteMessage).Methods("DELETE")

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
}