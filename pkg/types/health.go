package types

import (
    "time"
)

// HealthStatus représente le statut de santé de l'agent
type HealthStatus struct {
    Status    string    `json:"status"`
    Timestamp time.Time `json:"timestamp"`
    Version   string    `json:"version"`
    Uptime    string    `json:"uptime"`
}

// ReadyStatus représente le statut de readiness de l'agent
type ReadyStatus struct {
    Status      string           `json:"status"`
    Timestamp   time.Time        `json:"timestamp"`
    Components  []ComponentStatus `json:"components,omitempty"`
    Message     string           `json:"message,omitempty"`
}

// ComponentStatus représente le statut d'un composant spécifique
type ComponentStatus struct {
    Name    string `json:"name"`
    Status  string `json:"status"`
    Message string `json:"message,omitempty"`
}

// HealthCheckResult résultat d'un health check
type HealthCheckResult struct {
    Name    string        `json:"name"`
    Status  string        `json:"status"`
    Error   string        `json:"error,omitempty"`
    Latency time.Duration `json:"latency,omitempty"`
}