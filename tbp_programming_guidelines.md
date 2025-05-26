# TBP Programming Guidelines

## 1. Introduction

These programming guidelines define the standards and best practices for developing the Trusted Business Platform (TBP). They are based on official Go conventions but extend them with project-specific requirements. All developers in the TBP project commit to following these guidelines to ensure a consistent, maintainable, and high-quality codebase.

The guidelines should be understood as a living document. Improvement suggestions are welcome and should be discussed within the team. Deviations from the guidelines are possible in justified exceptional cases but must be documented.

## 2. Project Structure and Organization

### 2.1 Repository Layout

Each TBP repository follows the standard Go project layout with project-specific extensions:

```
repository-name/
├── cmd/                    # Executable programs
│   └── service-name/       # Main package for each service
├── internal/               # Private packages (not usable by other repos)
│   ├── domain/             # Domain models and business logic
│   ├── service/            # Service layer implementations
│   ├── repository/         # Data access layer
│   └── handler/            # gRPC/HTTP handlers
├── pkg/                    # Public packages (usable by other repos)
├── api/                    # API definitions (Proto files, OpenAPI)
│   └── v1/                 # Versioned API definitions
├── configs/                # Configuration files
├── scripts/                # Build and deployment scripts
├── docs/                   # Documentation
│   ├── architecture/       # Architecture decisions (ADRs)
│   └── api/                # API documentation
├── test/                   # Integration tests and test data
├── .github/                # GitHub-specific configuration
├── Makefile                # Build automation
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── doc.go                  # Package documentation
└── README.md               # Project overview
```

### 2.2 Package Naming

Packages are written in lowercase, are short and meaningful. They do not use underscores or camelCase. The package name should clearly describe the content:

```go
// Good
package auth
package timeentry
package tcol

// Bad
package authenticationManager
package time_entry
package TCOL
```

### 2.3 Import Organization

Imports are organized into three groups, separated by blank lines:

1. Standard Library
2. External Dependencies
3. Internal Packages

```go
import (
    "context"
    "fmt"
    "time"

    "google.golang.org/grpc"
    "github.com/shopspring/decimal"

    "github.com/msto63/tbp/tbp-foundation/pkg/errors"
    "github.com/msto63/tbp/tbp-foundation/pkg/logging"
)
```

### 2.4 Language Convention

All code, including comments, documentation, variable names, and commit messages must be written in English. This ensures the codebase is accessible to international contributors and maintains consistency.

```go
// Good - English throughout
// CustomerRepository provides persistence operations for customer entities.
// It implements the Repository pattern and ensures data consistency.
type CustomerRepository struct {
    db *sql.DB
    // cache holds recently accessed customers for performance optimization
    cache map[string]*Customer
}

// Bad - Mixed languages
// CustomerRepository bietet Persistenz-Operationen für Kunden
type CustomerRepository struct {
    db *sql.DB
    // Zwischenspeicher für Kunden
    cache map[string]*Customer
}
```

Even domain-specific terms should be translated to English:

```go
// Good
type Invoice struct {
    ID          string
    CustomerID  string
    DueDate     time.Time
    TotalAmount decimal.Decimal
}

// Bad - German terms
type Rechnung struct {
    ID          string
    KundenID    string
    Faelligkeit time.Time
    Gesamtbetrag decimal.Decimal
}
```

## 3. Naming Conventions

### 3.1 General Rules

Go's naming conventions are strictly followed:

- Exported names begin with a capital letter
- Non-exported names begin with a lowercase letter
- Acronyms are uniformly capitalized (URL, HTTP, ID)
- Names are self-explanatory and avoid abbreviations

### 3.2 Interface Names

Interfaces use verb suffixes for single-method interfaces and nouns for larger interfaces:

```go
// Single-method interfaces
type Reader interface {
    Read([]byte) (int, error)
}

type Validator interface {
    Validate() error
}

// Larger interfaces
type CustomerRepository interface {
    Find(ctx context.Context, id string) (*Customer, error)
    Save(ctx context.Context, customer *Customer) error
    Delete(ctx context.Context, id string) error
}
```

### 3.3 Variable and Function Names

Variable names are short but meaningful. Single-letter names are acceptable in short scopes:

```go
// Good - short scope
for i, v := range values {
    process(v)
}

// Good - descriptive in larger scope
customerRepository := repository.NewCustomerRepository(db)

// Bad - too short for the scope
cr := repository.NewCustomerRepository(db)
```

## 4. Documentation

### 4.1 Package Documentation

Each package has a `doc.go` file with comprehensive overview and standardized header:

