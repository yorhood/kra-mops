package processor

import (
	"time"

	"github.com/yorhood/kra-mops/pkg/types"
)

// MetricsProcessor traite et enrichit les métriques
type MetricsProcessor struct {
	clusterName string
}

func NewMetricsProcessor(clusterName string) *MetricsProcessor {
	return &MetricsProcessor{
		clusterName: clusterName,
	}
}

// Process enrichit les métriques avec des calculs supplémentaires
func (p *MetricsProcessor) Process(metrics *types.NodeMetrics) {
	// S'assurer que le nom du cluster est défini
	if metrics.ClusterName == "" {
		metrics.ClusterName = p.clusterName
	}

	// Calculer les pourcentages d'utilisation si nécessaire
	if metrics.CPUStats.TotalCores > 0 && metrics.CPUStats.UsagePercent == 0 {
		metrics.CPUStats.UsagePercent = (metrics.CPUStats.UsedCores / metrics.CPUStats.TotalCores) * 100
	}

	if metrics.MemoryStats.TotalBytes > 0 && metrics.MemoryStats.UsagePercent == 0 {
		metrics.MemoryStats.UsagePercent = (float64(metrics.MemoryStats.UsedBytes) / float64(metrics.MemoryStats.TotalBytes)) * 100
	}

	if metrics.StorageStats.TotalBytes > 0 && metrics.StorageStats.UsagePercent == 0 {
		metrics.StorageStats.UsagePercent = (float64(metrics.StorageStats.UsedBytes) / float64(metrics.StorageStats.TotalBytes)) * 100
	}
}

// ProcessBatch traite un lot de métriques
func (p *MetricsProcessor) ProcessBatch(metrics []types.NodeMetrics) []types.NodeMetrics {
	for i := range metrics {
		p.Process(&metrics[i])
	}
	return metrics
}

// FilterByTimestamp filtre les métriques par timestamp
func (p *MetricsProcessor) FilterByTimestamp(metrics []types.NodeMetrics, since time.Time) []types.NodeMetrics {
	var filtered []types.NodeMetrics
	for _, m := range metrics {
		if m.Timestamp.After(since) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}
