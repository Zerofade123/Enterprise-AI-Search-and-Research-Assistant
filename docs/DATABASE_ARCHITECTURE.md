# Database Architecture

## Overview

The platform uses a distributed database architecture with PostgreSQL as the primary relational database, Redis for caching and queues, and Milvus for vector embeddings.

## Database Technology Stack

### Primary Database: PostgreSQL
- **Version**: 14+
- **Deployment**: RDS Multi-AZ (production)
- **Replication**: Synchronous to standby
- **Backups**: Automated hourly snapshots + WAL archiving
- **Failover**: Automatic to read replica

### Cache Layer: Redis
- **Version**: 7+
- **Deployment**: ElastiCache (cluster mode)
- **Persistence**: RDB snapshots + AOF logs
- **TTL**: Configurable per key type

### Vector Database: Milvus
- **Version**: 2.3+
- **Index Type**: HNSW (Hierarchical Navigable Small World)
- **Deployment**: Kubernetes StatefulSet
- **Replication**: Built-in HA

## Multi-Tenancy Strategy

### Row-Level Security (RLS)

```sql
-- Enable RLS on all tables
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;

-- Create policy for workspace isolation
CREATE POLICY workspace_isolation ON documents
  USING (workspace_id = current_setting('app.workspace_id')::uuid);

-- Before each request, set workspace context
SET app.workspace_id = 'workspace-uuid';
```

### Workspace Context

Every query includes workspace isolation:

```sql
SELECT * FROM documents 
WHERE workspace_id = current_workspace_id();
```

## Schema Design

### Core Tables

#### Users
```sql
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  email_verified BOOLEAN DEFAULT FALSE,
  mfa_enabled BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
```

#### Workspaces
```sql
CREATE TABLE workspaces (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  slug VARCHAR(255) NOT NULL UNIQUE,
  owner_id UUID NOT NULL REFERENCES users(id),
  status VARCHAR(50) DEFAULT 'active',
  tier VARCHAR(50) DEFAULT 'free',
  settings JSONB DEFAULT '{}',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_workspaces_owner ON workspaces(owner_id);
CREATE INDEX idx_workspaces_slug ON workspaces(slug) WHERE deleted_at IS NULL;
```

#### Workspace Members
```sql
CREATE TABLE workspace_members (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workspace_id UUID NOT NULL REFERENCES workspaces(id),
  user_id UUID NOT NULL REFERENCES users(id),
  role VARCHAR(50) NOT NULL CHECK (role IN ('owner', 'admin', 'editor', 'viewer')),
  permissions JSONB DEFAULT '{}',
  invited_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  joined_at TIMESTAMP NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(workspace_id, user_id)
);

CREATE INDEX idx_workspace_members_user ON workspace_members(user_id);
CREATE INDEX idx_workspace_members_workspace ON workspace_members(workspace_id);
```

#### Documents
```sql
CREATE TABLE documents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workspace_id UUID NOT NULL REFERENCES workspaces(id),
  name VARCHAR(255) NOT NULL,
  file_size BIGINT NOT NULL,
  mime_type VARCHAR(100),
  status VARCHAR(50) DEFAULT 'uploading' CHECK (status IN ('uploading', 'processing', 'ready', 'failed')),
  s3_key VARCHAR(500) NOT NULL,
  s3_bucket VARCHAR(255),
  created_by UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_documents_workspace ON documents(workspace_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_created_by ON documents(created_by);
CREATE INDEX idx_documents_status ON documents(status);
```

#### Document Chunks
```sql
CREATE TABLE document_chunks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
  chunk_number INT NOT NULL,
  content TEXT NOT NULL,
  token_count INT,
  metadata JSONB DEFAULT '{}',
  embedding_id UUID NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(document_id, chunk_number)
);

CREATE INDEX idx_chunks_document ON document_chunks(document_id);
CREATE INDEX idx_chunks_embedding ON document_chunks(embedding_id);
```

#### Search Queries
```sql
CREATE TABLE search_queries (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workspace_id UUID NOT NULL REFERENCES workspaces(id),
  user_id UUID NOT NULL REFERENCES users(id),
  query TEXT NOT NULL,
  filters JSONB DEFAULT '{}',
  result_count INT,
  execution_time_ms INT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_search_queries_workspace ON search_queries(workspace_id);
CREATE INDEX idx_search_queries_user ON search_queries(user_id);
CREATE INDEX idx_search_queries_created ON search_queries(created_at DESC);
```

#### Audit Logs
```sql
CREATE TABLE audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workspace_id UUID NOT NULL REFERENCES workspaces(id),
  user_id UUID REFERENCES users(id),
  action VARCHAR(100) NOT NULL,
  resource_type VARCHAR(100) NOT NULL,
  resource_id UUID,
  status VARCHAR(50) DEFAULT 'success',
  ip_address INET,
  user_agent TEXT,
  changes JSONB DEFAULT '{}',
  error_message TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_workspace ON audit_logs(workspace_id);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
```

## Indexing Strategy

