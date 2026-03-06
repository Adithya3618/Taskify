package routes

import (
	"log"
	"net/http"

	"backend/internal/auth/controller"
	"backend/internal/auth/middleware"
	"backend/internal/auth/repository"
	"backend/internal/auth/services"
	"backend/internal/controllers"
	"backend/internal/database"
	projectServices "backend/internal/services"

	"github.com/gorilla/mux"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *mux.Router, db *database.DB) {
	// Initialize auth repository and create users table
	userRepo := repository.NewUserRepository(db.DB)
	if err := userRepo.InitTable(); err != nil {
		log.Printf("Warning: failed to initialize users table: %v", err)
	}

	// Initialize auth services
	jwtSecret := services.GetEnvJWTSecret()
	jwtService := services.NewJWTService(jwtSecret, 24) // 24 hour expiration
	authService := services.NewAuthService(userRepo, jwtService)
	authController := controller.NewAuthController(authService)

	// Initialize business services
	projectService := projectServices.NewProjectService(db.DB)
	stageService := projectServices.NewStageService(db.DB)
	taskService := projectServices.NewTaskService(db.DB)
	messageService := projectServices.NewMessageService(db.DB)

	// Initialize controllers
	projectController := controllers.NewProjectController(projectService)
	stageController := controllers.NewStageController(stageService)
	taskController := controllers.NewTaskController(taskService)
	messageController := controllers.NewMessageController(messageService)

	// Create JWT middleware
	jwtMiddleware := middleware.JWTAuthMiddleware(jwtService)

	// API Routes
	api := router.PathPrefix("/api").Subrouter()

	// Auth routes (public - no authentication required)
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", authController.Register).Methods("POST")
	auth.HandleFunc("/login", authController.Login).Methods("POST")

	// Protected routes - require JWT authentication
	protected := api.PathPrefix("").Subrouter()
	protected.Use(jwtMiddleware)

	// Protected auth route (GET /me)
	protected.HandleFunc("/auth/me", authController.GetMe).Methods("GET")

	// Project routes (protected)
	protected.HandleFunc("/projects", projectController.CreateProject).Methods("POST")
	protected.HandleFunc("/projects", projectController.GetAllProjects).Methods("GET")
	protected.HandleFunc("/projects/{id}", projectController.GetProject).Methods("GET")
	protected.HandleFunc("/projects/{id}", projectController.UpdateProject).Methods("PUT")
	protected.HandleFunc("/projects/{id}", projectController.DeleteProject).Methods("DELETE")

	// Stage routes (protected)
	protected.HandleFunc("/projects/{projectId}/stages", stageController.CreateStage).Methods("POST")
	protected.HandleFunc("/projects/{projectId}/stages", stageController.GetStagesByProject).Methods("GET")
	protected.HandleFunc("/stages/{id}", stageController.GetStage).Methods("GET")
	protected.HandleFunc("/stages/{id}", stageController.UpdateStage).Methods("PUT")
	protected.HandleFunc("/stages/{id}", stageController.DeleteStage).Methods("DELETE")

	// Task routes (protected)
	protected.HandleFunc("/projects/{projectId}/stages/{stageId}/tasks", taskController.CreateTask).Methods("POST")
	protected.HandleFunc("/projects/{projectId}/stages/{stageId}/tasks", taskController.GetTasksByStage).Methods("GET")
	protected.HandleFunc("/tasks/{id}", taskController.GetTask).Methods("GET")
	protected.HandleFunc("/tasks/{id}", taskController.UpdateTask).Methods("PUT")
	protected.HandleFunc("/tasks/{id}/move", taskController.MoveTask).Methods("PUT")
	protected.HandleFunc("/tasks/{id}", taskController.DeleteTask).Methods("DELETE")

	// Message routes (protected)
	protected.HandleFunc("/projects/{projectId}/messages", messageController.CreateMessage).Methods("POST")
	protected.HandleFunc("/projects/{projectId}/messages", messageController.GetMessagesByProject).Methods("GET")
	protected.HandleFunc("/projects/{projectId}/messages/recent", messageController.GetRecentMessages).Methods("GET")
	protected.HandleFunc("/messages/{id}", messageController.DeleteMessage).Methods("DELETE")

	// Health check endpoint (public)
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
}
