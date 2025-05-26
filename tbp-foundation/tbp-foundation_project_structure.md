# TBP Foundation - Complete Project Structure

```
tbp-foundation/
├── cmd/                                   # CLI tools and utilities
│   ├── tbp-codegen/                       # Code generator tool
│   │   ├── main.go
│   │   └── templates/
│   │       ├── service.go.tmpl
│   │       ├── repository.go.tmpl
│   │       └── handler.go.tmpl
│   ├── tbp-migrate/                       # Migration helper tool
│   │   └── main.go
│   └── tbp-diagnostics/                   # Service diagnostics tool
│       └── main.go
│
├── pkg/                                   # Public packages (importable by other modules)
│   ├── core/                              # Essential core functionality
│   │   ├── doc.go                         # Package documentation
│   │   ├── context.go                     # Extended context management
│   │   ├── context_test.go
│   │   ├── errors.go                      # Basic error types and handling
│   │   ├── errors_test.go
│   │   ├── types.go                       # Common types and interfaces
│   │   └── version.go                     # Version information
│   │
│   ├── config/                            # Configuration management
│   │   ├── doc.go
│   │   ├── config.go                      # Configuration loading and parsing
│   │   ├── config_test.go
│   │   ├── env.go                         # Environment variable handling
│   │   ├── env_test.go
│   │   ├── file.go                        # File-based configuration
│   │   ├── file_test.go
│   │   ├── validator.go                   # Configuration validation
│   │   └── validator_test.go
│   │
│   ├── logging/                           # Structured logging
│   │   ├── doc.go
│   │   ├── logger.go                      # Main logger interface and implementation
│   │   ├── logger_test.go
│   │   ├── context.go                     # Context-aware logging
│   │   ├── context_test.go
│   │   ├── fields.go                      # Structured field handling
│   │   ├── fields_test.go
│   │   ├── formatters.go                  # Output formatters (JSON, text)
│   │   ├── formatters_test.go
│   │   ├── levels.go                      # Log level management
│   │   └── benchmark_test.go              # Performance benchmarks
│   │
│   ├── errors/                            # Advanced error handling
│   │   ├── doc.go
│   │   ├── errors.go                      # Error types and wrapping
│   │   ├── errors_test.go
│   │   ├── codes.go                       # Error code definitions
│   │   ├── codes_test.go
│   │   ├── details.go                     # Error detail management
│   │   ├── details_test.go
│   │   ├── stack.go                       # Stack trace handling
│   │   ├── stack_test.go
│   │   └── classification.go              # Error classification (retryable, etc.)
│   │
│   ├── security/                          # Security components
│   │   ├── doc.go
│   │   ├── auth/                          # Authentication
│   │   │   ├── doc.go
│   │   │   ├── jwt.go                     # JWT token handling
│   │   │   ├── jwt_test.go
│   │   │   ├── oauth2.go                  # OAuth2 integration
│   │   │   ├── oauth2_test.go
│   │   │   ├── mtls.go                    # Mutual TLS authentication
│   │   │   └── mtls_test.go
│   │   ├── authz/                         # Authorization
│   │   │   ├── doc.go
│   │   │   ├── rbac.go                    # Role-based access control
│   │   │   ├── rbac_test.go
│   │   │   ├── abac.go                    # Attribute-based access control
│   │   │   ├── abac_test.go
│   │   │   ├── permissions.go             # Permission management
│   │   │   └── permissions_test.go
│   │   ├── crypto/                        # Cryptographic operations
│   │   │   ├── doc.go
│   │   │   ├── encryption.go              # Encryption/decryption
│   │   │   ├── encryption_test.go
│   │   │   ├── signing.go                 # Digital signatures
│   │   │   ├── signing_test.go
│   │   │   ├── keys.go                    # Key management
│   │   │   ├── keys_test.go
│   │   │   └── random.go                  # Secure random generation
│   │   └── audit/                         # Audit trail
│   │       ├── doc.go
│   │       ├── audit.go                   # Audit logging
│   │       ├── audit_test.go
│   │       ├── trail.go                   # Audit trail management
│   │       ├── trail_test.go
│   │       └── compliance.go              # Compliance helpers
│   │
│   ├── patterns/                          # Design pattern implementations
│   │   ├── doc.go
│   │   ├── command/                       # Command pattern
│   │   │   ├── doc.go
│   │   │   ├── command.go                 # Command interface and base
│   │   │   ├── command_test.go
│   │   │   ├── invoker.go                 # Command invoker
│   │   │   ├── invoker_test.go
│   │   │   ├── queue.go                   # Command queue
│   │   │   ├── queue_test.go
│   │   │   └── history.go                 # Command history (undo/redo)
│   │   ├── repository/                    # Repository pattern
│   │   │   ├── doc.go
│   │   │   ├── repository.go              # Generic repository interfaces
│   │   │   ├── repository_test.go
│   │   │   ├── memory.go                  # In-memory implementation
│   │   │   ├── memory_test.go
│   │   │   ├── sql.go                     # SQL database implementation
│   │   │   ├── sql_test.go
│   │   │   └── query.go                   # Query builder
│   │   └── events/                        # Event-driven patterns
│   │       ├── doc.go
│   │       ├── events.go                  # Event types and interfaces
│   │       ├── events_test.go
│   │       ├── bus.go                     # Event bus
│   │       ├── bus_test.go
│   │       ├── sourcing.go                # Event sourcing
│   │       ├── sourcing_test.go
│   │       └── projection.go              # Event projection
│   │
│   ├── utils/                             # Utility modules
│   │   ├── stringx/                       # Extended string operations
│   │   │   ├── doc.go
│   │   │   ├── abbrev.go                  # Abbreviation handling
│   │   │   ├── abbrev_test.go
│   │   │   ├── similarity.go              # String similarity
│   │   │   ├── similarity_test.go
│   │   │   ├── truncate.go                # Unicode-aware truncation
│   │   │   ├── truncate_test.go
│   │   │   └── benchmark_test.go
│   │   ├── mathx/                         # Business mathematics
│   │   │   ├── doc.go
│   │   │   ├── decimal.go                 # Decimal arithmetic
│   │   │   ├── decimal_test.go
│   │   │   ├── currency.go                # Currency operations
│   │   │   ├── currency_test.go
│   │   │   ├── interest.go                # Interest calculations
│   │   │   ├── interest_test.go
│   │   │   └── rounding.go                # Rounding methods
│   │   └── collections/                   # Type-safe collections
│   │       ├── doc.go
│   │       ├── set.go                     # Generic set implementation
│   │       ├── set_test.go
│   │       ├── orderedmap.go              # Ordered map
│   │       ├── orderedmap_test.go
│   │       ├── queue.go                   # Priority queue
│   │       ├── queue_test.go
│   │       ├── functional.go              # Map, Filter, Reduce operations
│   │       └── functional_test.go
│   │
│   ├── testing/                           # Testing utilities
│   │   ├── doc.go
│   │   ├── mock.go                        # Mocking utilities
│   │   ├── mock_test.go
│   │   ├── fixtures.go                    # Test data fixtures
│   │   ├── fixtures_test.go
│   │   ├── time.go                        # Time mocking
│   │   ├── time_test.go
│   │   ├── assert.go                      # Additional assertions
│   │   └── benchmark.go                   # Benchmark helpers
│   │
│   ├── metrics/                           # Performance monitoring
│   │   ├── doc.go
│   │   ├── metrics.go                     # Metrics collection
│   │   ├── metrics_test.go
│   │   ├── prometheus.go                  # Prometheus integration
│   │   ├── prometheus_test.go
│   │   ├── histogram.go                   # Latency histograms
│   │   └── regression.go                  # Performance regression detection
│   │
│   └── workflow/                          # Event-driven architecture
│       ├── doc.go
│       ├── engine.go                      # Workflow engine
│       ├── engine_test.go
│       ├── n8n.go                         # n8n integration
│       ├── n8n_test.go
│       ├── triggers.go                    # Event triggers
│       └── triggers_test.go
│
├── internal/                              # Private packages (not importable)
│   ├── shared/                            # Shared internal utilities
│   │   ├── constants.go                   # Internal constants
│   │   ├── helpers.go                     # Internal helper functions
│   │   └── testdata/                      # Test data files
│   │       ├── config.yaml
│   │       ├── test-cert.pem
│   │       └── test-key.pem
│   └── build/                             # Build-time utilities
│       ├── version.go                     # Version generation
│       └── embed.go                       # Embedded resources
│
├── api/                                   # API definitions
│   └── v1/                                # Version 1 APIs
│       ├── foundation.proto               # gRPC service definitions
│       ├── errors.proto                   # Error type definitions
│       └── events.proto                   # Event type definitions
│
├── configs/                               # Configuration files
│   ├── default.yaml                       # Default configuration
│   ├── development.yaml                   # Development settings
│   ├── production.yaml                    # Production settings
│   └── golangci.yml                       # Linter configuration
│
├── scripts/                               # Build and deployment scripts
│   ├── build.sh                           # Build script
│   ├── test.sh                            # Test script
│   ├── lint.sh                            # Linting script
│   ├── benchmark.sh                       # Benchmark script
│   └── generate.sh                        # Code generation script
│
├── docs/                                  # Documentation
│   ├── architecture/                      # Architecture decisions
│   │   ├── ADR-001-modular-design.md
│   │   ├── ADR-002-dependency-philosophy.md
│   │   ├── ADR-003-error-handling.md
│   │   └── ADR-004-security-model.md
│   ├── api/                               # API documentation
│   │   ├── core.md
│   │   ├── logging.md
│   │   ├── security.md
│   │   └── patterns.md
│   ├── examples/                          # Usage examples
│   │   ├── basic-service/
│   │   │   ├── main.go
│   │   │   └── README.md
│   │   ├── enterprise-service/
│   │   │   ├── main.go
│   │   │   └── README.md
│   │   └── migration-guide/
│   │       └── README.md
│   ├── performance/                       # Performance documentation
│   │   ├── benchmarks.md
│   │   └── optimization-guide.md
│   └── security/                          # Security documentation
│       ├── threat-model.md
│       └── compliance.md
│
├── test/                                  # Integration tests
│   ├── integration/                       # Integration test suites
│   │   ├── auth_test.go
│   │   ├── logging_test.go
│   │   └── end_to_end_test.go
│   ├── fixtures/                          # Test fixtures
│   │   ├── certificates/
│   │   ├── configs/
│   │   └── data/
│   └── performance/                       # Performance tests
│       ├── load_test.go
│       └── stress_test.go
│
├── tools/                                 # Development tools
│   ├── tools.go                           # Tool dependencies
│   └── install.sh                         # Tool installation script
│
├── .github/                               # GitHub specific files
│   ├── workflows/                         # GitHub Actions
│   │   ├── ci.yml                         # Continuous Integration
│   │   ├── release.yml                    # Release automation
│   │   └── security.yml                   # Security scanning
│   ├── ISSUE_TEMPLATE/                    # Issue templates
│   │   ├── bug_report.md
│   │   ├── feature_request.md
│   │   └── performance_issue.md
│   └── PULL_REQUEST_TEMPLATE.md           # PR template
│
├── deployments/                           # Deployment configurations
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   └── kubernetes/
│       ├── namespace.yaml
│       └── configmap.yaml
│
├── examples/                              # Example applications
│   ├── minimal/                           # Minimal foundation usage
│   │   ├── main.go
│   │   └── README.md
│   ├── standard/                          # Standard service example
│   │   ├── main.go
│   │   ├── config.yaml
│   │   └── README.md
│   └── enterprise/                        # Full enterprise example
│       ├── main.go
│       ├── config.yaml
│       ├── docker-compose.yml
│       └── README.md
│
├── migrations/                            # Database migration examples
│   ├── 001_initial_schema.up.sql
│   ├── 001_initial_schema.down.sql
│   └── README.md
│
├── third_party/                           # Third party integrations
│   ├── prometheus/
│   │   └── grafana-dashboard.json
│   └── n8n/
│       └── workflow-templates/
│
├── .gitignore                             # Git ignore file
├── .golangci.yml                          # Linter configuration
├── go.mod                                 # Go module definition
├── go.sum                                 # Dependency checksums
├── go.work.example                        # Workspace example
├── doc.go                                 # Root package documentation
├── Makefile                               # Build automation
├── README.md                              # Project overview
├── CHANGELOG.md                           # Change log
├── LICENSE                                # License file
├── CONTRIBUTING.md                        # Contribution guidelines
├── SECURITY.md                            # Security policy
└── CODE_OF_CONDUCT.md                     # Code of conduct
```

## Key Design Principles

### 1. **Modular Architecture**

- Each `pkg/` subdirectory is an independent module
- Clear separation between public (`pkg/`) and private (`internal/`) packages
- Gradual adoption possible through selective imports

### 2. **Comprehensive Testing**

- Unit tests alongside each module (`*_test.go`)
- Integration tests in dedicated `test/` directory
- Performance benchmarks (`benchmark_test.go`)
- Test utilities and fixtures

### 3. **Documentation First**

- Package documentation in `doc.go` files
- Architecture Decision Records (ADRs)
- API documentation and examples
- Migration guides and tutorials

### 4. **Enterprise Ready**

- Security components with auth, crypto, and audit
- Performance monitoring and metrics
- Configuration management for different environments
- CI/CD pipeline with GitHub Actions

### 5. **Developer Experience**

- CLI tools for code generation and diagnostics
- Example applications for different use cases
- Comprehensive build scripts and automation
- Clear contribution guidelines

### 6. **Production Deployment**

- Docker and Kubernetes configurations
- Multiple environment configs
- Database migration examples
- Third-party integration templates

This structure provides a solid foundation for the TBP ecosystem while maintaining modularity, testability, and enterprise-grade quality standards.
