package types

import (
    "time"
)

// NodeMetrics contient toutes les métriques collectées pour un nœud
type NodeMetrics struct {
    NodeName      string    `json:"nodeName"`
    Timestamp     time.Time `json:"timestamp"`
    ClusterName   string    `json:"clusterName,omitempty"`
    
    // Métriques pods
    PodStats PodStats `json:"podStats"`
    
    // Métriques CPU
    CPUStats CPUStats `json:"cpuStats"`
    
    // Métriques mémoire
    MemoryStats MemoryStats `json:"memoryStats"`
    
    // Métriques stockage
    StorageStats StorageStats `json:"storageStats"`
    
    // Informations sur le nœud
    NodeInfo NodeInfo `json:"nodeInfo"`
}

type PodStats struct {
    TotalPods        int `json:"totalPods"`
    ScheduledPods    int `json:"scheduledPods"`
    UnscheduledPods  int `json:"unscheduledPods"`
}

type CPUStats struct {
    TotalCores     float64 `json:"totalCores"`
    UsedCores      float64 `json:"usedCores"`
    AvailableCores float64 `json:"availableCores"`
    UsagePercent   float64 `json:"usagePercent"`
}

type MemoryStats struct {
    TotalBytes     int64   `json:"totalBytes"`
    UsedBytes      int64   `json:"usedBytes"`
    AvailableBytes int64   `json:"availableBytes"`
    UsagePercent   float64 `json:"usagePercent"`
}

type StorageStats struct {
    TotalBytes     int64   `json:"totalBytes"`
    FreeBytes      int64   `json:"freeBytes"`
    UsedBytes      int64   `json:"usedBytes"`
    UsagePercent   float64 `json:"usagePercent"`
    FileSystem     string  `json:"fileSystem,omitempty"`
}

type NodeInfo struct {
    KernelVersion    string `json:"kernelVersion"`
    OSImage          string `json:"osImage"`
    ContainerRuntime string `json:"containerRuntime"`
    KubeletVersion   string `json:"kubeletVersion"`
    Architecture     string `json:"architecture"`
    OperatingSystem  string `json:"operatingSystem"`
}

// Config représente la configuration de l'agent
type Config struct {
    Agent     AgentConfig     `yaml:"agent"`
    Exporter  ExporterConfig  `yaml:"exporter"`
    Logging   LoggingConfig   `yaml:"logging"`
    K8S       K8SConfig       `yaml:"k8s"`
}

type AgentConfig struct {
    CollectionInterval int `yaml:"collectionInterval"`
    APITimeout         int `yaml:"apiTimeout"`
}

type ExporterConfig struct {
    Type     string         `yaml:"type"`
    HTTP     HTTPConfig     `yaml:"http"`
    Stdout   StdoutConfig   `yaml:"stdout"`
}

type HTTPConfig struct {
    Endpoint string            `yaml:"endpoint"`
    Timeout  int               `yaml:"timeout"`
    Headers  map[string]string `yaml:"headers"`
}

type StdoutConfig struct {
    Enabled bool `yaml:"enabled"`
}

type LoggingConfig struct {
    Level  string `yaml:"level"`
    Format string `yaml:"format"`
}

type K8SConfig struct {
    InCluster  bool   `yaml:"inCluster"`
    Kubeconfig string `yaml:"kubeconfig"`
}