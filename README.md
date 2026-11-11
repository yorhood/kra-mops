# K8s Monitoring Agent

Agent de monitoring Kubernetes écrit en Go pour collecter les métriques des nœuds.

## Métriques collectées

- **Pods**: Nombre total, pods schedulés, pods non schedulés
- **CPU**: Cores totaux, utilisés, disponibles et pourcentage d'utilisation
- **Mémoire**: Octets totaux, utilisés, disponibles et pourcentage d'utilisation  
- **Stockage**: Espace total, libre, utilisé et pourcentage d'utilisation
- **Informations nœud**: Version noyau, OS, runtime container, etc.

## Déploiement

1. Build l'image Docker:
```bash
docker build -t k8s-monitoring-agent:latest .