# Repository Structure

## Overview

The repository follows a monorepo structure with clear separation between frontend, backend services, and shared utilities. This enables independent deployment while maintaining code organization.

## Directory Structure

```
enterprise-ai-search-assistant/
├── README.md                              # Project overview
├── LICENSE                                # Apache 2.0 License
├── .gitignore                             # Git ignore rules
│
├── docs/                                  # Architecture and design docs
│   ├── ARCHITECTURE.md                    # System architecture
│   ├── REPOSITORY_STRUCTURE.md           # This file
│   ├── SERVICE_BOUNDARIES.md             # Service definitions
│   ├── DATABASE_ARCHITECTURE.md          # Database design
│   ├── API_STRUCTURE.md                  # API definitions
│   ├── AUTHENTICATION_ARCHITECTURE.md    # Auth flows
│   ├── AI_PIPELINE_ARCHITECTURE.md       # AI pipeline design
│   ├── DEPLOYMENT_ARCHITECTURE.md        # Deployment guide
│   ├── OBSERVABILITY_STRATEGY.md         # Monitoring strategy
│   ├── SECURITY_ARCHITECTURE.md          # Security design
│   ├── CONTRIBUTING.md                   # Contribution guidelines
│   ├── DEVELOPMENT.md                    # Development setup
│   └── DEPLOYMENT.md                     # Deployment procedures
│
├── frontend/                              # Next.js frontend application
│   ├── package.json                      # Node dependencies
│   ├── tsconfig.json                     # TypeScript config
│   ├── tailwind.config.js                # Tailwind CSS config
│   ├── jest.config.js                    # Jest testing config
│   ├── next.config.js                    # Next.js config
│   ├── .env.example                      # Environment variables example
│   │
│   ├── public/                           # Static assets
│   │   ├── images/
│   │   ├── icons/
│   │   └── favicon.ico
│   │
│   ├── src/
│   │   ├── app/                          # Next.js app directory
│   │   │   ├── layout.tsx               # Root layout
│   │   │   ├── page.tsx                 # Home page
│   │   │   │
│   │   │   ├── (auth)/                  # Auth routes
│   │   │   │   ├── login/page.tsx
│   │   │   │   ├── register/page.tsx
│   │   │   │   ├── callback/page.tsx
│   │   │   │   └── logout/page.tsx
│   │   │   │
│   │   │   ├── (workspace)/             # Authenticated routes
│   │   │   │   ├── layout.tsx           # Workspace layout
│   │   │   │   ├── dashboard/page.tsx
│   │   │   │   ├── workspace/           # Workspace management
│   │   │   │   │   └── [id]/
│   │   │   │   │       ├── layout.tsx
│   │   │   │   │       ├── page.tsx
│   │   │   │   │       ├── settings/
│   │   │   │   │       ├── members/
│   │   │   │   │       └── documents/
│   │   │   │   ├── documents/           # Document management
│   │   │   │   │   ├── page.tsx
│   │   │   │   │   └── [id]/page.tsx
│   │   │   │   ├── search/              # Search interface
│   │   │   │   │   ├── page.tsx
│   │   │   │   │   └── [id]/page.tsx
│   │   │   │   ├── research/            # Research workspace
│   │   │   │   │   ├── page.tsx
│   │   │   │   │   └── [id]/page.tsx
│   │   │   │   ├── settings/            # User settings
│   │   │   │   │   ├── layout.tsx
│   │   │   │   │   ├── profile/
│   │   │   │   │   ├── preferences/
│   │   │   │   │   └── billing/
│   │   │   │   └── admin/               # Admin dashboard
│   │   │   │       ├── layout.tsx
│   │   │   │       ├── users/
│   │   │   │       ├── workspaces/
│   │   │   │       └── audit-logs/
│   │   │   │
│   │   │   ├── api/                     # API routes (edge functions)
│   │   │   │   ├── auth/
│   │   │   │   ├── documents/
│   │   │   │   ├── search/
│   │   │   │   └── webhooks/
│   │   │   │
│   │   │   └── error.tsx                # Error boundary
│   │   │
│   │   ├── components/                  # Reusable React components
│   │   │   ├── common/                  # Common components
│   │   │   │   ├── Header.tsx
│   │   │   │   ├── Sidebar.tsx
│   │   │   │   ├── Footer.tsx
│   │   │   │   ├── LoadingSpinner.tsx
│   │   │   │   ├── ErrorBoundary.tsx
│   │   │   │   ├── Pagination.tsx
│   │   │   │   ├── Modal.tsx
│   │   │   │   ├── Button.tsx
│   │   │   │   ├── Input.tsx
│   │   │   │   ├── Card.tsx
│   │   │   │   └── Badge.tsx
│   │   │   │
│   │   │   ├── document/                # Document-related components
│   │   │   │   ├── DocumentUpload.tsx
│   │   │   │   ├── DocumentCard.tsx
│   │   │   │   ├── DocumentList.tsx
│   │   │   │   ├── DocumentViewer.tsx
│   │   │   │   └── DocumentPreview.tsx
│   │   │   │
│   │   │   ├── search/                  # Search components
│   │   │   │   ├── SearchBar.tsx
│   │   │   │   ├── SearchResults.tsx
│   │   │   │   ├── ResultCard.tsx
│   │   │   │   ├── FilterPanel.tsx
│   │   │   │   └── SearchHistory.tsx
│   │   │   │
│   │   │   ├── research/                # Research components
│   │   │   │   ├── ResearchWorkspace.tsx
│   │   │   │   ├── ResearchChat.tsx
│   │   │   │   ├── AgentPanel.tsx
│   │   │   │   ├── ContextPanel.tsx
│   │   │   │   └── ResearchHistory.tsx
│   │   │   │
│   │   │   ├── workspace/               # Workspace components
│   │   │   │   ├── WorkspaceSelector.tsx
│   │   │   │   ├── WorkspaceSettings.tsx
│   │   │   │   ├── MemberList.tsx
│   │   │   │   └── RoleSelector.tsx
│   │   │   │
│   │   │   ├── auth/                    # Auth components
│   │   │   │   ├── LoginForm.tsx
│   │   │   │   ├── RegisterForm.tsx
│   │   │   │   ├── MFASetup.tsx
│   │   │   │   └── SSO.tsx
│   │   │   │
│   │   │   ├── admin/                   # Admin components
│   │   │   │   ├── UserManagement.tsx
│   │   │   │   ├── AuditLogViewer.tsx
│   │   │   │   ├── SystemStats.tsx
│   │   │   │   └── RoleManagement.tsx
│   │   │   │
│   │   │   ├── dashboard/               # Dashboard components
│   │   │   │   ├── StatCard.tsx
│   │   │   │   ├── Charts.tsx
│   │   │   │   ├── QuickActions.tsx
│   │   │   │   └── ActivityFeed.tsx
│   │   │   │
│   │   │   └── layout/                  # Layout components
│   │   │       ├── MainLayout.tsx
│   │   │       ├── AuthLayout.tsx
│   │   │       └── AdminLayout.tsx
│   │   │
│   │   ├── hooks/                       # Custom React hooks
│   │   │   ├── useAuth.ts              # Auth state hook
│   │   │   ├── useWorkspace.ts         # Workspace hook
│   │   │   ├── useSearch.ts            # Search hook
│   │   │   ├── useNotification.ts      # Notification hook
│   │   │   ├── useWebSocket.ts         # WebSocket hook
│   │   │   ├── usePagination.ts        # Pagination hook
│   │   │   ├── useDebounce.ts          # Debounce hook
│   │   │   └── useLocalStorage.ts      # Local storage hook
│   │   │
│   │   ├── contexts/                   # React contexts
│   │   │   ├── AuthContext.tsx         # Auth context
│   │   │   ├── WorkspaceContext.tsx    # Workspace context
│   │   │   ├── NotificationContext.tsx # Notification context
│   │   │   └── ThemeContext.tsx        # Theme context
│   │   │
│   │   ├── services/                   # API service clients
│   │   │   ├── api.ts                  # API client configuration
│   │   │   ├── auth.ts                 # Auth API client
│   │   │   ├── documents.ts            # Document API client
│   │   │   ├── search.ts               # Search API client
│   │   │   ├── workspace.ts            # Workspace API client
│   │   │   ├── users.ts                # User API client
│   │   │   ├── admin.ts                # Admin API client
│   │   │   └── websocket.ts            # WebSocket service
│   │   │
│   │   ├── stores/                     # Zustand state stores
│   │   │   ├── authStore.ts            # Auth store
│   │   │   ├── workspaceStore.ts       # Workspace store
│   │   │   ├── documentStore.ts        # Document store
│   │   │   ├── searchStore.ts          # Search store
│   │   │   ├── uiStore.ts              # UI state store
│   │   │   └── notificationStore.ts    # Notification store
│   │   │
│   │   ├── utils/                      # Utility functions
│   │   │   ├── auth.ts                 # Auth utilities
│   │   │   ├── validation.ts           # Form validation
│   │   │   ├── formatting.ts           # Data formatting
│   │   │   ├── date.ts                 # Date utilities
│   │   │   ├── file.ts                 # File utilities
│   │   │   ├── string.ts               # String utilities
│   │   │   ├── http.ts                 # HTTP utilities
│   │   │   ├── analytics.ts            # Analytics tracking
│   │   │   └── constants.ts            # App constants
│   │   │
│   │   ├── types/                      # TypeScript types
│   │   │   ├── index.ts                # Main types export
│   │   │   ├── api.ts                  # API types
│   │   │   ├── auth.ts                 # Auth types
│   │   │   ├── documents.ts            # Document types
│   │   │   ├── search.ts               # Search types
│   │   │   ├── workspace.ts            # Workspace types
│   │   │   ├── user.ts                 # User types
│   │   │   ├── research.ts             # Research types
│   │   │   └── common.ts               # Common types
│   │   │
│   │   ├── middleware.ts               # Next.js middleware
│   │   ├── globals.css                 # Global styles
│   │   └── globals.ts                  # Global configurations
│   │
│   ├── __tests__/                       # Frontend tests
│   │   ├── unit/
│   │   ├── integration/
│   │   └── e2e/
│   │
│   └── .env.local                       # Local environment (gitignored)
│
├── backend/                              # Go backend services
│   ├── go.mod                           # Go module definition
│   ├── go.sum                           # Go dependencies lock
│   ├── Dockerfile                       # Multi-stage docker build
│   ├── Makefile                         # Build automation
│   │
│   ├── cmd/                             # Service entrypoints
│   │   ├── api-gateway/
│   │   │   └── main.go
│   │   ├── auth-service/
│   │   │   └── main.go
│   │   ├── document-service/
│   │   │   └── main.go
│   │   ├── search-service/
│   │   │   └── main.go
│   │   ├── workspace-service/
│   │   │   └── main.go
│   │   ├── user-service/
│   │   │   └── main.go
│   │   ├── notification-service/
│   │   │   └── main.go
│   │   └── audit-service/
│   │       └── main.go
│   │
│   ├── internal/                        # Internal packages
│   │   ├── middleware/                  # HTTP middleware
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   ├── ratelimit.go
│   │   │   ├── logging.go
│   │   │   ├── tracing.go
│   │   │   └── errorhandler.go
│   │   │
│   │   ├── auth/                        # Auth service
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   ├── jwt.go
│   │   │   ├── oauth.go
│   │   │   ├── mfa.go
│   │   │   └── types.go
│   │   │
│   │   ├── document/                    # Document service
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   ├── storage.go
│   │   │   └── types.go
│   │   │
│   │   ├── search/                      # Search service
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   ├── vectordb.go
│   │   │   └── types.go
│   │   │
│   │   ├── workspace/                   # Workspace service
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   └── types.go
│   │   │
│   │   ├── user/                        # User service
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   └── types.go
│   │   │
│   │   ├── notification/                # Notification service
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── websocket.go
│   │   │   ├── repository.go
│   │   │   └── types.go
│   │   │
│   │   ├── audit/                       # Audit service
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   └── types.go
│   │   │
│   │   ├── database/                    # Database utilities
│   │   │   ├── postgres.go              # PostgreSQL client
│   │   │   ├── migration.go             # Schema migrations
│   │   │   ├── transactions.go          # Transaction handling
│   │   │   └── types.go
│   │   │
│   │   ├── cache/                       # Cache utilities
│   │   │   ├── redis.go                 # Redis client
│   │   │   ├── manager.go               # Cache manager
│   │   │   └── types.go
│   │   │
│   │   ├── queue/                       # Message queue
│   │   │   ├── rabbitmq.go              # RabbitMQ client
│   │   │   ├── producer.go              # Message producer
│   │   │   ├── consumer.go              # Message consumer
│   │   │   └── types.go
│   │   │
│   │   ├── storage/                     # Object storage
│   │   │   ├── s3.go                    # S3 client
│   │   │   ├── manager.go               # Storage manager
│   │   │   └── types.go
│   │   │
│   │   ├── vectordb/                    # Vector database
│   │   │   ├── milvus.go                # Milvus client
│   │   │   ├── manager.go               # VectorDB manager
│   │   │   └── types.go
│   │   │
│   │   ├── observability/               # Observability
│   │   │   ├── logger.go                # Structured logging
│   │   │   ├── metrics.go               # Metrics collection
│   │   │   ├── tracing.go               # Distributed tracing
│   │   │   └── healthcheck.go           # Health checks
│   │   │
│   │   ├── security/                    # Security utilities
│   │   │   ├── vault.go                 # Secrets management
│   │   │   ├── encryption.go            # Encryption utilities
│   │   │   ├── rbac.go                  # RBAC enforcement
│   │   │   └── audit.go                 # Audit logging
│   │   │
│   │   ├── grpc/                        # gRPC definitions
│   │   │   ├── auth/
│   │   │   │   ├── auth.pb.go
│   │   │   │   ├── auth.pb.gw.go
│   │   │   │   └── auth_grpc.pb.go
│   │   │   ├── document/
│   │   │   ├── search/
│   │   │   ├── workspace/
│   │   │   └── user/
│   │   │
│   │   ├── config/                      # Configuration
│   │   │   ├── config.go                # Config loading
│   │   │   ├── env.go                   # Environment parsing
│   │   │   └── constants.go             # Constants
│   │   │
│   │   ├── errors/                      # Error handling
│   │   │   ├── errors.go                # Error definitions
│   │   │   ├── http.go                  # HTTP error mapping
│   │   │   └── codes.go                 # Error codes
│   │   │
│   │   └── utils/                       # Utility functions
│   │       ├── pagination.go
│   │       ├── validation.go
│   │       ├── json.go
│   │       ├── time.go
│   │       ├── uuid.go
│   │       └── helpers.go
│   │
│   ├── migrations/                      # Database migrations
│   │   ├── 001_initial_schema.sql
│   │   ├── 002_add_audit_tables.sql
│   │   ├── 003_add_indexes.sql
│   │   └── ...
│   │
│   ├── proto/                           # Protocol Buffer definitions
│   │   ├── auth/
│   │   │   └── auth.proto
│   │   ├── document/
│   │   │   └── document.proto
│   │   ├── search/
│   │   │   └── search.proto
│   │   ├── workspace/
│   │   │   └── workspace.proto
│   │   └── common/
│   │       └── common.proto
│   │
│   ├── tests/                           # Backend tests
│   │   ├── unit/
│   │   ├── integration/
│   │   └── fixtures/
│   │
│   └── .env.example                     # Environment variables
│
├── ai-services/                         # Python AI services
│   ├── pyproject.toml                   # Poetry config
│   ├── poetry.lock                      # Poetry lock file
│   ├── Dockerfile                       # Docker build
│   ├── requirements.txt                 # Requirements
│   ├── Makefile                         # Build automation
│   │
│   ├── document_processor/              # Document processing service
│   │   ├── __init__.py
│   │   ├── main.py                      # Service entrypoint
│   │   ├── service.py                   # Business logic
│   │   ├── parsers/                     # Document parsers
│   │   │   ├── __init__.py
│   │   │   ├── pdf_parser.py
│   │   │   ├── docx_parser.py
│   │   │   ├── txt_parser.py
│   │   │   └── base_parser.py
│   │   ├── processors/                  # Document processors
│   │   │   ├── __init__.py
│   │   │   ├── chunker.py               # Chunking strategies
│   │   │   ├── cleaner.py               # Text cleaning
│   │   │   ├── metadata_extractor.py    # Metadata extraction
│   │   │   └── preprocessor.py          # Preprocessing
│   │   ├── queue/                       # Queue consumers
│   │   │   ├── __init__.py
│   │   │   └── consumer.py
│   │   └── models/                      # Data models
│   │       ├── __init__.py
│   │       └── types.py
│   │
│   ├── embedding_service/               # Embedding generation service
│   │   ├── __init__.py
│   │   ├── main.py                      # Service entrypoint
│   │   ├── service.py                   # Business logic
│   │   ├── models/                      # Model management
│   │   │   ├── __init__.py
│   │   │   ├── loader.py
│   │   │   ├── openai_model.py
│   │   │   ├── huggingface_model.py
│   │   │   └── base_model.py
│   │   ├── cache/                       # Embedding cache
│   │   │   ├── __init__.py
│   │   │   ├── cache_manager.py
│   │   │   └── redis_cache.py
│   │   ├── queue/                       # Queue consumers
│   │   │   ├── __init__.py
│   │   │   └── consumer.py
│   │   └── models/                      # Data models
│   │       ├── __init__.py
│   │       └── types.py
│   │
│   ├── rag_service/                     # RAG orchestration service
│   │   ├── __init__.py
│   │   ├── main.py                      # Service entrypoint
│   │   ├── service.py                   # Business logic
│   │   ├── pipelines/                   # RAG pipelines
│   │   │   ├── __init__.py
│   │   │   ├── rag_pipeline.py          # Main RAG pipeline
│   │   │   ├── retrieval.py             # Retrieval stage
│   │   │   ├── ranking.py               # Result ranking
│   │   │   ├── context.py               # Context preparation
│   │   │   └── generation.py            # Text generation
│   │   ├── llm/                         # LLM management
│   │   │   ├── __init__.py
│   │   │   ├── llm_client.py
│   │   │   ├── openai_client.py
│   │   │   ├── anthropic_client.py
│   │   │   ├── llama_client.py
│   │   │   ├── prompt_manager.py
│   │   │   └── response_parser.py
│   │   ├── models/                      # Data models
│   │   │   ├── __init__.py
│   │   │   └── types.py
│   │   └── queue/                       # Queue consumers
│   │       ├── __init__.py
│   │       └── consumer.py
│   │
│   ├── agent_service/                   # AI agent service
│   │   ├── __init__.py
│   │   ├── main.py                      # Service entrypoint
│   │   ├── service.py                   # Business logic
│   │   ├── agents/                      # Agent implementations
│   │   │   ├── __init__.py
│   │   │   ├── base_agent.py
│   │   │   ├── research_agent.py
│   │   │   ├── analysis_agent.py
│   │   │   └── planning_agent.py
│   │   ├── tools/                       # Agent tools
│   │   │   ├── __init__.py
│   │   │   ├── search_tool.py
│   │   │   ├── calculator_tool.py
│   │   │   ├── code_tool.py
│   │   │   └── base_tool.py
│   │   ├── state/                       # State management
│   │   │   ├── __init__.py
│   │   │   └── state_manager.py
│   │   ├── models/                      # Data models
│   │   │   ├── __init__.py
│   │   │   └── types.py
│   │   └── queue/                       # Queue consumers
│   │       ├── __init__.py
│   │       └── consumer.py
│   │
│   ├── shared/                          # Shared utilities
│   │   ├── __init__.py
│   │   ├── config.py                    # Configuration
│   │   ├── logger.py                    # Logging setup
│   │   ├── database.py                  # Database connections
│   │   ├── cache.py                     # Cache management
│   │   ├── queue.py                     # Queue management
│   │   ├── vectordb.py                  # VectorDB connections
│   │   ├── http_client.py               # HTTP client
│   │   ├── telemetry.py                 # Telemetry/tracing
│   │   └── utils.py                     # General utilities
│   │
│   ├── tests/                           # AI service tests
│   │   ├── unit/
│   │   ├── integration/
│   │   └── fixtures/
│   │
│   └── .env.example                     # Environment variables
│
├── infrastructure/                      # Infrastructure as Code
│   ├── docker-compose.yml              # Local development
│   ├── docker-compose.prod.yml         # Production compose
│   │
│   ├── kubernetes/                      # Kubernetes manifests
│   │   ├── namespaces/
│   │   ├── config/                      # ConfigMaps and Secrets
│   │   │   ├── app-config.yaml
│   │   │   └── secrets.yaml
│   │   │
│   │   ├── databases/                   # Database deployments
│   │   │   ├── postgres.yaml
│   │   │   ├── redis.yaml
│   │   │   └── milvus.yaml
│   │   │
│   │   ├── services/                    # Service deployments
│   │   │   ├── api-gateway.yaml
│   │   │   ├── auth-service.yaml
│   │   │   ├── document-service.yaml
│   │   │   ├── search-service.yaml
│   │   │   ├── workspace-service.yaml
│   │   │   ├── user-service.yaml
│   │   │   ├── notification-service.yaml
│   │   │   └── audit-service.yaml
│   │   │
│   │   ├── jobs/                        # Batch jobs
│   │   │   ├── document-processor-job.yaml
│   │   │   ├── embedding-job.yaml
│   │   │   ├── cleanup-job.yaml
│   │   │   └── backup-job.yaml
│   │   │
│   │   ├── ingress/
│   │   │   └── ingress.yaml
│   │   │
│   │   ├── monitoring/                  # Observability deployments
│   │   │   ├── prometheus.yaml
│   │   │   ├── grafana.yaml
│   │   │   └── jaeger.yaml
│   │   │
│   │   ├── storage/                     # PersistentVolumes
│   │   │   ├── postgres-pvc.yaml
│   │   │   ├── redis-pvc.yaml
│   │   │   └── milvus-pvc.yaml
│   │   │
│   │   ├── rbac/                        # RBAC configuration
│   │   │   ├── serviceaccounts.yaml
│   │   │   └── roles.yaml
│   │   │
│   │   ├── network-policies/
│   │   │   └── network-policies.yaml
│   │   │
│   │   └── kustomization.yaml           # Kustomize base
│   │
│   ├── terraform/                       # Terraform IaC
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   ├── backend.tf
│   │   │
│   │   ├── modules/
│   │   │   ├── vpc/
│   │   │   ├── kubernetes/
│   │   │   ├── databases/
│   │   │   ├── storage/
│   │   │   ├── monitoring/
│   │   │   └── security/
│   │   │
│   │   ├── environments/
│   │   │   ├── dev/
│   │   │   ├── staging/
│   │   │   └── production/
│   │   │
│   │   └── state/                       # Terraform state (not in git)
│   │
│   ├── helm/                            # Helm charts
│   │   └── enterprise-ai/
│   │       ├── Chart.yaml
│   │       ├── values.yaml
│   │       ├── templates/
│   │       │   ├── deployment.yaml
│   │       │   ├── service.yaml
│   │       │   ├── configmap.yaml
│   │       │   └── ingress.yaml
│   │       └── values/
│   │           ├── values-dev.yaml
│   │           ├── values-staging.yaml
│   │           └── values-prod.yaml
│   │
│   └── scripts/                         # Infrastructure scripts
│       ├── deploy.sh
│       ├── backup.sh
│       ├── restore.sh
│       ├── migrate.sh
│       └── healthcheck.sh
│
├── scripts/                             # Project automation scripts
│   ├── setup.sh                         # Project setup
│   ├── dev.sh                           # Development environment
│   ├── test.sh                          # Run all tests
│   ├── build.sh                         # Build all services
│   ├── docker-build.sh                  # Build docker images
│   ├── lint.sh                          # Code linting
│   ├── format.sh                        # Code formatting
│   └── ci.sh                            # CI pipeline script
│
├── .github/                             # GitHub configuration
│   ├── workflows/                       # CI/CD workflows
│   │   ├── frontend-ci.yml
│   │   ├── backend-ci.yml
│   │   ├── ai-services-ci.yml
│   │   ├── docker-build.yml
│   │   ├── deploy-staging.yml
│   │   └── deploy-production.yml
│   │
│   ├── ISSUE_TEMPLATE/
│   │   ├── bug_report.md
│   │   ├── feature_request.md
│   │   └── discussion.md
│   │
│   └── pull_request_template.md
│
├── docker-compose.yml                  # Local development environment
├── docker-compose.override.yml         # Local overrides
├── .dockerignore                       # Docker ignore rules
├── .gitignore                          # Git ignore rules
├── .editorconfig                       # Editor configuration
├── .pre-commit-config.yaml            # Pre-commit hooks
├── CONTRIBUTING.md                     # Contribution guidelines
├── CODE_OF_CONDUCT.md                 # Code of conduct
├── SECURITY.md                         # Security policy
│
└── .env.example                        # Root environment template
```

