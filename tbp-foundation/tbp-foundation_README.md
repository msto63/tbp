# tbp-foundation - README

## 1. Executive Summary

The tbp-foundation forms the technological foundation of the Trusted Business Platform and provides essential base functionalities for all components built upon it. As a shared codebase, it defines standards for logging, error handling, configuration management, security, and other cross-cutting concerns that are critically important in an enterprise application landscape.

The Foundation follows the principle "Convention over Configuration" while providing the necessary flexibility for specific requirements. Through consistent use of modern Go features like generics and context-based programming, a type-safe, performant, and maintainable foundation emerges. The Foundation is not conceived as a monolithic library but as a collection of focused modules that can be integrated as needed.

## 2. Architectural Principles

### 2.1 Modularity and Independence

The tbp-foundation follows a strictly modular structure where each module has a clearly defined responsibility and minimal dependencies on other modules. This architectural decision is based on the recognition that different services have different requirements. A lean microservice may only need logging and error handling, while a complex business service utilizes the full range of Foundation features.

The modules are designed to be individually integrated and tested. Circular dependencies are avoided through clean interface definitions and the Dependency Inversion Principle. Each module exports clear interfaces and hides its implementation details, facilitating future refactoring.

### 2.2 Modular Adoption Strategy

The Foundation supports flexible adoption through a layered module system:

```go
// Core Foundation - minimal requirements
import "github.com/msto63/tbp/foundation/core"     // Context + Essential Errors
import "github.com/msto63/tbp/foundation/config"   // Basic Configuration

// Extended Features - as needed
import "github.com/msto63/tbp/foundation/logging"  // Structured Logging
import "github.com/msto63/tbp/foundation/security" // Auth + Crypto + Audit
import "github.com/msto63/tbp/foundation/patterns" // Repository + Command + Events
import "github.com/msto63/tbp/foundation/utils"    // stringx, mathx, collections

// Specialized Modules - for specific use cases
import "github.com/msto63/tbp/foundation/testing"  // Test Utilities
import "github.com/msto63/tbp/foundation/metrics"  // Performance Monitoring
import "github.com/msto63/tbp/foundation/workflow" // Event-driven Architecture
```

This modular approach enables:

- **Gradual Migration**: Existing services can adopt Foundation features incrementally
- **Lightweight Services**: Simple services use only core modules
- **Enterprise Services**: Full-featured services leverage all capabilities
- **Dependency Management**: Each module manages its own external dependencies independently

### 2.3 Low-Dependency Philosophy

A core principle of the Foundation is the mindful management of external dependencies. While modern Go development has a temptation to include an external library for every problem, tbp follows a more deliberate approach. External dependencies are only accepted when they provide significant value and are themselves stable and well-maintained.

**Dependency Evaluation Criteria:**

- **Critical Functionality**: Does the dependency provide functionality that would be complex and error-prone to implement in-house?
- **Stability and Maintenance**: Is the library actively maintained with a stable API and good track record?
- **Security**: Does the library have a clean security history and responsive maintainers?
- **Performance**: Does the library meet our performance requirements without introducing overhead?
- **Ecosystem Fit**: Is the library widely accepted in the Go community?

**Acceptable Dependencies:**

- **Standard Extensions**: `golang.org/x/...` packages for extended standard library functionality
- **Protocol Implementations**: Well-established protocols like gRPC, HTTP/2
- **Security Libraries**: Crypto libraries from trusted sources (Go team, major vendors)
- **Domain-Specific Standards**: Libraries for well-defined formats (JSON, Protocol Buffers, decimal arithmetic)
- **Testing and Development**: Quality tools like `testify`, benchmarking utilities, code generation tools

**Dependencies to Avoid:**

- **Trivial Utilities**: Libraries that duplicate simple standard library functionality
- **Experimental Packages**: Unstable or alpha-stage libraries
- **Heavy Frameworks**: Libraries with extensive dependency trees
- **Convenience Wrappers**: Libraries that merely wrap standard functionality without significant value

