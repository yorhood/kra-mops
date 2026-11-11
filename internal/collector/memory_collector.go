package collector

import (
	"context"

	"github.com/yorhood/kra-mops/pkg/types"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *NodeCollector) collectMemoryStats(ctx context.Context, node v1.Node) (types.MemoryStats, error) {
	// Obtenir les capacités du nœud
	totalMemory := node.Status.Capacity.Memory()
	totalMemoryValue := totalMemory.Value()

	// Obtenir les métriques d'utilisation
	nodeMetrics, err := c.metricsClient.MetricsV1beta1().NodeMetricses().Get(ctx, node.Name, metav1.GetOptions{})
	if err != nil {
		return types.MemoryStats{}, err
	}

	usedMemory := nodeMetrics.Usage.Memory().Value()
	availableMemory := totalMemoryValue - usedMemory
	usagePercent := (float64(usedMemory) / float64(totalMemoryValue)) * 100

	return types.MemoryStats{
		TotalBytes:     totalMemoryValue,
		UsedBytes:      usedMemory,
		AvailableBytes: availableMemory,
		UsagePercent:   usagePercent,
	}, nil
}
