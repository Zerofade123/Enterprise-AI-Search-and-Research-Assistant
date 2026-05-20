# Deployment Architecture

## Overview

The deployment architecture ensures production-grade reliability, scalability, and observability across development, staging, and production environments.

## Environment Strategy

### Development
```yaml
Deployment: Local Docker Compose
Database: PostgreSQL (single instance)
Cache: Redis (single instance)
VectorDB: Milvus (minimal config)
Scale: 1 replica per service
Monitoring: Prometheus + Grafana (optional)
Backups: Disabled
Cost: ~$0 (local)
```

### Staging
```yaml
Deployment: Kubernetes (AWS/GCP/Azure)
Database: RDS Multi-AZ (managed)
Cache: ElastiCache (single node)
VectorDB: Milvus (2 replicas)
Scale: 2-3 replicas per service
Monitoring: Full stack (Prometheus, Grafana, Jaeger, ELK)
Backups: Daily snapshots
Cost: ~$2,000/month
Integration Tests: Automated
Performance Tests: Weekly
```

### Production
```yaml
Deployment: Kubernetes (AWS/GCP/Azure, multi-AZ)
Database: RDS Multi-AZ + read replicas
Cache: ElastiCache cluster mode
VectorDB: Milvus HA cluster
Scale: Auto-scaling (2-50 replicas)
Monitoring: Full observability
Backups: Hourly + cross-region replication
Cost: ~$10,000+/month
Canary Deployments: Enabled
A/B Testing: Supported
Disaster Recovery: Multi-region
```

## Kubernetes Architecture

### Cluster Configuration
```yaml
Kubernetes Version: 1.28+
Container Runtime: containerd
Network: Flannel/Calico
Storage: EBS/GCS Persistent Volumes
Ingress: NGINX Ingress Controller
DNS: CoreDNS
Monitoring: Prometheus + Grafana
```

### Node Groups
```yaml
Master Nodes:
  Count: 3 (HA)
  Machine: t3.large (4GB RAM, 2 CPU)
  Managed: AWS EKS / GCP GKE

Worker Nodes:
  Count: 5-50 (auto-scaling)
  Machine Type: t3.xlarge (16GB RAM, 4 CPU)
  Auto-scaling:
    Min: 5
    Max: 50
    Target CPU: 70%
    Target Memory: 80%

GPU Nodes (Optional):
  Count: 1-5
  Machine: g4dn.xlarge (GPU for embedding generation)
  Auto-scaling: Based on queue depth
```

### Service Deployments

#### API Gateway
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-gateway
  template:
    metadata:
      labels:
        app: api-gateway
    spec:
      containers:
      - name: api-gateway
        image: enterprise-ai/api-gateway:latest
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: url
        resources:
          requests:
            cpu: 200m
            memory: 256Mi
          limits:
            cpu: 500m
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: api-gateway
spec:
  type: LoadBalancer
  selector:
    app: api-gateway
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
```

#### Autoscaling
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-gateway-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-gateway
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

## Database Deployment

### PostgreSQL (RDS)
```yaml
Version: 14+
Instance Class: db.r5.xlarge
Storage: 1000 GB (io1)
IOPS: 3000
Backup Retention: 30 days
Multi-AZ: Enabled
  - Primary: us-east-1a
  - Standby: us-east-1b
  - Read Replica: us-west-2
Encryption:
  - At Rest: AES-256
  - In Transit: SSL/TLS
Parameter Groups:
  - max_connections: 1000
  - shared_buffers: 262144
  - work_mem: 16384
  - effective_cache_size: 3932160
  - log_statement: 'all'
  - log_duration: 'on'
```

### Redis (ElastiCache)
```yaml
Version: 7+
Node Type: cache.r7g.xlarge
Num Cache Nodes: 3 (cluster mode)
Automatic Failover: Enabled
Multi-AZ: Enabled
Encryption:
  - At Rest: Enabled
  - In Transit: Enabled
Backup:
  - Snapshot Frequency: Daily
  - Retention: 7 days
Engine Log Enabled: true
Public Accessibility: false
```

### Milvus Cluster
```yaml
apiVersion: milvus.io/v1beta1
kind: Milvus
metadata:
  name: milvus-cluster
spec:
  mode: cluster
  dependencies:
    etcd:
      endpoints:
      - milvus-etcd:2379
    msgStream:
      type: pulsar
      pulsar:
        endpoints:
        - pulsar://milvus-pulsar:6650
    storage:
      type: minio
      minio:
        bucketName: milvus
        rootPath: file
        useSSL: true
  components:
    coordinator:
      replicas: 3
    proxy:
      replicas: 3
    worker:
      replicas: 5
```

## Container Registry

### Image Management
```bash
# Build and push
docker build -t enterprise-ai/api-gateway:v1.2.3 .
docker push enterprise-ai/api-gateway:v1.2.3

# Image naming convention
REGISTRY/SERVICE:VERSION
Example: gcr.io/enterprise-ai/api-gateway:v1.2.3

