# Observability Strategy

## Overview

Observability is fundamental to operating a production-grade enterprise platform. The strategy encompasses logging, metrics, tracing, and alerting across all system components.

## Three Pillars of Observability

### 1. Logs

**Purpose**: Understand what happened

#### Structured Logging

All logs use JSON format for easy parsing and analysis:

```go
logger.Info("document_processed",
  zap.String("document_id", docID),
  zap.String("workspace_id", workspaceID),
  zap.String("request_id", requestID),
  zap.Int64("duration_ms", duration),
  zap.String("status", "success"),
  zap.String("trace_id", traceID),
)

// Output:
// {
//   "timestamp": "2024-05-20T10:30:00.123Z",
//   "level": "info",
//   "message": "document_processed",
//   "document_id": "doc-123",
//   "workspace_id": "ws-123",
//   "request_id": "req-456",
//   "duration_ms": 1234,
//   "status": "success",
//   "trace_id": "trace-789"
// }
```

#### Log Levels

```
DEBUG    - Detailed diagnostic information (development only)
INFO     - General informational messages
WARN     - Warning messages (potential issues)
ERROR    - Error messages (action failed)
CRITICAL - Critical errors (system failure imminent)
```

#### Log Retention

```yaml
Hot Storage (ELK Stack):
  - Duration: 7 days
  - Index per day: enterprise-ai-logs-2024-05-20
  - Retention policy: Delete after 7 days
  
Warm Storage (S3):
  - Duration: 30 days
  - Format: Compressed JSON (gzip)
  - Location: s3://enterprise-ai-logs/archive/
  
Cold Storage (Glacier):
  - Duration: 90 days
  - Archive format: Compressed JSON + Parquet
  - Location: s3://enterprise-ai-logs/glacier/
```

#### Log Sampling

For high-volume services, use sampling to reduce costs:

```go
if shouldSample(requestID) {
  logger.Debug("request_details",
    zap.Any("headers", headers),
    zap.Any("body", body),
  )
}

func shouldSample(requestID string) bool {
  // Sample 1% of traffic
  hash := hashString(requestID)
  return hash%100 < 1
}
```

#### Log Aggregation (ELK Stack)

```yaml
ELK Configuration:
  Elasticsearch:
    Version: 8.0+
    Nodes: 3 (cluster mode)
    Storage: 500GB SSD
    Retention: 7 days (hot)
    Shards: 5
    Replicas: 2
    
  Logstash:
    Input: Docker logs, syslog
    Filter: Parse JSON, enrich with metadata
    Output: Elasticsearch indices
    
  Kibana:
    Dashboards: Pre-built for common queries
    Alerts: Query-based alerts
    Access Control: Role-based
```

### 2. Metrics

**Purpose**: Understand system behavior over time

#### Metrics Collection

```yaml
Prometheus Configuration:
  Version: 2.40+
  Scrape Interval: 15 seconds
  Retention: 30 days
  Storage: 100GB (time-series database)
  
  Targets:
    - API Gateway (:9090/metrics)
    - All microservices (:9090/metrics)
    - PostgreSQL exporter
    - Redis exporter
    - Kubernetes metrics
```

#### Application Metrics

```go
// HTTP metrics
var (
  httpRequestsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
      Name: "http_requests_total",
      Help: "Total HTTP requests",
    },
    []string{"method", "path", "status"},
  )
  
  httpRequestDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
      Name: "http_request_duration_seconds",
      Help: "HTTP request latency",
      Buckets: []float64{.01, .05, .1, .5, 1, 5},
    },
    []string{"method", "path"},
  )
)

// Business metrics
var (
  documentsUploaded = prometheus.NewCounterVec(
    prometheus.CounterOpts{
      Name: "documents_uploaded_total",
      Help: "Total documents uploaded",
    },
    []string{"workspace_id", "file_type"},
  )
  
  searchQueries = prometheus.NewCounterVec(
    prometheus.CounterOpts{
      Name: "search_queries_total",
      Help: "Total search queries",
    },
    []string{"workspace_id", "query_type"},
  )
)
```

#### Metric Types

```
Counter:     Always increases (requests_total, errors_total)
Gauge:       Can go up or down (memory_usage, active_connections)
Histogram:   Distribution over time (request_duration, file_size)
Summary:     Quantiles (p50, p95, p99 latencies)
```

#### Key Metrics by Service

**API Gateway:**
- Request rate (req/sec)
- Error rate (%)
- Response latency (p50, p95, p99)
- Concurrent connections
- Cache hit rate (%)

**Document Service:**
- Documents uploaded (count)
- Average file size
- Processing success rate
- Storage used (GB)
- API response time

**Search Service:**
- Search queries per second
- Search latency (p50, p95, p99)
- Result quality score
- Cache hit rate
- Vector DB query time

