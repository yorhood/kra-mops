package collector

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yorhood/kra-mops/pkg/types"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

type NodeCollector struct {
	k8sClient     *kubernetes.Clientset
	metricsClient *metrics.Clientset
	clusterName   string
}

func NewNodeCollector(k8sClient *kubernetes.Clientset, metricsClient *metrics.Clientset, clusterName string) *NodeCollector {
	return &NodeCollector{
		k8sClient:     k8sClient,
		metricsClient: metricsClient,
		clusterName:   clusterName,
	}
}

// CollectAllNodeMetrics collecte les métriques pour tous les nœuds
func (c *NodeCollector) CollectAllNodeMetrics(ctx context.Context) ([]types.NodeMetrics, error) {
	nodes, err := c.k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %v", err)
	}

	var allMetrics []types.NodeMetrics
	errors := make([]error, 0)

	for _, node := range nodes.Items {
		nodeMetrics, err := c.collectMetricsForNode(ctx, node)
		if err != nil {
			log.Printf("Error collecting metrics for node %s: %v", node.Name, err)
			errors = append(errors, err)
			continue
		}
		allMetrics = append(allMetrics, nodeMetrics)
	}

	if len(errors) > 0 && len(allMetrics) == 0 {
		return nil, fmt.Errorf("failed to collect metrics for any node: %v", errors)
	}

	return allMetrics, nil
}

// collectMetricsForNode collecte les métriques pour un nœud spécifique
func (c *NodeCollector) collectMetricsForNode(ctx context.Context, node v1.Node) (types.NodeMetrics, error) {
	metrics := types.NodeMetrics{
		NodeName:    node.Name,
		Timestamp:   time.Now().UTC(),
		ClusterName: c.clusterName,
		NodeInfo: types.NodeInfo{
			KernelVersion:    node.Status.NodeInfo.KernelVersion,
			OSImage:          node.Status.NodeInfo.OSImage,
			ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
			KubeletVersion:   node.Status.NodeInfo.KubeletVersion,
			Architecture:     node.Status.NodeInfo.Architecture,
			OperatingSystem:  node.Status.NodeInfo.OperatingSystem,
		},
	}

	// Collecter les métriques pods
	podStats, err := c.collectPodStats(ctx, node.Name)
	if err != nil {
		return metrics, fmt.Errorf("failed to collect pod stats: %v", err)
	}
	metrics.PodStats = podStats

	// Collecter les métriques CPU
	cpuStats, err := c.collectCPUStats(ctx, node)
	if err != nil {
		return metrics, fmt.Errorf("failed to collect CPU stats: %v", err)
	}
	metrics.CPUStats = cpuStats

	// Collecter les métriques mémoire
	memoryStats, err := c.collectMemoryStats(ctx, node)
	if err != nil {
		return metrics, fmt.Errorf("failed to collect memory stats: %v", err)
	}
	metrics.MemoryStats = memoryStats

	// Collecter les métriques stockage
	storageStats, err := c.collectStorageStats(ctx, node)
	if err != nil {
		return metrics, fmt.Errorf("failed to collect storage stats: %v", err)
	}
	metrics.StorageStats = storageStats

	return metrics, nil
}