This philosophy provides several advantages. The attack surface for supply chain attacks is minimized. Updates are simpler since fewer dependencies need coordination. Binary sizes remain manageable. And most importantly, the development team retains full control over the code and can make adaptations without waiting for external maintainers.

### 2.4 Performance by Design

Performance is not an afterthought but an integral part of the design. Every component is developed with efficiency in mind. This doesn't mean performing premature optimizations but choosing efficient data structures and algorithms from the beginning.

Specifically, this manifests in the use of object pools for frequently allocated structures, preference for zero-allocation APIs where possible, and careful interface design to avoid unnecessary indirections. Benchmarks are an integral part of the test suite and ensure that performance regressions are detected early.

## 3. Core Modules

### 3.1 Context Management

The Context module extends Go's standard Context functionality with features essential in an enterprise environment. While Go's context.Context already supports cancellation and deadlines, tbp adds structured metadata, request tracing, and security context.

The extended context system enables information like User-ID, Tenant-ID, Request-ID, and Correlation-ID to be propagated through the entire call chain without explicitly passing them as parameters. This significantly simplifies API signatures while ensuring important contextual information is never lost.

```go
// Example usage of extended Context
ctx := tbpcontext.WithUser(ctx, userID)
ctx = tbpcontext.WithTenant(ctx, tenantID)
ctx = tbpcontext.WithRequestID(ctx, requestID)

// Anywhere in the code
if user, ok := tbpcontext.GetUser(ctx); ok {
    // User-specific logic
}
```

The implementation uses Go's context.Value sparingly and avoids the pitfalls of excessive context usage through clear conventions and type-safe wrapper functions.

### 3.2 Structured Logging

The Foundation's logging system goes far beyond simple text output. It implements structured logging with JSON output, context-dependent fields, and performance-optimized outputs. The API is deliberately kept simple to encourage adoption but provides powerful features for complex scenarios.

A central feature is automatic enrichment of log entries with contextual information. When a logger is used in a context with User-ID and Request-ID, this information is automatically added to every log entry. This enables tracking requests across service boundaries in a microservice environment.

The logging system supports various output formats and destinations. In development, a human-readable format can be used, while in production, structured JSON logs are sent to a central log aggregator. Performance is optimized through zero-allocation techniques and an efficient buffering system.

### 3.3 Error Handling and Error Wrapping

Go's built-in error handling is minimalistic, which suffices for many use cases but falls short in enterprise applications. The tbp-foundation extends error handling with structured errors including classification, stack traces, and machine-readable error codes.

The error system distinguishes between different error classes: temporary errors that are retry-capable; permanent errors that require input changes; and system-critical errors that require immediate attention. This classification enables higher layers to respond intelligently to errors.

```go
// Example of structured error handling
err := tbperrors.New(
    tbperrors.CodeValidationFailed,
    "invalid customer data",
    tbperrors.WithDetail("field", "email"),
    tbperrors.WithDetail("value", userInput),
    tbperrors.WithRetryable(false),
)
```

Stack traces are automatically captured during error creation but can be disabled for performance reasons. The system supports error wrapping according to the Go 1.13+ standard, extended with additional metadata.

### 3.4 Configuration Management

Configuration management in tbp is based on a multi-layered approach that combines flexibility with security. Configurations can come from various sources: environment variables, configuration files, command-line flags, or remote configuration services. The system implements a clear hierarchy where more specific sources override more general ones.

Special attention is paid to type safety of configurations. Instead of working with unstructured maps, each service defines its configuration as a Go struct. The Foundation provides mechanisms for automatic unmarshalling and validation. Sensitive configuration values like passwords or API keys are specially handled and never logged in clear text.

The configuration system supports hot-reloading for certain configuration values. This enables behavior changes without service restart. The implementation ensures that configuration changes are applied atomically and no inconsistent intermediate states occur.