**Auth Service:**
- Login attempts
- Login success rate
- Token validation time
- MFA usage rate
- API key usage

#### Dashboards

```yaml
Grafana Dashboards:
  
  System Overview:
    - Node CPU usage
    - Node memory usage
    - Network I/O
    - Disk usage
    - Pod count
    
  API Performance:
    - Request rate by service
    - Error rate by service
    - Latency heatmap
    - Top slow endpoints
    - Cache hit rates
    
  Business Metrics:
    - Daily active users
    - Documents uploaded
    - Search queries
    - User retention
    - Feature usage
    
  Database:
    - Connections (used/max)
    - Query latency
    - Slow queries
    - Replication lag
    - Cache efficiency
    
  Infrastructure:
    - Kubernetes node status
    - Pod restarts
    - PVC usage
    - Network policies
```

### 3. Tracing

**Purpose**: Understand request flow across services

#### Distributed Tracing Setup

```yaml
Jaeger Configuration:
  Version: 1.40+
  Deployment: All-in-one (or distributed)
  Backend: Elasticsearch
  Sampling:
    Type: Adaptive
    Initial Rate: 0.1% (high volume)
    Max Traces per Second: 1000
  Retention: 72 hours
```

#### Trace Context Propagation

```go
import "go.opentelemetry.io/otel"

// Generate trace ID for incoming request
tracer := otel.Tracer("api-gateway")
ctx, span := tracer.Start(ctx, "ProcessRequest")
defer span.End()

// Extract trace context
traceID := span.SpanContext().TraceID().String()
spanID := span.SpanContext().SpanID().String()

// Propagate to downstream services
req.Header.Set("traceparent", fmt.Sprintf("00-%s-%s-01", traceID, spanID))
```

#### Span Types

```
Service Span:  High-level service operation (e.g., ProcessDocument)
Database Span: Query execution (e.g., Query: SELECT * FROM documents)
HTTP Span:     External API calls (e.g., POST /search)
GRPC Span:     gRPC calls between services
Cache Span:    Cache operations (get, set, delete)
```

#### Example Trace

```
Request: POST /api/v1/search
  TraceID: 550e8400-e29b-41d4-a716-446655440000
  
  ├── Span: api-gateway (50ms total)
  │   ├── Span: auth-validation (5ms)
  │   ├── Span: query-enhancement (10ms)
  │   └── Span: search-service-call (35ms)
  │
  ├── Span: search-service (35ms)
  │   ├── Span: vector-embedding (15ms)
  │   ├── Span: vectordb-query (15ms)
  │   └── Span: result-formatting (5ms)
  │
  └── Span: notification-service (5ms)
      └── Span: websocket-publish (5ms)
```

## Alerting Strategy

### Alert Rules

#### Severity Levels

```yaml
Critical:
  - Service down (no healthy pods)
  - Database connection failure
  - Disk full (>95%)
  - Memory critical (>95%)
  - High error rate (>5% for 2 min)
  
High:
  - Pod restart loop
  - High CPU (>80% for 5 min)
  - High memory (>80% for 5 min)
  - Slow queries (>5s for 5 min)
  - Deployment lag (>1 min)
  
Medium:
  - Pod pending (>5 min)
  - High latency (p99 > 5s)
  - Cache miss rate (>50%)
  - Failed backup
  
Low:
  - Deprecated API usage
  - Unused resources
  - Configuration drift
```

#### Alert Rules (Prometheus)

```yaml
groups:
- name: application
  rules:
  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
    for: 2m
    annotations:
      severity: critical
      summary: "High error rate detected"
      
  - alert: SlowQueries
    expr: histogram_quantile(0.99, http_request_duration_seconds) > 5
    for: 5m
    annotations:
      severity: high
      summary: "99th percentile latency > 5s"
      
  - alert: HighCPUUsage
    expr: node_cpu_seconds_total > 0.8
    for: 5m
    annotations:
      severity: high
      summary: "High CPU usage"
```

### Notification Channels

```yaml
Alerting Configuration:
  
  Critical Alerts:
    - PagerDuty (immediate escalation)
    - SMS (to on-call)
    - Slack #incidents
    - Email
    
  High Alerts:
    - Slack #alerts
    - Email
    
  Medium Alerts:
    - Slack #alerts
    
  Low Alerts:
    - Email (daily digest)
```

### On-Call Management

```yaml
PagerDuty Integration:
  Services:
    - Production API Gateway
    - Production Database
    - Production Search
    - Production Auth
    
  Escalation Policies:
    - Level 1: On-call engineer (5 min)
    - Level 2: Lead engineer (10 min)
    - Level 3: Team lead (15 min)
    
  Incident Response:
    - Auto-page on critical alert
    - Create incident ticket
    - Open war room (Zoom)
    - Post-mortem within 24h
```

## SLOs & SLIs

### Service Level Objectives

