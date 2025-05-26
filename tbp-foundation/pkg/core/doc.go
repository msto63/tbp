// Package core implements the essential foundation functionality for the
// Trusted Business Platform (TBP). It provides the fundamental building blocks
// that all other TBP components depend on, including context management,
// error handling, common types, and version information.
//
// The core package follows enterprise-grade design principles with a focus on
// type safety, performance, and maintainability. It serves as the foundation
// layer for the entire TBP ecosystem.
//
// Key Features:
//
// Context Management: Extended context functionality for propagating user
// information, tenant data, request IDs, and correlation IDs throughout
// the entire call chain. Enables comprehensive request tracing and
// multi-tenant operations.
//
// Error Handling: Structured error system with error codes, context information,
// and Go 1.13+ compatibility. Supports error wrapping, classification,
// and detailed error chains for enterprise debugging.
//
// Common Types: Fundamental types and interfaces used throughout TBP,
// including generic repository patterns, pagination support, and
// business domain types with JSON serialization.
//
// Version Management: Comprehensive version information and semantic
// versioning support with build-time injection, compatibility checks,
// and component versioning for service coordination.
//
// Basic Usage:
//
//	// Context with user and tenant information
//	ctx := core.NewUserContext(context.Background(), "user123", "tenant456")
//	
//	// Structured error handling
//	if err := someOperation(); err != nil {
//		return core.Wrap(err, "operation failed")
//	}
//	
//	// Generic repository usage
//	var repo core.Repository[*MyEntity]
//	entities, err := repo.List(ctx, core.NewListOptions().WithLimit(50))
//	
//	// Version information
//	info := core.GetVersionInfoForComponent("my-service")
//	fmt.Printf("Running %s\n", info.String())
//
// Enterprise Features:
//
// The core package is designed for enterprise environments with features
// like audit trails, multi-tenant isolation, performance monitoring,
// and comprehensive error classification. All components are thread-safe
// and optimized for high-performance scenarios.
//
// Package: core
// Title: TBP Core Foundation
// Description: Essential foundation functionality including context management,
//              error handling, common types, and version information for the
//              Trusted Business Platform ecosystem.
// Author: msto63 with Claude Sonnet 4.0
// Version: v0.1.0
// Created: 2025-05-26
// Modified: 2025-05-26
//
// Change History:
// - 2025-05-26 v0.1.0: Initial core foundation implementation
package core