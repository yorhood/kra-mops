package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/yorhood/kra-mops/pkg/types"
)

type StdoutExporter struct {
	enabled bool
}

func NewStdoutExporter(enabled bool) *StdoutExporter {
	return &StdoutExporter{
		enabled: enabled,
	}
}

func (e *StdoutExporter) Export(ctx context.Context, metrics []types.NodeMetrics) error {
	if !e.enabled {
		return nil
	}

	if len(metrics) == 0 {
		log.Println("No metrics to export to stdout")
		return nil
	}

	jsonData, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics to JSON: %v", err)
	}

	log.Printf("Exporting %d metrics to stdout:\n%s", len(metrics), string(jsonData))
	return nil
}

func (e *StdoutExporter) HealthCheck(ctx context.Context) error {
	// stdout exporter is always healthy
	return nil
}