```go
// Package tcol implements the Terminal Command Object Language parser
// and interpreter for the tbp platform.
//
// TCOL is a domain-specific language designed for efficient interaction
// with business objects through terminal commands. It supports object-oriented
// syntax with method calls, filtering, and command chaining.
//
// Basic usage:
//
//     parser := tcol.NewParser()
//     cmd, err := parser.Parse("CUSTOMER.LIST")
//     if err != nil {
//         // Handle error
//     }
//     result, err := cmd.Execute(ctx)
//
// For more information, see the TCOL specification in the docs directory.
//
// Package: tcol
// Title: Terminal Command Object Language Parser and Interpreter
// Description: This package provides comprehensive parsing and execution
//              capabilities for TCOL commands, including syntax validation,
//              command optimization, and secure execution contexts.
// Author: msto63 with Claude Sonnet 4.0
// Version: v1.0.0
// Created: 2024-01-15
// Modified: 2024-01-15
//
// Change History:
// - 2024-01-15 v1.0.0: Initial implementation with basic parsing
package tcol
```

### 4.2 File-Level Documentation

For individual files within a package, we continue to use structured header comments:

```go
// File: parser.go
// Title: TCOL Command Parser Implementation
// Description: Implements the core parsing logic for TCOL commands,
//              including tokenization, syntax analysis, and AST generation.
//              Handles command abbreviations and provides detailed error messages.
// Author: msto63 with Claude Sonnet 4.0
// Version: v1.0.0
// Created: 2024-01-15
// Modified: 2024-01-15
//
// Change History:
// - 2024-01-15 v1.0.0: Initial parser implementation

package tcol

import (
    "context"
    "fmt"
    "strings"
)

// Parser implements the TCOL command parsing functionality.
// It maintains state for command abbreviations and provides
// detailed error reporting for syntax issues.
type Parser struct {
    // commands holds the registry of available commands
    commands map[string]Command
    // abbreviations caches computed abbreviations for performance
    abbreviations map[string]string
}
```

### 4.3 Code Documentation

Every exported function, type, and constant is documented:

```go
// Customer represents a business customer entity.
// It contains all relevant information for customer management
// and is the central domain object for customer-related operations.
type Customer struct {
    // ID is the unique identifier for the customer.
    // It is immutable once created.
    ID string

    // Name is the customer's display name.
    // It must not be empty and is limited to 255 characters.
    Name string

    // CreatedAt indicates when the customer was first created.
    // It is set automatically and cannot be modified.
    CreatedAt time.Time
}

// CreateCustomer creates a new customer with the given name.
// It returns an error if the name is empty or exceeds 255 characters.
// The created customer will have a generated ID and the current timestamp.
func CreateCustomer(name string) (*Customer, error) {
    // Implementation...
}
```

## 5. Error Handling

### 5.1 Error Wrapping

Errors are always provided with context before being returned:

```go
// Use tbp-foundation's error package
import "github.com/msto63/tbp/tbp-foundation/pkg/errors"

func (s *Service) ProcessInvoice(ctx context.Context, id string) error {
    invoice, err := s.repo.FindInvoice(ctx, id)
    if err != nil {
        return errors.Wrap(err, "failed to find invoice",
            errors.WithCode(errors.CodeNotFound),
            errors.WithDetail("invoice_id", id),
        )
    }

    // Business logic...

    return nil
}
```

### 5.2 Error Handling Patterns

Errors are handled immediately, not later:

```go
// Good
result, err := doSomething()
if err != nil {
    return nil, errors.Wrap(err, "failed to do something")
}

// Bad
result, err := doSomething()
// ... other operations ...
if err != nil {
    return nil, err
}
```

### 5.3 Panic Usage

Panics are only used in truly unexpected situations that represent a programming error:

```go
// Acceptable - programming error
func NewService(repo Repository) *Service {
    if repo == nil {
        panic("repository is required")
    }
    return &Service{repo: repo}
}

// Not acceptable - normal error case
func (s *Service) Process(id string) error {
    if id == "" {
        panic("id is empty") // WRONG - return error
    }
}
```

## 6. Concurrency

### 6.1 Goroutine Management

Goroutines are always created with clear lifecycle management:

```go
// Good - with Context for cancellation
func (s *Service) StartWorker(ctx context.Context) error {
    g, ctx := errgroup.WithContext(ctx)

    g.Go(func() error {
        return s.processLoop(ctx)
    })

    return g.Wait()
}

// Bad - uncontrolled goroutine
func (s *Service) StartWorker() {
    go s.processLoop() // No way to stop
}
```

### 6.2 Channel Usage

