package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yorhood/kra-mops/internal/api"
	"github.com/yorhood/kra-mops/internal/collector"
	"github.com/yorhood/kra-mops/internal/exporter"
	"github.com/yorhood/kra-mops/internal/health"
	"github.com/yorhood/kra-mops/internal/k8s"
	"github.com/yorhood/kra-mops/internal/processor"
	"github.com/yorhood/kra-mops/pkg/types"
	"gopkg.in/yaml.v2"
)

var (
	configFile = flag.String("config", "configs/config.yaml", "Path to configuration file")
	version    = "1.0.0"
)

func main() {
	flag.Parse()

	log.Printf("Starting K8s Monitoring Agent version %s", version)

	// Charger la configuration
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialiser le service de health check
	healthService := health.NewHealthService(version)

	// Initialiser le client Kubernetes
	k8sClient, err := k8s.NewClient(&config.K8S)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Ajouter les health checks
	healthService.AddCheck("kubernetes-api", health.K8sHealthCheck(k8sClient), true)

	// Vérifier la connectivité initiale
	ctx := context.Background()
	if err := k8sClient.HealthCheck(ctx); err != nil {
		log.Fatalf("Kubernetes health check failed: %v", err)
	}

	// Initialiser les composants
	nodeCollector := collector.NewNodeCollector(
		k8sClient.K8SClient,
		k8sClient.MetricsClient,
		"my-cluster",
	)

	metricsProcessor := processor.NewMetricsProcessor("my-cluster")

	var metricsExporter exporter.Exporter

	switch config.Exporter.Type {
	case "http":
		httpExporter := exporter.NewHTTPExporter(&config.Exporter.HTTP)
		metricsExporter = httpExporter

		// Ajouter health check pour l'exporter HTTP
		healthService.AddCheck("http-exporter",
			health.HTTPExporterHealthCheck(
				config.Exporter.HTTP.Endpoint,
				time.Duration(config.Exporter.HTTP.Timeout)*time.Second,
			),
			false,
		)
		log.Println("Using HTTP exporter")

	case "stdout":
		stdoutExporter := exporter.NewStdoutExporter(config.Exporter.Stdout.Enabled)
		metricsExporter = stdoutExporter
		log.Println("Using stdout exporter - metrics will be logged to console")

	case "none":
		noneExporter := exporter.NewNoneExporter()
		metricsExporter = noneExporter
		log.Println("Exporter disabled - metrics are collected but not exported")

	default:
		log.Printf("Unsupported exporter type: %s, using stdout exporter", config.Exporter.Type)
		metricsExporter = exporter.NewStdoutExporter(true)
	}

	// Démarrer le serveur HTTP
	apiServer := api.NewServer(healthService, "8080")
	go func() {
		if err := apiServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Attendre que le serveur HTTP soit prêt
	time.Sleep(2 * time.Second)

	// Marquer l'agent comme ready après l'initialisation
	healthService.SetReady(true, "Agent fully initialized")

	// Démarrer la boucle de collecte
	runCollectionLoop(ctx, config, nodeCollector, metricsProcessor, metricsExporter, healthService)

	// Arrêter le serveur HTTP
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := apiServer.Stop(shutdownCtx); err != nil {
		log.Printf("Error stopping HTTP server: %v", err)
	}

	log.Println("K8s Monitoring Agent stopped gracefully")
}

func loadConfig(configPath string) (*types.Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config types.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

func runCollectionLoop(
	ctx context.Context,
	config *types.Config,
	collector *collector.NodeCollector,
	processor *processor.MetricsProcessor,
	exporter exporter.Exporter,
	healthService *health.HealthService,
) {
	ticker := time.NewTicker(time.Duration(config.Agent.CollectionInterval) * time.Second)
	defer ticker.Stop()

	// Canal pour les signaux d'arrêt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Première collecte immédiate
	if err := collectAndExport(ctx, collector, processor, exporter); err != nil {
		log.Printf("Initial collection failed: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := collectAndExport(ctx, collector, processor, exporter); err != nil {
				log.Printf("Collection failed: %v", err)
			}

		case sig := <-sigChan:
			log.Printf("Received shutdown signal: %s", sig)
			// Marquer l'agent comme not ready pendant l'arrêt
			healthService.SetReady(false, "Shutting down")
			return

		case <-ctx.Done():
			log.Println("Context cancelled")
			healthService.SetReady(false, "Context cancelled")
			return
		}
	}
}

func collectAndExport(
	ctx context.Context,
	collector *collector.NodeCollector,
	processor *processor.MetricsProcessor,
	exporter exporter.Exporter,
) error {
	startTime := time.Now()

	// Collecter les métriques
	metrics, err := collector.CollectAllNodeMetrics(ctx)
	if err != nil {
		return fmt.Errorf("failed to collect metrics: %v", err)
	}

	// Traiter les métriques
	processedMetrics := processor.ProcessBatch(metrics)

	// Exporter les métriques
	if err := exporter.Export(ctx, processedMetrics); err != nil {
		return fmt.Errorf("failed to export metrics: %v", err)
	}

	log.Printf("Successfully collected and exported %d metrics in %v",
		len(processedMetrics), time.Since(startTime))

	return nil
}
