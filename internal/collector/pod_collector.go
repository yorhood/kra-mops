package collector

import (
	"context"

	"github.com/yorhood/kra-mops/pkg/types"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *NodeCollector) collectPodStats(ctx context.Context, nodeName string) (types.PodStats, error) {
	pods, err := c.k8sClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	})
	if err != nil {
		return types.PodStats{}, err
	}

	scheduledPods := 0
	for _, pod := range pods.Items {
		// Compter les pods qui sont réellement schedulés et running
		if pod.Status.Phase == v1.PodRunning || pod.Status.Phase == v1.PodPending {
			scheduledPods++
		}
	}

	return types.PodStats{
		TotalPods:       len(pods.Items),
		ScheduledPods:   scheduledPods,
		UnscheduledPods: len(pods.Items) - scheduledPods,
	}, nil
}