## Key Principles

### Frontend Structure
- **App Directory**: Leverages Next.js 14+ App Router for file-based routing
- **Componentization**: Atomic component design for reusability
- **Type Safety**: Full TypeScript coverage with strict mode
- **State Management**: Zustand for application state
- **Data Fetching**: React Query for server state management
- **Testing**: Jest + React Testing Library for unit/integration tests

### Backend Structure
- **Service Isolation**: Each service has independent codebase and deployment
- **Handler-Service-Repository Pattern**: Clear separation of concerns
- **gRPC for Internal**: gRPC for service-to-service communication
- **REST for External**: REST APIs for client and third-party integration
- **Database Migrations**: Version-controlled schema migrations
- **Proto-first**: Protocol Buffers define service contracts

### AI Services Structure
- **Modular Pipelines**: Independent, reusable pipeline components
- **Service-oriented**: Each AI service runs independently
- **Async-first**: Queue-based processing for long-running tasks
- **Model Abstraction**: Pluggable LLM and embedding models
- **Testing**: Comprehensive unit and integration tests

### Infrastructure Structure
- **IaC-first**: All infrastructure in Terraform and Kubernetes manifests
- **Multi-environment**: Dev, staging, and production configurations
- **Helm Charts**: Reusable Kubernetes deployments
- **Scripts**: Automation for common operations
- **CI/CD**: GitHub Actions workflows for automated deployment

## File Naming Conventions

| File Type | Convention | Example |
|-----------|-----------|---------|
| React Components | PascalCase | `UserProfile.tsx` |
| Utilities | camelCase | `formatDate.ts` |
| Types | PascalCase | `User.ts` |
| Services | camelCase | `userService.ts` |
| Tests | `.test.` or `.spec.` | `Button.test.tsx` |
| Go packages | lowercase | `auth`, `document` |
| Go files | lowercase | `handler.go` |
| Database migrations | numbered | `001_initial_schema.sql` |
| Config files | snake_case | `docker-compose.yml` |

## Module Organization Rules

1. **Single Responsibility**: Each module handles one domain
2. **Minimal Exports**: Export only what's needed from modules
3. **Internal Packages**: Use `/internal` for non-exported code
4. **Shared Utilities**: Place shared code in `shared/` or `common/`
5. **Clear Dependencies**: Avoid circular dependencies
6. **Type Definitions**: Colocate types with domain logic

## Future Structure Considerations

- Plugin system for extending AI capabilities
- Multi-workspace data isolation at directory level
- Event sourcing for audit trail
- Command Query Responsibility Segregation (CQRS)
- API versioning strategies as needed