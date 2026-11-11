package exporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/yorhood/kra-mops/pkg/types"
)

type HTTPExporter struct {
	config     *types.HTTPConfig
	httpClient *http.Client
}

func NewHTTPExporter(config *types.HTTPConfig) *HTTPExporter {
	return &HTTPExporter{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}
}

// Export implements Exporter interface
func (e *HTTPExporter) Export(ctx context.Context, metrics []types.NodeMetrics) error {
	if len(metrics) == 0 {
		log.Println("No metrics to export, skipping")
		return nil
	}

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics to JSON: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.config.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Traiter les headers avec expansion des variables d'environnement
	for key, value := range e.config.Headers {
		expandedValue := e.expandEnvVars(value)
		// Ne pas ajouter le header Authorization s'il n'y a pas de token valide
		if key == "Authorization" && (expandedValue == "Bearer " || expandedValue == "Bearer default-token") {
			log.Println("Warning: No valid API token provided, skipping Authorization header")
			continue
		}
		req.Header.Set(key, expandedValue)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	log.Printf("Successfully exported %d metrics to %s", len(metrics), e.config.Endpoint)
	return nil
}

// HealthCheck implements Exporter interface
func (e *HTTPExporter) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", e.config.Endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %v", err)
	}

	// Ajouter les headers sans le token d'autorisation pour le health check
	for key, value := range e.config.Headers {
		if key != "Authorization" { // Ne pas envoyer le token pour les health checks
			expandedValue := e.expandEnvVars(value)
			req.Header.Set(key, expandedValue)
		}
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// expandEnvVars remplace les variables d'environnement dans les strings
func (e *HTTPExporter) expandEnvVars(s string) string {
	return os.Expand(s, func(key string) string {
		if val := os.Getenv(key); val != "" {
			return val
		}
		// Retourner la valeur par défaut si spécifiée après :-
		if idx := strings.Index(key, ":-"); idx != -1 {
			return key[idx+2:]
		}
		return ""
	})
}
