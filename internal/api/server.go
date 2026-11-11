package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/yorhood/kra-mops/internal/health"
	"github.com/yorhood/kra-mops/pkg/types"
)

type Server struct {
	healthService *health.HealthService
	server        *http.Server
}

func NewServer(healthService *health.HealthService, port string) *Server {
	mux := http.NewServeMux()
	server := &Server{
		healthService: healthService,
		server: &http.Server{
			Addr:         ":" + port,
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	// Register routes
	mux.HandleFunc("/health", server.healthHandler)
	mux.HandleFunc("/ready", server.readyHandler)
	mux.HandleFunc("/metrics", server.metricsHandler)
	mux.HandleFunc("/", server.rootHandler)

	return server
}

// Start démarre le serveur HTTP
func (s *Server) Start() error {
	log.Printf("Starting HTTP server on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop arrête le serveur HTTP
func (s *Server) Stop(ctx context.Context) error {
	log.Println("Shutting down HTTP server...")
	return s.server.Shutdown(ctx)
}

// healthHandler gère les requêtes /health
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Exécuter les health checks
	results := s.healthService.RunHealthChecks(ctx)

	// Vérifier s'il y a des erreurs critiques
	hasCriticalError := false
	for _, result := range results {
		if result.Status == "unhealthy" {
			hasCriticalError = true
			break
		}
	}

	healthStatus := s.healthService.GetHealthStatus(ctx)

	// Ajouter les résultats des checks détaillés
	detailedHealth := struct {
		types.HealthStatus
		Checks []types.HealthCheckResult `json:"checks,omitempty"`
	}{
		HealthStatus: healthStatus,
		Checks:       results,
	}

	w.Header().Set("Content-Type", "application/json")

	if hasCriticalError {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(detailedHealth); err != nil {
		log.Printf("Error encoding health response: %v", err)
	}
}

// readyHandler gère les requêtes /ready
func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	readyStatus := s.healthService.GetReadyStatus(ctx)

	w.Header().Set("Content-Type", "application/json")

	switch readyStatus.Status {
	case "ready":
		w.WriteHeader(http.StatusOK)
	case "degraded":
		w.WriteHeader(http.StatusOK) // Still accepting traffic
	case "not_ready":
		w.WriteHeader(http.StatusServiceUnavailable)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err := json.NewEncoder(w).Encode(readyStatus); err != nil {
		log.Printf("Error encoding ready response: %v", err)
	}
}

// metricsHandler gère les requêtes /metrics (pour exposition Prometheus optionnelle)
func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Pour l'instant, retourner des métriques basiques de l'agent
	metrics := map[string]interface{}{
		"agent_uptime_seconds": s.healthService.GetUptime().Seconds(),
		"agent_version":        s.healthService.GetHealthStatus(r.Context()).Version,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		log.Printf("Error encoding metrics response: %v", err)
	}
}

// rootHandler gère la racine
func (s *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	info := map[string]interface{}{
		"name":        "K8s Monitoring Agent",
		"version":     s.healthService.GetHealthStatus(r.Context()).Version,
		"description": "Kubernetes node metrics collection agent",
		"endpoints": map[string]string{
			"health":  "/health",
			"ready":   "/ready",
			"metrics": "/metrics",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(info); err != nil {
		log.Printf("Error encoding root response: %v", err)
	}
}