Channels are used with clear ownership. The sender closes the channel:

```go
func (s *Service) Produce(ctx context.Context) <-chan Item {
    ch := make(chan Item)

    go func() {
        defer close(ch) // Sender closes

        for {
            select {
            case <-ctx.Done():
                return
            case ch <- s.nextItem():
                // Continue producing
            }
        }
    }()

    return ch
}
```

### 6.3 Mutex Conventions

Mutexes are placed directly above the fields they protect:

```go
type Service struct {
    mu        sync.RWMutex
    customers map[string]*Customer // protected by mu

    // other fields...
}

func (s *Service) GetCustomer(id string) *Customer {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.customers[id]
}
```

## 7. Testing

### 7.1 Test Organization

Tests follow the structure of the tested code:

```go
// For service.go
func TestService_ProcessInvoice(t *testing.T) {
    t.Run("success case", func(t *testing.T) {
        // Test setup
        repo := &mockRepository{
            invoice: &Invoice{ID: "123"},
        }
        svc := NewService(repo)

        // Execution
        err := svc.ProcessInvoice(context.Background(), "123")

        // Assertions
        assert.NoError(t, err)
        assert.Equal(t, 1, repo.findCalled)
    })

    t.Run("invoice not found", func(t *testing.T) {
        // ...
    })
}
```

### 7.2 Table-Driven Tests

Table-driven tests are used for tests with multiple similar cases:

```go
func TestValidateCommand(t *testing.T) {
    tests := []struct {
        name    string
        command string
        want    bool
        wantErr bool
    }{
        {
            name:    "valid customer command",
            command: "CUSTOMER.CREATE",
            want:    true,
            wantErr: false,
        },
        {
            name:    "invalid syntax",
            command: "INVALID",
            want:    false,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ValidateCommand(tt.command)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
        })
    }
}
```

### 7.3 Mocking

Interfaces are used for mocking, not concrete types:

```go
// Interface for mockability
type TimeProvider interface {
    Now() time.Time
}

// Concrete implementation
type systemTimeProvider struct{}

func (p *systemTimeProvider) Now() time.Time {
    return time.Now()
}

// Mock for tests
type mockTimeProvider struct {
    now time.Time
}

func (p *mockTimeProvider) Now() time.Time {
    return p.now
}
```

## 8. Performance

### 8.1 Benchmarking

Performance-critical code is equipped with benchmarks:

```go
func BenchmarkTCOLParser_Parse(b *testing.B) {
    parser := NewTCOLParser()
    command := "CUSTOMER[status='active'].LIST"

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := parser.Parse(command)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### 8.2 Allocation Awareness

Code is written with allocations in mind:

```go
// Good - reusable buffer
type Parser struct {
    buf []byte // Reusable
}

func (p *Parser) Parse(input string) (*Command, error) {
    p.buf = p.buf[:0] // Reset but keep capacity
    // Parse logic...
}

// Bad - new allocation on each call
func Parse(input string) (*Command, error) {
    buf := make([]byte, len(input)) // New allocation
    // Parse logic...
}
```

## 9. Security

### 9.1 Input Validation

All external inputs are validated:

```go
func (s *Service) CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*Customer, error) {
    // Validation
    if err := validateCustomerName(req.Name); err != nil {
        return nil, errors.Wrap(err, "invalid customer name",
            errors.WithCode(errors.CodeValidationFailed),
        )
    }

    if err := validateEmail(req.Email); err != nil {
        return nil, errors.Wrap(err, "invalid email",
            errors.WithCode(errors.CodeValidationFailed),
        )
    }

    // Only process after validation
    return s.repo.CreateCustomer(ctx, req)
}
```

### 9.2 No Sensitive Data in Logs

Sensitive data is never logged:

```go
// Good
log.Info("user login attempt", 
    logging.Field("user_id", userID),
    logging.Field("ip", request.RemoteAddr),
)