# Image scanning
# Vulnerability scanning on push
# Policy: Block high/critical CVEs
```

### Container Security
```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
  capabilities:
    drop:
    - ALL
```

## CI/CD Pipeline

### GitHub Actions Workflow
```yaml
name: Deploy to Production

on:
  push:
    branches: [main]
    tags: ['v*']

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Run tests
      run: make test
    - name: Upload coverage
      uses: codecov/codecov-action@v3

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Build images
      run: make docker-build
    - name: Push to registry
      run: make docker-push
    - name: Scan images
      run: make docker-scan

  deploy-staging:
    needs: build
    runs-on: ubuntu-latest
    environment: staging
    steps:
    - name: Deploy to staging
      run: |
        kubectl set image deployment/api-gateway \
          api-gateway=${{ env.IMAGE }}:${{ github.sha }} \
          -n staging
    - name: Wait for rollout
      run: kubectl rollout status deployment/api-gateway -n staging
    - name: Run smoke tests
      run: make smoke-tests

  deploy-production:
    needs: deploy-staging
    runs-on: ubuntu-latest
    environment: production
    steps:
    - name: Canary deployment (10% traffic)
      run: |
        kubectl patch svc api-gateway -p '{"spec":{"selector":{"version":"canary"}}}'
    - name: Monitor metrics (5 minutes)
      run: |
        ./scripts/monitor-metrics.sh 300
    - name: Full rollout
      run: |
        kubectl patch svc api-gateway -p '{"spec":{"selector":{"version":"stable"}}}'
        kubectl set image deployment/api-gateway \
          api-gateway=${{ env.IMAGE }}:${{ github.sha }}
    - name: Verify deployment
      run: make verify-deployment
```

## Networking

### Ingress Controller
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: api-ingress
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - api.enterprise-ai-search.com
    secretName: tls-cert
  rules:
  - host: api.enterprise-ai-search.com
    http:
      paths:
      - path: /api/v1
        pathType: Prefix
        backend:
          service:
            name: api-gateway
            port:
              number: 80
```

### Network Policies
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-api-gateway
spec:
  podSelector:
    matchLabels:
      app: api-gateway
  policyTypes:
  - Ingress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
```

## Secrets Management

### HashiCorp Vault
```bash
# Store secrets
vault kv put secret/enterprise-ai/database \
  url="postgresql://..." \
  username="user" \
  password="pass"

# Access in deployments
vault kv get -field=url secret/enterprise-ai/database
```

### Kubernetes Secrets Integration
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: db-credentials
type: Opaque
stringData:
  url: postgresql://user:pass@host:5432/db
```

## Backup & Recovery

### Backup Schedule
```yaml
Database:
  - Hourly snapshots (7-day retention)
  - Daily backups (30-day retention)
  - Weekly full backups (90-day retention)
  - Cross-region replication (daily)

Application State:
  - Git commits (permanent)
  - Docker images (90 days)
  - Configuration (version controlled)
```

### Disaster Recovery
```bash
# Point-in-time recovery
aws rds restore-db-instance-to-point-in-time \
  --source-db-instance-identifier prod-db \
  --db-instance-identifier prod-db-recovered \
  --restore-time '2024-05-20T14:00:00Z'

# Failover to standby
aws rds failover-db-cluster --db-cluster-identifier prod-cluster

# Re-deploy from image
kubectl set image deployment/api-gateway \
  api-gateway=enterprise-ai/api-gateway:v1.2.3
```

## Cost Optimization

### Resource Sizing
```yaml
Environment: Production
Estimated Monthly Cost:
  - EKS Cluster: $2,000 (control plane + compute)
  - RDS: $1,500 (db.r5.xlarge)
  - ElastiCache: $500 (r7g.xlarge)
  - Milvus: $1,000 (5 nodes)
  - S3: $1,000 (document storage)
  - NAT Gateway: $500
  - CloudFront CDN: $200
  - Monitoring: $500
  ─────────────────────
  Total: ~$7,200/month
```

### Cost Reduction Strategies
1. **Spot Instances**: 70% of worker nodes on spot (70% discount)
2. **Reserved Capacity**: 50% of baseline on reserved instances (40% discount)
3. **Auto-scaling**: Scale down during off-peak hours
4. **Storage Tiering**: Archive old documents to Glacier
5. **Caching**: Reduce database queries with Redis

## Monitoring & Observability

### Deployment Status
```bash
# Check rollout status
kubectl rollout status deployment/api-gateway -n production

# View pod logs
kubectl logs -f deployment/api-gateway -n production

# Check metrics
kubectl top nodes
kubectl top pods -n production
```

### Alerting
```yaml
Alerts:
  - Pod CrashLooping
  - High CPU (>80% for 5 min)
  - High Memory (>80% for 5 min)
  - High Error Rate (>1% for 1 min)
  - Slow Response Time (p99 > 5s)
  - Database Connection Failures
  - Disk Space Low (<10%)
```