```yaml
API Gateway:
  Availability: 99.99% (52.6 min/year downtime)
  Latency (p99): < 500ms
  Error Rate: < 0.1%
  
Document Service:
  Availability: 99.95% (2.2 hours/year)
  Upload Success Rate: > 99.9%
  Processing Time: < 60 seconds (99%)
  
Search Service:
  Availability: 99.95%
  Query Latency (p99): < 2 seconds
  Result Quality: > 0.85 (internal metrics)
  
Auth Service:
  Availability: 99.99%
  Login Success Rate: > 99.5%
  Token Validation Time: < 10ms (p99)
```

### Service Level Indicators

```yaml
Availability SLI:
  Definition: |(successful_requests) / (total_requests)|
  Calculation: rate(http_requests_total{status!~"5.."}[5m]) / rate(http_requests_total[5m])
  
Latency SLI:
  Definition: |(requests < 500ms) / (total_requests)|
  Calculation: histogram_quantile(0.99, http_request_duration_seconds)
  
Error Rate SLI:
  Definition: |(error_requests) / (total_requests)|
  Calculation: rate(http_requests_total{status=~"5.."}[5m])
```

## Error Tracking

### Error Aggregation

```yaml
Sentry Configuration:
  Organization: enterprise-ai
  Projects: One per service
  Environments: staging, production
  
  Issue Grouping:
    - By fingerprint (default)
    - By component + error message
    - By URL pattern
    
  Release Tracking:
    - Link commits to errors
    - Track error regressions
    - Monitor deploy impact
    
  Sampling:
    - Errors: 100% sampling
    - Transactions: 10% sampling
```

### Error Categories

```
Application Errors:
  - Validation errors (4xx)
  - Business logic errors
  - Integration errors
  
Infrastructure Errors:
  - Database connection errors
  - Service unavailable (5xx)
  - Timeout errors
  
Client Errors:
  - Bad request (400)
  - Unauthorized (401)
  - Forbidden (403)
  - Not found (404)
```

## Performance Profiling

### Continuous Profiling

```yaml
Pyroscope Configuration:
  Agent: Installed in all services
  Sampling Rate: 100 Hz (CPU)
  Retention: 30 days
  
  Profiles:
    - CPU profile (function execution time)
    - Memory profile (heap allocation)
    - Goroutine profile (blocking)
    - Mutex profile (lock contention)
```

### Performance Optimization

```
Optimization Workflow:
1. Identify slow operation in traces/metrics
2. Profile with Pyroscope
3. Analyze flamegraph
4. Implement optimization
5. Verify improvement with benchmarks
6. Roll out with monitoring
```

## Cost Monitoring

### Cloud Cost Tracking

```yaml
AWS Cost Explorer:
  - Daily cost tracking
  - Cost by service (EC2, RDS, S3)
  - Cost anomaly detection
  - Budget alerts (>10% variance)
  
Estimated Monthly Breakdown:
  - Compute (EKS): $2,000 (30%)
  - Database (RDS): $1,500 (22%)
  - Storage (S3): $1,000 (15%)
  - Cache (ElastiCache): $500 (7%)
  - Monitoring: $500 (7%)
  - Other: $1,200 (19%)
  - Total: $7,200/month
```

## Health Checks

### Liveness Probes

```go
// Kubernetes liveness probe
// If this fails, pod is restarted
func healthzHandler(w http.ResponseWriter, r *http.Request) {
  if !isHealthy() {
    w.WriteHeader(http.StatusServiceUnavailable)
    return
  }
  w.WriteHeader(http.StatusOK)
  json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Configuration
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

### Readiness Probes

```go
// Kubernetes readiness probe
// If this fails, pod is removed from service
func readyHandler(w http.ResponseWriter, r *http.Request) {
  checks := map[string]bool{
    "database": isDBConnected(),
    "cache": isCacheConnected(),
    "vectordb": isVectorDBConnected(),
  }
  
  allReady := true
  for _, ready := range checks {
    if !ready {
      allReady = false
      break
    }
  }
  
  if !allReady {
    w.WriteHeader(http.StatusServiceUnavailable)
  } else {
    w.WriteHeader(http.StatusOK)
  }
  json.NewEncoder(w).Encode(checks)
}

// Configuration
readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 2
```

## Compliance Monitoring

### Audit Trail

```yaml
Audit Events Tracked:
  - User authentication (login/logout)
  - Data access (read/write/delete)
  - Permission changes
  - Admin actions
  - API key creation/revocation
  - Workspace changes
  
Retention: 90 days (configurable per org)
Compliance: GDPR, SOC2, HIPAA
```

## Future Enhancements

- [ ] OpenTelemetry full migration
- [ ] Custom metrics for ML model performance
- [ ] Real-time anomaly detection
- [ ] eBPF-based system call tracing
- [ ] Browser monitoring for frontend