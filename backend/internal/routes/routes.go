package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"backend/internal/auth/controller"
	authmiddleware "backend/internal/auth/middleware"
	"backend/internal/auth/repository"
	"backend/internal/auth/services"
	"backend/internal/controllers"
	"backend/internal/database"
	"backend/internal/middleware"
	projectServices "backend/internal/services"

	"github.com/gorilla/mux"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *mux.Router, db *database.DB) {
	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}).Methods("GET")

	// Initialize auth repository and create users table
	userRepo := repository.NewUserRepository(db.DB)
	if err := userRepo.InitTable(); err != nil {
		log.Printf("Warning: failed to initialize users table: %v", err)
	}
	identityRepo := repository.NewAuthIdentityRepository(db.DB)
	if err := identityRepo.InitTable(); err != nil {
		log.Printf("Warning: failed to initialize auth identities table: %v", err)
	}

	// Initialize auth services
	jwtSecret := services.GetEnvJWTSecret()
	jwtService := services.NewJWTService(jwtSecret, 24) // 24 hour expiration
	otpService := services.NewOTPService()
	emailService := services.NewEmailService()
	googleService := services.NewGoogleAuthService()
	oauthStateService := services.NewOAuthStateService(10 * time.Minute)
	authService := services.NewAuthService(userRepo, identityRepo, jwtService, otpService, emailService, googleService, oauthStateService)
	authController := controller.NewAuthController(authService)

	// Initialize business services
	projectService := projectServices.NewProjectService(db.DB)
	stageService := projectServices.NewStageService(db.DB)
	messageService := projectServices.NewMessageService(db.DB)
	projectMemberService := projectServices.NewProjectMemberService(db.DB)
	activityService := projectServices.NewActivityService(db.DB, projectMemberService)
	taskService := projectServices.NewTaskService(db.DB, activityService)
	commentService := projectServices.NewCommentService(db.DB)
	subtaskService := projectServices.NewSubtaskService(db.DB)
	labelService := projectServices.NewLabelService(db.DB, projectMemberService, activityService)
	taskLabelService := projectServices.NewTaskLabelService(db.DB, projectMemberService, activityService)
	notificationService := projectServices.NewNotificationService(db.DB, emailService)

	// Initialize controllers
	projectController := controllers.NewProjectController(projectService)
	stageController := controllers.NewStageController(stageService)
	taskController := controllers.NewTaskController(taskService)
	commentController := controllers.NewCommentController(commentService)
	subtaskController := controllers.NewSubtaskController(subtaskService)
	messageController := controllers.NewMessageController(messageService)
	projectMemberController := controllers.NewProjectMemberController(projectMemberService)
	activityController := controllers.NewActivityController(activityService)
	labelController := controllers.NewLabelController(labelService)
	taskLabelController := controllers.NewTaskLabelController(taskLabelService)
	notificationController := controllers.NewNotificationController(notificationService)

	// Start deadline checker background job (runs every 15 minutes)
	notificationService.StartDeadlineChecker(15 * time.Minute)

	// Create JWT middleware
	jwtMiddleware := authmiddleware.JWTAuthMiddleware(jwtService)

	// Create project access middleware
	projectAccessMiddleware := middleware.ProjectAccessMiddleware(projectMemberService)

	// API Routes
	api := router.PathPrefix("/api").Subrouter()

	// Auth routes (public - no authentication required)
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", authController.Register).Methods("POST")
	auth.HandleFunc("/login", authController.Login).Methods("POST")
	auth.HandleFunc("/google/id-token", authController.GoogleLoginWithIDToken).Methods("POST")
	auth.HandleFunc("/google/login", authController.GoogleLoginRedirect).Methods("GET")
	auth.HandleFunc("/google/callback", authController.GoogleCallback).Methods("GET")
	auth.HandleFunc("/forgot-password", authController.ForgotPassword).Methods("POST")
	auth.HandleFunc("/verify-otp", authController.VerifyOTP).Methods("POST")
	auth.HandleFunc("/reset-password", authController.ResetPassword).Methods("POST")

	// Protected routes - require JWT authentication
	protected := api.PathPrefix("").Subrouter()
	protected.Use(jwtMiddleware)

	// Protected auth route (GET /me, PUT /me)
	protected.HandleFunc("/auth/me", authController.GetMe).Methods("GET")
	protected.HandleFunc("/auth/me", authController.UpdateMe).Methods("PUT")

	// Project routes (protected)
	protected.HandleFunc("/projects", projectController.CreateProject).Methods("POST")
	protected.HandleFunc("/projects", projectController.GetAllProjects).Methods("GET")
	protected.HandleFunc("/projects/{id}", projectController.GetProject).Methods("GET")
	protected.HandleFunc("/projects/{id}/stats", projectController.GetProjectStats).Methods("GET")
	protected.HandleFunc("/projects/{id}", projectController.UpdateProject).Methods("PUT")
	protected.HandleFunc("/projects/{id}", projectController.DeleteProject).Methods("DELETE")

	// Timeline routes (protected with project access check)
	timelineRoutes := api.PathPrefix("/projects/{id}/timeline").Subrouter()
	timelineRoutes.Use(jwtMiddleware)
	timelineRoutes.Use(projectAccessMiddleware)
	timelineRoutes.HandleFunc("", taskController.GetProjectTimeline).Methods("GET")

	// Project task search routes (protected with project access check)
	taskSearchRoutes := api.PathPrefix("/projects/{id}/tasks/search").Subrouter()
	taskSearchRoutes.Use(jwtMiddleware)
	taskSearchRoutes.Use(projectAccessMiddleware)
	taskSearchRoutes.HandleFunc("", taskController.SearchProjectTasks).Methods("GET")

	// Project Member routes (protected with project access check)
	projectMemberRoutes := api.PathPrefix("/projects/{id}/members").Subrouter()
	projectMemberRoutes.Use(jwtMiddleware)
	projectMemberRoutes.Use(projectAccessMiddleware)
	projectMemberRoutes.HandleFunc("", projectMemberController.AddMember).Methods("POST")
	projectMemberRoutes.HandleFunc("", projectMemberController.GetMembers).Methods("GET")
	projectMemberRoutes.HandleFunc("/{userId}", projectMemberController.RemoveMember).Methods("DELETE")

	// Invite routes (protected)
	inviteRoutes := api.PathPrefix("/projects/{id}/invites").Subrouter()
	inviteRoutes.Use(jwtMiddleware)
	inviteRoutes.Use(projectAccessMiddleware)
	inviteRoutes.HandleFunc("", projectMemberController.CreateInvite).Methods("POST")

	// Public invite acceptance (needs auth but not project access)
	protected.HandleFunc("/invites/{id}", projectMemberController.GetInvite).Methods("GET")
	protected.HandleFunc("/invites/{id}/accept", projectMemberController.AcceptInvite).Methods("POST")

	// Stage routes (protected)
	stageReorderRoutes := api.PathPrefix("/projects/{id}/stages/reorder").Subrouter()
	stageReorderRoutes.Use(jwtMiddleware)
	stageReorderRoutes.Use(projectAccessMiddleware)
	stageReorderRoutes.HandleFunc("", stageController.ReorderStages).Methods("PUT")

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
	protected.HandleFunc("/tasks/{id}/assign", taskController.AssignTask).Methods("PUT")
	protected.HandleFunc("/tasks/{id}", taskController.DeleteTask).Methods("DELETE")
	protected.HandleFunc("/tasks/{id}/comments", commentController.CreateComment).Methods("POST")
	protected.HandleFunc("/tasks/{id}/comments", commentController.GetCommentsByTask).Methods("GET")
	protected.HandleFunc("/tasks/{id}/subtasks", subtaskController.CreateSubtask).Methods("POST")
	protected.HandleFunc("/tasks/{id}/subtasks", subtaskController.GetSubtasksByTask).Methods("GET")
	protected.HandleFunc("/comments/{id}", commentController.UpdateComment).Methods("PATCH")
	protected.HandleFunc("/comments/{id}", commentController.DeleteComment).Methods("DELETE")
	protected.HandleFunc("/subtasks/{id}", subtaskController.UpdateSubtask).Methods("PATCH")
	protected.HandleFunc("/subtasks/{id}", subtaskController.DeleteSubtask).Methods("DELETE")

	// Message routes (protected)
	protected.HandleFunc("/projects/{projectId}/messages", messageController.CreateMessage).Methods("POST")
	protected.HandleFunc("/projects/{projectId}/messages", messageController.GetMessagesByProject).Methods("GET")
	protected.HandleFunc("/projects/{projectId}/messages/recent", messageController.GetRecentMessages).Methods("GET")
	protected.HandleFunc("/messages/{id}", messageController.DeleteMessage).Methods("DELETE")

	// Activity routes (protected with project access check)
	activityRoutes := api.PathPrefix("/projects/{id}/activity").Subrouter()
	activityRoutes.Use(jwtMiddleware)
	activityRoutes.Use(projectAccessMiddleware)
	activityRoutes.HandleFunc("", activityController.GetActivity).Methods("GET")
	activityRoutes.HandleFunc("/recent", activityController.GetRecentActivity).Methods("GET")

	// Label routes (protected with project access check)
	labelRoutes := api.PathPrefix("/projects/{id}/labels").Subrouter()
	labelRoutes.Use(jwtMiddleware)
	labelRoutes.Use(projectAccessMiddleware)
	labelRoutes.HandleFunc("", labelController.CreateLabel).Methods("POST")
	labelRoutes.HandleFunc("", labelController.GetLabels).Methods("GET")

	// Single label route (protected)
	protected.HandleFunc("/labels/{id}", labelController.DeleteLabel).Methods("DELETE")

	// Task label routes (protected)
	protected.HandleFunc("/tasks/{id}/labels", taskLabelController.AssignLabel).Methods("POST")
	protected.HandleFunc("/tasks/{id}/labels", taskLabelController.GetTaskLabels).Methods("GET")
	protected.HandleFunc("/tasks/{id}/labels/{labelId}", taskLabelController.RemoveLabel).Methods("DELETE")

	// Notification routes (protected)
	protected.HandleFunc("/notifications", notificationController.GetNotifications).Methods("GET")
	protected.HandleFunc("/notifications/read-all", notificationController.MarkAllAsRead).Methods("PATCH")
	protected.HandleFunc("/notifications/{id}/read", notificationController.MarkAsRead).Methods("PATCH")

	// Health check endpoint (public)
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
}