// Bad
log.Info("user login attempt",
    logging.Field("password", password), // NEVER!
)
```

## 10. Dependency Management

### 10.1 Low-Dependency Philosophy

The TBP project follows a "Low-Dependency Philosophy" that consciously minimizes the number of external dependencies while being pragmatic enough to use proven and stable libraries. New dependencies are only added after careful evaluation:

**Evaluation Criteria for New Dependencies:**

- Is the functionality really necessary?
- Can it be reasonably implemented in-house?
- Is the library actively maintained and has a stable API?
- Does it itself have minimal dependencies?
- Is the license compatible with the project?
- Does it offer significant value over a custom implementation?

**Acceptable Dependencies:**

- Proven standard extensions (e.g., `golang.org/x/...`)
- Testing utilities (e.g., `testify` for more comprehensive assertions)
- Established protocols (e.g., `google.golang.org/grpc`)
- Security-critical libraries (e.g., crypto libraries from trusted sources)
- Domain-specific formats (e.g., `shopspring/decimal` for financial calculations)

**Dependencies to Avoid:**

- Libraries for trivial functions
- Experimental or unstable packages
- Dependencies with extensive dependency trees
- Libraries that want to "improve" core Go functionality

### 10.2 Version Pinning

Dependencies are always pinned to specific versions:

```go
// go.mod
require (
    google.golang.org/grpc v1.58.0
    github.com/shopspring/decimal v1.3.1
    github.com/stretchr/testify v1.8.4 // Acceptable for testing
)

// No unspecified versions like "latest"
```

### 10.3 Dependency Review Process

Each new dependency goes through a review process:

1. **Technical Assessment**: Performance, API design, maintenance quality
2. **Security Review**: CVE history, update cycles, maintainer reputation
3. **License Compliance**: Compatibility with project license
4. **Alternatives Analysis**: Comparison with custom implementation and other libraries
5. **Team Approval**: Discussion and consensus within the development team

## 11. Workspace Management

### 11.1 go.work Usage

The TBP project uses `go.work` for development across multiple repositories:

```bash
tbp/                         # Workspace root
├── go.work                  # Workspace definition
├── go.work.sum              # Workspace checksums
├── tbp-foundation/          # Foundation repository
├── tbp-server/              # Application server
├── tbp-tui-client/          # TUI client
├── services/                # Service repositories
│   ├── task-service/
│   ├── customer-service/
│   └── timeentry-service/
└── tools/                   # Development tools
    └── tbp-cli/
```

The `go.work` file:

```go
go 1.21

use (
    ./tbp-foundation
    ./tbp-server
    ./tbp-tui-client
    ./services/task-service
    ./services/customer-service
    ./services/timeentry-service
    ./tools/tbp-cli
)
```

### 11.2 Development Workflow

```bash
# Initial setup
mkdir tbp && cd tbp
git clone https://github.com/msto63/tbp/tbp-foundation.git
# ... clone other repos

# Initialize workspace
go work init
go work use ./tbp-foundation ./tbp-server

# Daily development - changes in foundation are immediately available
cd tbp-foundation
# make changes...
cd ../tbp-server
go test ./... # Uses local foundation

