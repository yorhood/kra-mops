package health

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/yorhood/kra-mops/internal/k8s"
)

// K8sHealthCheck vérifie la connectivité Kubernetes
func K8sHealthCheck(k8sClient *k8s.ClientWrapper) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return k8sClient.HealthCheck(ctx)
	}
}

// HTTPExporterHealthCheck vérifie la connectivité à l'endpoint d'export
func HTTPExporterHealthCheck(endpoint string, timeout time.Duration) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		client := &http.Client{Timeout: timeout}

		req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to connect to exporter endpoint: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("exporter endpoint returned status: %d", resp.StatusCode)
		}

		return nil
	}
}

// MemoryHealthCheck vérifie l'utilisation mémoire de l'agent
func MemoryHealthCheck(threshold uint64) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		var stats runtime.MemStats
		runtime.ReadMemStats(&stats)

		if stats.Alloc > threshold {
			return fmt.Errorf("memory usage too high: %d bytes (threshold: %d)", stats.Alloc, threshold)
		}

		return nil
	}
}