### B-Tree Indexes (Default)
```sql
-- Foreign key lookups
CREATE INDEX idx_documents_workspace ON documents(workspace_id);
CREATE INDEX idx_workspace_members_user ON workspace_members(user_id);

-- Time-based queries
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at DESC);
```

### BRIN Indexes (Time-series)
```sql
-- For large audit log tables
CREATE INDEX idx_audit_logs_created_brin ON audit_logs USING BRIN (created_at);

-- For search queries
CREATE INDEX idx_search_queries_created_brin ON search_queries USING BRIN (created_at);
```

### Full-Text Search Indexes
```sql
-- For document search
CREATE INDEX idx_documents_name_fts ON documents 
  USING GIN (to_tsvector('english', name));

CREATE INDEX idx_chunks_content_fts ON document_chunks 
  USING GIN (to_tsvector('english', content));
```

### JSONB Indexes
```sql
-- For flexible metadata queries
CREATE INDEX idx_document_metadata_gin ON documents USING GIN (settings);
```

## Query Optimization

### Connection Pooling
```yaml
# PgBouncer configuration
[databases]
enterprise_ai = host=rds.amazonaws.com port=5432 dbname=enterprise_ai

[pgbouncer]
pool_mode = transaction
max_client_conn = 1000
default_pool_size = 25
min_pool_size = 10
reserve_pool_size = 5
```

### Query Patterns

```go
// Efficient paginated query
SELECT * FROM documents 
WHERE workspace_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

// Batch insertion with UPSERT
INSERT INTO search_queries (workspace_id, user_id, query, created_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT DO UPDATE SET ...;
```

## Caching Strategy

### Redis Key Naming
```
user:{user_id}                    # User cache (TTL: 1h)
workspace:{workspace_id}          # Workspace cache (TTL: 30m)
workspace:{workspace_id}:members  # Members list (TTL: 15m)
document:{doc_id}:metadata        # Document metadata (TTL: 1h)
search:query:{hash}:results       # Search results (TTL: 24h)
ratelimit:{user_id}               # Rate limit counter (TTL: 1m)
session:{token_hash}              # Session data (TTL: 7d)
```

### Cache Invalidation
```go
// On document update
redis.Del(ctx, fmt.Sprintf("document:%s:metadata", docID))

// On workspace member change
redis.Del(ctx, fmt.Sprintf("workspace:%s:members", workspaceID))
```

## Vector Database Design

### Milvus Collection Schema
```python
collection_name = "document_embeddings"
fields = [
    FieldSchema(name="chunk_id", dtype=DataType.VARCHAR, is_primary=True),
    FieldSchema(name="workspace_id", dtype=DataType.VARCHAR),
    FieldSchema(name="document_id", dtype=DataType.VARCHAR),
    FieldSchema(name="embedding", dtype=DataType.FLOAT_VECTOR, dim=1536),
    FieldSchema(name="metadata", dtype=DataType.JSON),
    FieldSchema(name="created_at", dtype=DataType.INT64),
]

index_params = {
    "metric_type": "cosine",
    "index_type": "HNSW",
    "params": {"M": 16, "efConstruction": 200}
}
```

### Vector Search Query
```python
search_params = {"metric_type": "cosine", "params": {"ef": 64}}

results = collection.search(
    data=[query_embedding],
    anns_field="embedding",
    param=search_params,
    limit=10,
    expr=f'workspace_id == "{workspace_id}"'
)
```

## Data Retention Policies

### Audit Logs
- **Default**: 90 days
- **Enterprise**: 1 year
- **Archive**: S3 Glacier after retention

### Search History
- **Default**: 30 days
- **User Delete**: Cascade delete

### Session Data
- **TTL**: 7 days
- **Auto-cleanup**: Redis TTL

## Backup & Recovery

### Backup Schedule
```yaml
Backups:
  - Type: Continuous
    Method: WAL archiving to S3
    Retention: 30 days
    
  - Type: Hourly Snapshots
    Method: RDS automated backup
    Retention: 7 days
    
  - Type: Daily Snapshots
    Method: RDS manual snapshot
    Retention: 30 days
    
  - Type: Weekly Full Backup
    Method: pg_dump to S3
    Retention: 90 days
```

### Recovery Procedures
```bash
# Point-in-time recovery
aws rds restore-db-instance-to-point-in-time \
  --source-db-instance-identifier prod-db \
  --db-instance-identifier prod-db-recovered \
  --restore-time 2024-05-20T10:00:00Z

# From backup
aws rds restore-db-instance-from-db-snapshot \
  --db-instance-identifier prod-db-restored \
  --db-snapshot-identifier snapshot-id
```

## Monitoring & Maintenance

### Key Metrics
- Connection count
- Query duration (p50, p95, p99)
- Cache hit rate
- Replication lag
- Disk usage
- Index bloat

### Maintenance Tasks
```sql
-- Weekly: Vacuum and analyze
VACUUM ANALYZE documents;
VACUUM ANALYZE audit_logs;

-- Monthly: Index rebuild
REINDEX INDEX CONCURRENTLY idx_documents_workspace;

-- Quarterly: Table bloat analysis
SELECT schemaname, tablename, 
       pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) 
FROM pg_tables 
WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```