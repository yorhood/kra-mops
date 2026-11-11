package exporter

import (
	"context"

	"github.com/yorhood/kra-mops/pkg/types"
)

// Exporter defines the interface that all exporters must implement
type Exporter interface {
	Export(ctx context.Context, metrics []types.NodeMetrics) error
	HealthCheck(ctx context.Context) error
}