## 4. Utility Modules

### 4.1 Enhanced String Operations (stringx)

The stringx module provides extended string operations frequently needed in business applications but missing from the standard library. A central feature is intelligent abbreviation recognition for TCOL commands. The algorithm analyzes inputs and finds the shortest unique abbreviation from a set of possibilities.

Other important functions include Unicode-aware truncation (which doesn't cut in the middle of a multi-byte character), similarity matching for fuzzy searches, and template processing with security features against injection attacks. All functions are optimized for performance and avoid unnecessary allocations.

### 4.2 Business Mathematics (mathx)

Standard floating-point arithmetic is unsuitable for financial calculations. The mathx module implements decimal arithmetic with configurable precision, various rounding methods (commercial, mathematical, banker's rounding), and currency-specific rules.

The module goes beyond pure arithmetic and offers business-specific functions like interest calculations, currency conversions with historical rates, and tax calculations. All operations are deterministic and reproducible, which is essential for audit requirements.

### 4.3 Type-Safe Collections (collections)

Go's built-in slices and maps are powerful but insufficient for many business use cases. The collections module uses Go's generics to provide type-safe implementations of sets, ordered maps, priority queues, and other data structures.

Special attention is paid to functional operations like Map, Filter, Reduce, implemented in a type-safe and performance-optimized manner. These functions enable complex data manipulations to be formulated in a declarative way, increasing readability and reducing error sources.

## 5. Pattern Implementations

### 5.1 Command Pattern for TCOL

The Command Pattern is central to TCOL implementation. The Foundation provides a generic implementation supporting undo/redo functionality, command queuing, and command logging. Each command is a self-describing unit that can be validated, executed, and undone.

The implementation cleverly uses Go's interfaces to achieve maximum flexibility with minimal complexity. Commands can be executed synchronously or asynchronously, grouped into transactions, and automatically made retry-capable. The system also supports command priorities and deadline-based execution.

### 5.2 Repository Pattern with Generics

The Repository Pattern abstracts data access and enables separation of business logic from persistence details. The tbp-foundation provides a generic repository implementation that can work with various backends: in-memory for tests, SQL databases for relational data, or NoSQL stores for unstructured data.

The implementation uses Go's generics to enable type-safe repositories without code generation. Standard operations like Find, FindAll, Save, and Delete are predefined but can be extended for specific entities. The system also supports complex queries with type-safe query builders.

### 5.3 Event-Driven Architecture Support

Modern microservice architectures are often event-driven. The Foundation provides building blocks for event sourcing, event publishing, and event handling. The system is designed provider-agnostically and can work with various message brokers.

Events are strongly typed and self-describing. The system supports event versioning to enable schema evolution. Event handlers can be made idempotent, and the system provides built-in support for event replay and event projection.

## 6. Security Components

### 6.1 Authentication and Authorization Framework

Security is not an afterthought but an integral part of the Foundation. The Auth framework provides pluggable authentication mechanisms (JWT, OAuth2, mTLS) and a flexible authorization system based on RBAC with extensions for Attribute-Based Access Control.

The system is multi-tenant capable from the ground up. Every operation is executed in the context of a tenant, and data isolation is guaranteed at the framework level. The permission system is fine-grained and allows permissions at the object level.

### 6.2 Cryptography and Key Management

The crypto component encapsulates cryptographic operations and makes them secure and easy to use. The system implements envelope encryption for data at rest, supports key rotation without downtime, and provides secure random number generation.

A special feature is integration with Hardware Security Modules (HSMs) and cloud-based Key Management Services. The API abstracts the complexity of these systems and provides a unified interface for cryptographic operations.

### 6.3 Audit Trail and Compliance

Every business-critical operation must be traceable. The Foundation's audit system automatically records all relevant actions, including who, what, when, and why. Audit logs are tamper-evident through cryptographic chaining.

The system supports various compliance standards out-of-the-box and can be extended for specific requirements. Audit logs can be streamed in real-time to external SIEM systems. Privacy requirements like GDPR-compliant deletion are considered.

## 7. Testing and Quality Assurance

### 7.1 Test Utilities and Mocking

The Foundation provides comprehensive test utilities that simplify and standardize test writing. This includes factories for test data, mocking frameworks for external dependencies, and utilities for integration tests.

A special feature is the time-mocking system that allows time-dependent tests to be made deterministic. The system can "freeze" time, accelerate it, or jump to specific points in time. This is essential for testing time-based business logic.

### 7.2 Benchmarking Framework

Performance tests are first-class citizens in tbp. The Foundation extends Go's benchmarking capabilities with features like regression detection, memory profiling, and latency histograms. Benchmarks can run against defined SLAs and fail builds when performance targets are not met.

The system also collects production metrics and can correlate them with benchmark results. This enables setting realistic performance targets and detecting performance problems early.

## 8. Documentation and Developer Experience

### 8.1 Self-Documenting Code

The Foundation follows strict documentation standards. Every exported function, type, and constant has a meaningful comment. Examples are part of the documentation and are executed as tests to ensure they stay current.

Documentation goes beyond pure API descriptions and explains concepts, best practices, and common pitfalls. Architecture Decision Records (ADRs) document important design decisions and their rationale.

### 8.2 Developer Tools

The Foundation includes CLI tools that simplify development. A code generator can create boilerplate for new services. A migration tool helps with updates between Foundation versions. A diagnostics tool can detect common issues in service implementations.

The tools are themselves written in Go and use the Foundation, making them good examples of usage. They follow Unix philosophy and can be integrated into pipelines and automation.

## 9. Versioning and Compatibility

### 9.1 Semantic Versioning and API Stability

The tbp-foundation follows strict semantic versioning. Minor releases add functionality but don't break existing APIs. Major releases are rare and accompanied by comprehensive migration guides.

API stability is guaranteed through extensive testing. A compatibility test suite ensures that updates don't introduce breaking changes. Deprecated features are clearly marked and maintained for at least two minor releases.

### 9.2 Migration Support

When breaking changes are unavoidable, the Foundation provides tool support for migrations. An AST-based migration tool can automatically perform many code adaptations. For more complex migrations, detailed guides and example migrations are provided.

## 10. Performance and Scalability

### 10.1 Zero-Allocation APIs

Wherever possible, the Foundation provides zero-allocation APIs. This is particularly important for hot-path code like logging or error handling. The APIs use techniques like object pooling, stack allocation, and careful interface design.

Performance is measurable. Every module has benchmarks that track allocations. CI/CD pipelines monitor performance metrics and alert on regressions.

### 10.2 Scalability Patterns

The Foundation implements patterns for horizontal scalability. This includes distributed tracing for request flow analysis, circuit breakers for resilience, and bulkhead patterns for isolation. These patterns are implemented as reusable components.

## 11. Future-Proofing

### 11.1 Extensibility

The Foundation is designed for extension. New modules can be added without affecting existing ones. The plugin system allows functionality to be extended at runtime. Clear interface definitions make it easy to provide alternative implementations.

### 11.2 Cloud-Native and Container-Ready

Although the Foundation doesn't assume specific cloud providers, it is designed cloud-native. Health checks, graceful shutdown, and configuration via environment variables make integration into Kubernetes and similar platforms trivial.

## 12. Summary

The tbp-foundation is more than a collection of utilities - it is the DNA of the Trusted Business Platform. By providing standardized, high-quality solutions to recurring problems, it enables developers to focus on business logic instead of reimplementing infrastructure.

The combination of thoughtful architecture, strict quality standards, and pragmatic solutions makes the Foundation a solid foundation for tbp's ambitious goals. It proves that enterprise-grade doesn't have to mean complexity, but can be achieved through clear design and focused implementation.
