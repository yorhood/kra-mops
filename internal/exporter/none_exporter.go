package exporter

import (
	"context"

	"github.com/yorhood/kra-mops/pkg/types"
)

type NoneExporter struct{}

func NewNoneExporter() *NoneExporter {
	return &NoneExporter{}
}

func (e *NoneExporter) Export(ctx context.Context, metrics []types.NodeMetrics) error {
	// Do nothing - metrics are collected but not exported
	return nil
}

func (e *NoneExporter) HealthCheck(ctx context.Context) error {
	return nil
}
