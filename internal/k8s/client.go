package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yorhood/kra-mops/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

// ClientWrapper encapsule les clients Kubernetes
type ClientWrapper struct {
	K8SClient     *kubernetes.Clientset
	MetricsClient *metrics.Clientset
	Config        *types.K8SConfig
}

// NewClient crée un nouveau client Kubernetes
func NewClient(config *types.K8SConfig) (*ClientWrapper, error) {
	var k8sConfig *rest.Config
	var err error

	if config.InCluster {
		k8sConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create in-cluster config: %v", err)
		}
	} else {
		kubeconfig := config.Kubeconfig
		if kubeconfig == "" {
			if home := homedir.HomeDir(); home != "" {
				kubeconfig = filepath.Join(home, ".kube", "config")
			}
		}

		if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
			return nil, fmt.Errorf("kubeconfig file not found: %s", kubeconfig)
		}

		k8sConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig: %v", err)
		}
	}

	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	metricsClient, err := metrics.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %v", err)
	}

	return &ClientWrapper{
		K8SClient:     k8sClient,
		MetricsClient: metricsClient,
		Config:        config,
	}, nil
}

// HealthCheck vérifie la connectivité avec le cluster Kubernetes
func (c *ClientWrapper) HealthCheck(ctx context.Context) error {
	_, err := c.K8SClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %v", err)
	}
	return nil
}
