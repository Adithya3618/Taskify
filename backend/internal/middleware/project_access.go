package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/services"

	"github.com/gorilla/mux"
)

// getUserID retrieves the user ID from the request context (from auth middleware)
func getUserID(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

// ProjectMemberServiceInterface for dependency injection
type ProjectMemberServiceInterface interface {
	HasAccess(projectID int64, userID string) (bool, error)
}

// ProjectAccessMiddleware creates middleware that checks if user has access to a project
func ProjectAccessMiddleware(pmService *services.ProjectMemberService) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user ID from JWT context
			userID := getUserID(r.Context())
			if userID == "" {
				writeJSONError(w, http.StatusUnauthorized, "Authentication required")
				return
			}

			// Get project ID from URL - check both parameter names
			vars := mux.Vars(r)
			projectIDStr, ok := vars["id"]
			if !ok {
				projectIDStr, ok = vars["projectId"]
				if !ok {
					// No project ID in URL, skip this middleware
					next.ServeHTTP(w, r)
					return
				}
			}

			projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
			if err != nil {
				writeJSONError(w, http.StatusBadRequest, "Invalid project ID")
				return
			}

			// Check if user has access to the project
			hasAccess, err := pmService.HasAccess(projectID, userID)
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, "Failed to check project access")
				return
			}

			if !hasAccess {
				writeJSONError(w, http.StatusForbidden, "Access denied to this project")
				return
			}

			// Add project ID to context for use in handlers
			ctx := context.WithValue(r.Context(), ProjectIDKey, projectID)

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OwnerOnlyMiddleware creates middleware that checks if user is the owner of a project
func OwnerOnlyMiddleware(pmService *services.ProjectMemberService) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user ID from JWT context
			userID := getUserID(r.Context())
			if userID == "" {
				writeJSONError(w, http.StatusUnauthorized, "Authentication required")
				return
			}

			// Get project ID from URL
			vars := mux.Vars(r)
			projectIDStr, ok := vars["id"]
			if !ok {
				projectIDStr, ok = vars["projectId"]
				if !ok {
					next.ServeHTTP(w, r)
					return
				}
			}

			projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
			if err != nil {
				writeJSONError(w, http.StatusBadRequest, "Invalid project ID")
				return
			}

			// Check if user is the owner
			isOwner, err := pmService.IsOwner(projectID, userID)
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, "Failed to check ownership")
				return
			}

			if !isOwner {
				writeJSONError(w, http.StatusForbidden, "Only project owner can perform this action")
				return
			}

			// Add project ID to context
			ctx := context.WithValue(r.Context(), ProjectIDKey, projectID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// ProjectIDKey is the context key for project ID
const ProjectIDKey = "project_id"

// GetProjectIDFromContext retrieves the project ID from the request context
func GetProjectIDFromContext(ctx context.Context) int64 {
	if projectID, ok := ctx.Value(ProjectIDKey).(int64); ok {
		return projectID
	}
	return 0
}
