package collector

import (
	"context"

	"github.com/yorhood/kra-mops/pkg/types"

	v1 "k8s.io/api/core/v1"
)

func (c *NodeCollector) collectStorageStats(ctx context.Context, node v1.Node) (types.StorageStats, error) {
	// Pour le stockage, nous utilisons la capacité éphemère comme référence
	// Note: Cette approche donne une estimation, pour des métriques précises
	// il faudrait utiliser un DaemonSet avec des privilèges
	storage := node.Status.Capacity.StorageEphemeral()
	if storage == nil {
		// Fallback sur la capacité générale de stockage
		storage = node.Status.Capacity.Storage()
	}

	totalStorage := int64(0)
	if storage != nil {
		totalStorage = storage.Value()
	}

	// Dans une implémentation réelle, vous voudriez peut-être:
	// 1. Utiliser un DaemonSet avec des privilèges pour exécuter 'df'
	// 2. Implémenter un sidecar qui collecte les métriques de stockage
	// 3. Utiliser des outils comme node-exporter

	// Pour l'instant, nous retournons des valeurs par défaut
	// Vous devrez adapter cette partie selon vos besoins
	usedStorage := int64(0)
	if totalStorage > 0 {
		usedStorage = totalStorage / 3 // Estimation
	}

	freeStorage := totalStorage - usedStorage
	usagePercent := 0.0
	if totalStorage > 0 {
		usagePercent = (float64(usedStorage) / float64(totalStorage)) * 100
	}

	return types.StorageStats{
		TotalBytes:   totalStorage,
		FreeBytes:    freeStorage,
		UsedBytes:    usedStorage,
		UsagePercent: usagePercent,
		FileSystem:   "ephemeral",
	}, nil
}