# Production builds ignore workspace
GOWORK=off go build ./cmd/server
```

**Important**: `go.work` and `go.work.sum` are not committed. Use `go.work.example` as template.

## 12. Code Review Checklist

Before each pull request, this checklist is reviewed:

- [ ] Code follows TBP programming guidelines
- [ ] All tests pass
- [ ] New features have tests
- [ ] Documentation is updated (including doc.go and file headers)
- [ ] No TODO comments without issue reference
- [ ] Performance-critical code has benchmarks
- [ ] Error handling is complete
- [ ] No goroutine leaks
- [ ] Sensitive data is not logged
- [ ] Dependencies are minimal and justified
- [ ] Code and comments are in English
- [ ] Package documentation in doc.go is complete
- [ ] File-level documentation headers are present

## 13. Continuous Integration & Continuous Deployment

### 13.1 CI Pipeline Requirements

Every pull request must pass all CI checks before merging. The pipeline is the quality gate that ensures code consistency and prevents regressions.

**Mandatory CI Steps**:

```yaml
# Example GitHub Actions workflow
name: CI
on:
  pull_request:
    branches: [main, develop]
  push:
    branches: [main, develop]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --config=.golangci.yml

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
      - name: Run tests
        run: |
          go test -v -race -coverprofile=coverage.out ./...
          go tool cover -html=coverage.out -o coverage.html
      - name: Check coverage
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Coverage: $COVERAGE%"
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "Coverage is below 80%"
            exit 1
          fi

  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
      - name: Build
        run: |
          go build -v ./...
          # For services with cmd/
          for cmd in cmd/*; do
            if [ -d "$cmd" ]; then
              go build -v "./$cmd"
            fi
          done

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run gosec
        uses: securecode/gosec@master
        with:
          args: ./...
      - name: Run dependency check
        run: |
          go list -json -m all | nancy sleuth
```

### 13.2 Branch Protection Rules

The `main` branch is protected with the following rules:

- Direct pushes are forbidden
- Pull requests require at least one approval
- All CI checks must pass
- Branches must be up to date with main before merging
- Commits must be signed

### 13.3 Commit Message Convention

Commits follow the Conventional Commits specification for automated changelog generation:

```bash
# Format
<type>(<scope>): <subject>

<body>

<footer>

# Examples
feat(tcol): add support for abbreviated commands
fix(auth): resolve token expiration issue
docs(readme): update installation instructions
test(customer): add integration tests for service layer
refactor(repository): simplify query builder logic
perf(parser): optimize command parsing for large inputs
build(deps): update grpc to v1.58.0
```

**Types**:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `test`: Adding or updating tests
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `perf`: Performance improvement
- `build`: Changes to build system or dependencies
- `ci`: Changes to CI configuration
- `chore`: Other changes that don't modify src or test files

### 13.4 Version Tagging

Releases follow semantic versioning with automated tagging:

```bash
# Version format: v<major>.<minor>.<patch>
v1.0.0    # Initial stable release
v1.1.0    # New feature (backwards compatible)
v1.1.1    # Bug fix
v2.0.0    # Breaking change

# Pre-releases
v1.0.0-alpha.1
v1.0.0-beta.1
v1.0.0-rc.1
```

### 13.5 CD Pipeline

Continuous Deployment is triggered by tags on the main branch:

```yaml
# .github/workflows/release.yml
name: Release
on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4

      - name: Run tests
        run: go test -v ./...

      - name: Build binaries
        run: |
          # Build for multiple platforms
          GOOS=linux GOARCH=amd64 go build -o dist/tbp-linux-amd64 ./cmd/tbp
          GOOS=darwin GOARCH=amd64 go build -o dist/tbp-darwin-amd64 ./cmd/tbp
          GOOS=windows GOARCH=amd64 go build -o dist/tbp-windows-amd64.exe ./cmd/tbp

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: dist/*
          generate_release_notes: true
```

### 13.6 Quality Metrics

Each repository maintains quality metrics that are tracked over time:

**Required Metrics**:

- Code coverage: Minimum 80%, target 90%
- Cyclomatic complexity: Maximum 10 per function
- Duplication: Maximum 3%
- Technical debt ratio: Maximum 5%

**Monitoring Dashboard**:

```go
// Each service exposes metrics endpoint
func (s *Server) registerMetrics() {
    // Custom business metrics
    requestDuration := prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "tbp_request_duration_seconds",
            Help: "Request duration in seconds",
        },
        []string{"service", "method", "status"},
    )

    prometheus.MustRegister(requestDuration)
}
```

### 13.7 Database Migrations

Database changes follow a strict migration process:

```bash
# Migration files follow timestamp naming
migrations/
├── 20240115143022_create_customers_table.up.sql
├── 20240115143022_create_customers_table.down.sql
├── 20240120091511_add_customer_status.up.sql
└── 20240120091511_add_customer_status.down.sql
```

**Migration Rules**:

- Every `up` migration must have a corresponding `down` migration
- Migrations must be backwards compatible for zero-downtime deployments
- Large data migrations run in batches
- Schema changes are tested in staging environment first

### 13.8 Dependency Updates

Dependencies are updated regularly with automated PRs:

```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 10
    reviewers:
      - "tbp-maintainers"
```

**Update Process**:

1. Dependabot creates PR with update
2. CI runs full test suite
3. Manual review for breaking changes
4. Merge if all checks pass

### 13.9 Performance Benchmarks

Performance-critical code includes benchmarks that run in CI:

```bash
# Benchmark results are tracked over time
go test -bench=. -benchmem -benchtime=10s | tee benchmark.txt

# Compare with baseline
benchstat baseline.txt benchmark.txt

# Fail if regression detected (>10% slower)
```

### 13.10 Release Checklist

Before each release:

- [ ] All tests pass
- [ ] Coverage meets minimum requirements
- [ ] No security vulnerabilities (gosec, dependency scan)
- [ ] Documentation is updated
- [ ] CHANGELOG.md is updated
- [ ] Migration scripts tested
- [ ] Performance benchmarks show no regression
- [ ] API changes are backwards compatible (or major version bump)
- [ ] Release notes drafted
- [ ] Stakeholders notified

## 14. Continuous Improvement

These guidelines evolve with the project. Improvement suggestions are welcome and should be created as an issue with the "guidelines" label. Changes are discussed in the team and adopted upon consensus.

Compliance with the guidelines is supported by automated tools:

- `golangci-lint` with project-specific configuration
- `go fmt` and `goimports` for formatting
- Custom linter for TBP-specific rules

Every developer is responsible for the quality of their own code and constructive review of others' code. Together, we create a codebase that meets the high standards of the Trusted Business Platform.
