package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"backend/internal/chat"
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/routes"

	"github.com/gorilla/mux"
)

// Custom CORS middleware that allows all localhost origins
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Allow any localhost origin for development
		if strings.HasPrefix(origin, "http://localhost:") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.NewDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize router
	router := mux.NewRouter()

	// Initialize routes
	routes.SetupRoutes(router, db)

	// Setup chat hub
	hub := chat.NewHub()
	go hub.Run()

	// WebSocket endpoint
	router.HandleFunc("/ws/{projectId}", func(w http.ResponseWriter, r *http.Request) {
		chat.ServeWs(hub, w, r)
	})

	// Enable CORS - allow all localhost origins for development
	corsHandler := enableCORS(router)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on http://localhost:%s/api/projects", cfg.Port)
	log.Fatal(http.ListenAndServe(addr, corsHandler))
}
