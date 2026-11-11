package collector

import (
	"context"

	"github.com/yorhood/kra-mops/pkg/types"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *NodeCollector) collectCPUStats(ctx context.Context, node v1.Node) (types.CPUStats, error) {
	// Obtenir les capacités du nœud
	totalCores := node.Status.Capacity.Cpu()
	totalCoresValue := totalCores.AsApproximateFloat64()

	// Obtenir les métriques d'utilisation
	nodeMetrics, err := c.metricsClient.MetricsV1beta1().NodeMetricses().Get(ctx, node.Name, metav1.GetOptions{})
	if err != nil {
		return types.CPUStats{}, err
	}

	usedCores := nodeMetrics.Usage.Cpu().AsApproximateFloat64()
	availableCores := totalCoresValue - usedCores
	usagePercent := (usedCores / totalCoresValue) * 100

	return types.CPUStats{
		TotalCores:     totalCoresValue,
		UsedCores:      usedCores,
		AvailableCores: availableCores,
		UsagePercent:   usagePercent,
	}, nil
}
