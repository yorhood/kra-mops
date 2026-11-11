package health

import (
	"context"
	"sync"
	"time"

	"github.com/yorhood/kra-mops/pkg/types"
)

type HealthService struct {
	startTime    time.Time
	version      string
	mutex        sync.RWMutex
	checks       []HealthCheck
	ready        bool
	readyMessage string
}

type HealthCheck struct {
	Name     string
	CheckFn  func(ctx context.Context) error
	Critical bool
}

func NewHealthService(version string) *HealthService {
	return &HealthService{
		startTime: time.Now(),
		version:   version,
		checks:    make([]HealthCheck, 0),
		ready:     false,
	}
}

// AddCheck ajoute un health check
func (h *HealthService) AddCheck(name string, checkFn func(ctx context.Context) error, critical bool) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.checks = append(h.checks, HealthCheck{
		Name:     name,
		CheckFn:  checkFn,
		Critical: critical,
	})
}

// SetReady marque l'agent comme ready
func (h *HealthService) SetReady(ready bool, message string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.ready = ready
	h.readyMessage = message
}

// IsReady vérifie si l'agent est ready
func (h *HealthService) IsReady() bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.ready
}

// GetHealthStatus retourne le statut de santé
func (h *HealthService) GetHealthStatus(ctx context.Context) types.HealthStatus {
	return types.HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Version:   h.version,
		Uptime:    time.Since(h.startTime).String(),
	}
}

// GetReadyStatus retourne le statut de readiness
func (h *HealthService) GetReadyStatus(ctx context.Context) types.ReadyStatus {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if !h.ready {
		return types.ReadyStatus{
			Status:    "not_ready",
			Timestamp: time.Now().UTC(),
			Message:   h.readyMessage,
		}
	}

	// Exécuter les checks de readiness
	components := h.runReadyChecks(ctx)

	status := "ready"
	for _, comp := range components {
		if comp.Status == "unhealthy" {
			status = "degraded"
			break
		}
	}

	return types.ReadyStatus{
		Status:     status,
		Timestamp:  time.Now().UTC(),
		Components: components,
	}
}

// RunHealthChecks exécute tous les health checks
func (h *HealthService) RunHealthChecks(ctx context.Context) []types.HealthCheckResult {
	h.mutex.RLock()
	checks := make([]HealthCheck, len(h.checks))
	copy(checks, h.checks)
	h.mutex.RUnlock()

	var results []types.HealthCheckResult
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, check := range checks {
		wg.Add(1)
		go func(check HealthCheck) {
			defer wg.Done()

			start := time.Now()
			err := check.CheckFn(ctx)
			latency := time.Since(start)

			mutex.Lock()
			result := types.HealthCheckResult{
				Name:    check.Name,
				Latency: latency,
			}

			if err != nil {
				result.Status = "unhealthy"
				result.Error = err.Error()
			} else {
				result.Status = "healthy"
			}

			results = append(results, result)
			mutex.Unlock()
		}(check)
	}

	wg.Wait()
	return results
}

func (h *HealthService) runReadyChecks(ctx context.Context) []types.ComponentStatus {
	results := h.RunHealthChecks(ctx)
	var components []types.ComponentStatus

	for _, result := range results {
		component := types.ComponentStatus{
			Name:   result.Name,
			Status: result.Status,
		}

		if result.Error != "" {
			component.Message = result.Error
		}

		components = append(components, component)
	}

	return components
}

// GetUptime retourne le temps de fonctionnement
func (h *HealthService) GetUptime() time.Duration {
	return time.Since(h.startTime)
}
