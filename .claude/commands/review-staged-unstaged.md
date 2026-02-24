---
description: Review all staged and unstaged Go code changes comprehensively
allowed-tools:
  - Read
  - Grep
  - Glob
  - Bash
  - Task
  - TaskCreate
  - TaskUpdate
  - TaskList
  - mcp__sequential-thinking__sequentialthinking
model: opus
---

# Review Staged and Unstaged Changes for Go Code

Review all files in the staging area (both staged and unstaged changes).
Ensure comprehensive review of both new files and modified files according to our Go project standards.

Previous review report: $ARGUMENTS
(May or may not be added, ignore the previous review if not specified)

**ULTRA THINK**

## Required Skills

Before beginning review, load these skills for comprehensive Go code review:

1. **code-review** - Go idioms, error handling, testing patterns, project standards
2. **review-verification-protocol** - MANDATORY before reporting ANY findings

These skills contain detailed checklists and patterns. Reference them instead of duplicating content here.

## Execution Process

1. Use the **product-manager-orchestrator** to coordinate specialized agents for comprehensive code review
2. Run all Go validation checks to ensure code quality
3. Verify compliance with project guidelines in `docs/CODING_GUIDELINES.md` and `docs/examples/` directory
4. Apply **review-verification-protocol** before reporting findings to reduce false positives

## Pre-Review Validation Checks

**Run these commands and ensure ZERO errors before proceeding:**

```bash
# Format verification
task fmt

# Comprehensive linting
task lint

# Run tests with race detector
task test-race

# Run tests with coverage
task coverage

# Run benchmarks
task bench

# Build validation
task build

# Module tidiness
task tidy
git diff --exit-code go.mod go.sum

# Run all checks (fmt, vet, lint, test)
task check

# Run pre-commit checks
task check

# Security scan (if gosec is installed)
gosec -quiet ./...

# Shadow variable check (use go vet directly for specific checks)
go vet -shadow ./...

# Verify all exported items have comments (use golint directly if needed)
golint ./... | grep -c "exported" | test $(cat) -eq 0
```

**Dead Code Detection Commands:**

```bash
# Find unused functions/variables (install: go install honnef.co/go/tools/cmd/staticcheck@latest)
staticcheck -checks="U1000" ./...

# Find unreachable code
staticcheck -checks="SA9003" ./...

# Find commented code (excluding TODO/FIXME/NOTE)
rg "^[[:space:]]*//.*" --type go | grep -v "TODO\|FIXME\|NOTE\|Copyright\|Package\|^//$"

# Find empty files
find . -type f -name "*.go" -size 0

# Check for unused dependencies
go mod why -m all | grep "# "

# Find ignored errors
rg "_ =" --type go -g '!*_test.go'
```

## Core Development Philosophy Compliance

### Go Philosophy (`docs/CODING_GUIDELINES.md`)

- **Simplicity First**: Favor simple, obvious solutions over clever ones
- **Explicit Over Implicit**: Make intentions clear in code
- **Composition Over Inheritance**: Use interfaces and embedding
- **Error Handling**: Handle errors explicitly, wrap with context
- **Early Returns**: Reduce nesting with early returns

### Principles

- **Interface Segregation**: Small, focused interfaces
- **Dependency Injection**: Constructor injection for testability
- **Package Design**: internal/ for private, pkg/ for public
- **Table-Driven Tests**: Comprehensive test coverage
- **Idiomatic Go**: Follow Go proverbs and effective Go

## Review Focus Areas

### 1. **🚨 CRITICAL: Interface-First Design**

- [ ] Interfaces defined before implementations
- [ ] Small, focused interfaces (≤5 methods)
- [ ] Interfaces defined in consumer package
- [ ] Accept interfaces, return concrete types
- [ ] Compile-time interface checks (`var _ Interface = (*Type)(nil)`)
- [ ] No large "god" interfaces

### 2. **Error Handling Patterns**

- [ ] **NO ignored errors** (`_ = someFunc()`)
- [ ] Errors wrapped with context using `fmt.Errorf("%w")`
- [ ] quanta errors for known conditions
- [ ] Custom error types implement `error` interface
- [ ] Error messages lowercase, no punctuation
- [ ] No `panic` for normal error handling
- [ ] Proper error checking before operations

### 3. **Testing Standards**

- [ ] Table-driven tests with descriptive test cases
- [ ] Subtests using `t.Run()` for test isolation
- [ ] Test helpers marked with `t.Helper()`
- [ ] Benchmarks for performance-critical code
- [ ] Race condition testing with `-race` flag
- [ ] Mock interfaces for external dependencies
- [ ] Tests co-located with source (\*\_test.go)
- [ ] Integration tests in separate files (\*\_integration_test.go)
- [ ] Minimum 80% code coverage

### 4. **Go Code Quality**

- [ ] **NO magic numbers** - use named constants
- [ ] Consistent receiver types (all pointer or all value)
- [ ] Context as first parameter in functions
- [ ] Proper defer usage immediately after resource acquisition
- [ ] No naked returns in functions >10 lines
- [ ] Preallocation of slices when size is known
- [ ] strings.Builder for string concatenation
- [ ] sync.Pool for frequently allocated objects

### 5. **Concurrency Patterns**

- [ ] No goroutine leaks (proper cleanup)
- [ ] Channels closed only by sender
- [ ] Context for cancellation and timeouts
- [ ] Worker pools for parallel processing
- [ ] Proper synchronization with sync primitives
- [ ] No shared memory without synchronization
- [ ] Rate limiting where appropriate
- [ ] Graceful shutdown handling

### 6. **Package Structure**

- [ ] Follows `cmd/`, `internal/`, `pkg/` structure
- [ ] internal/ packages truly private
- [ ] pkg/ packages have stable APIs
- [ ] No circular dependencies
- [ ] Package names are short and lowercase
- [ ] No util/common/misc packages
- [ ] Clear package boundaries

### 7. **Documentation Standards**

- [ ] Package comment explains purpose
- [ ] All exported items have godoc comments
- [ ] Comments start with item name
- [ ] Examples provided for complex APIs
- [ ] No obvious/redundant comments
- [ ] Business logic documented
- [ ] README.md updated for new features

### 8. **CLI Development (Cobra/Viper)**

- [ ] Commands follow verb-noun pattern
- [ ] Comprehensive help text
- [ ] Flag validation and defaults
- [ ] Configuration hierarchy (flags > env > config file)
- [ ] Proper subcommand structure
- [ ] Shell completion support
- [ ] Error messages guide users

### 9. **Performance & Optimization**

- [ ] Benchmarks for critical paths
- [ ] Profile-guided optimization (pprof)
- [ ] Minimize allocations in hot paths
- [ ] Buffer pools for IO operations
- [ ] Efficient serialization/deserialization
- [ ] Proper database connection pooling
- [ ] Caching where appropriate

### 10. **Security Compliance**

- [ ] Input validation on all user inputs
- [ ] Path traversal prevention
- [ ] SQL injection prevention (prepared statements)
- [ ] No hardcoded secrets or credentials
- [ ] Proper authentication/authorization
- [ ] Secure random number generation
- [ ] Context values use typed keys
- [ ] Sensitive data not logged

### 11. **Build & Dependencies**

- [ ] go.mod properly maintained
- [ ] Minimal external dependencies
- [ ] Vendoring decision documented
- [ ] Build tags used appropriately
- [ ] CGO usage justified if present
- [ ] Module proxy configured
- [ ] Reproducible builds

### 12. **Logging & Observability**

- [ ] Structured logging (e.g., zap, zerolog)
- [ ] Appropriate log levels
- [ ] No sensitive data in logs
- [ ] Request IDs for tracing
- [ ] Metrics for monitoring
- [ ] Health check endpoints
- [ ] Graceful degradation

### 13. **Database & Persistence**

- [ ] Prepared statements for queries
- [ ] Transaction management
- [ ] Connection pool configuration
- [ ] Migration strategy defined
- [ ] Proper NULL handling
- [ ] Index usage optimized
- [ ] Query timeouts configured

### 14. **Project Guidelines Compliance**

- [ ] Follows patterns from `docs/examples/patterns/`
  - [ ] `concurrency.md` for goroutines
  - [ ] `cli.md` for Cobra commands
  - [ ] `testing.md` for test structure
  - [ ] `mocking.md` for test doubles
- [ ] Follows standards from `docs/examples/standards/`
  - [ ] `documentation.md` for godoc
  - [ ] `interfaces.md` for design
  - [ ] `go-specific.md` for idioms
- [ ] Adheres to `docs/CODING_GUIDELINES.md`

### 15. **Dead Code & Unused Elements**

- [ ] No unused functions or types
- [ ] No unreachable code blocks
- [ ] No commented-out code (except TODO/FIXME)
- [ ] No unused imports
- [ ] No empty interfaces
- [ ] No unused struct fields
- [ ] No orphaned test files
- [ ] No unused configuration
- [ ] Dependencies actually used
- [ ] No placeholder implementations

## Review Output

Create a comprehensive review report with:

```markdown
# Go Code Review #[number] - Standards Compliance

## Executive Summary

[2-3 sentence overview focusing on Go idioms, interface design, error handling, and adherence to project standards]

## Validation Results

### Build & Test Commands

- [ ] ✅ `gofmt -l .` - ZERO formatting issues
- [ ] ✅ `task fmt` - Code formatted (gofmt & goimports)
- [ ] ✅ `task lint` - ZERO linting errors
- [ ] ✅ `task test-race` - ALL tests pass, NO races
- [ ] ✅ `task build` - Successful build
- [ ] ✅ `task tidy` - Module is tidy
- [ ] ✅ `gosec ./...` - ZERO security issues (if installed)

## Issues Found

### 🔴 Critical (MUST Fix Before Commit)

#### Error Handling Violations

- [File:line - Ignored error with _ = assignment]
- [File:line - panic used for error handling]
- [File:line - Error not wrapped with context]

#### Interface Design Issues

- [File:line - Interface too large (>5 methods)]
- [File:line - Interface defined in wrong package]
- [File:line - Missing compile-time interface check]

#### Security & Safety

- [File:line - Hardcoded credentials]
- [File:line - SQL injection vulnerability]
- [File:line - Path traversal risk]
- [File:line - Race condition detected]

### 🟡 Important (Should Fix)

#### Testing Issues

- [File:line - Not using table-driven tests]
- [File:line - Missing test coverage (<80%)]
- [File:line - No benchmarks for critical path]
- [File:line - Test not using t.Run for subtests]

#### Code Quality

- [File:line - Magic numbers without constants]
- [File:line - Inconsistent receiver types]
- [File:line - Context not first parameter]
- [File:line - Missing defer for cleanup]

#### Documentation

- [File:line - Exported item missing godoc]
- [File:line - Package missing documentation]
- [File:line - Complex API without examples]

### 🟢 Minor (Consider)

- [Performance optimization opportunities]
- [Code simplification suggestions]
- [Additional test scenarios]
- [Documentation improvements]

## Good Practices Observed

### ✅ Interface Design

- [Examples of small, focused interfaces]
- [Proper interface segregation]
- [Good use of composition]

### ✅ Error Handling

- [Proper error wrapping examples]
- [Good quanta error usage]
- [Clear error messages]

### ✅ Testing Excellence

- [Comprehensive table-driven tests]
- [Good benchmark coverage]
- [Effective use of mocks]

### ✅ Go Idioms

- [Idiomatic Go patterns used]
- [Effective use of channels]
- [Clean package structure]

## Comprehensive Checklist

### 🎯 Interface Design

- [ ] Interfaces ≤5 methods
- [ ] Defined in consumer package
- [ ] Compile-time checks present
- [ ] Small and composable

### 🔄 Error Handling

- [ ] NO ignored errors
- [ ] Errors wrapped with context
- [ ] quanta errors defined
- [ ] No panic for errors

### 🧪 Testing

- [ ] Table-driven tests
- [ ] ≥80% coverage
- [ ] Race detector passes
- [ ] Benchmarks present

### 🔒 Security

- [ ] Input validation
- [ ] No hardcoded secrets
- [ ] SQL injection prevented
- [ ] Path traversal prevented

### 📊 Code Coverage

- Unit: X% (Required: ≥80%)
- Integration: Y% (Required: ≥70%)
- Missing coverage: [list specific files/functions]

### 🏗️ Architecture

- [ ] Follows cmd/internal/pkg structure
- [ ] No circular dependencies
- [ ] Clear package boundaries
- [ ] Dependency injection used

### 🧹 Dead Code Analysis

- [ ] No unused functions
- [ ] No unreachable code
- [ ] No commented-out code
- [ ] No unused imports
- [ ] No orphaned files

## Pattern Compliance

### Concurrency (`docs/examples/patterns/concurrency.md`)

- [Worker pool implementation]
- [Channel usage patterns]
- [Context cancellation]

### CLI (`docs/examples/patterns/cli.md`)

- [Cobra command structure]
- [Viper configuration]
- [Flag validation]

### Testing (`docs/examples/patterns/testing.md`)

- [Table-driven test structure]
- [Test helper usage]
- [Mock patterns]

## Action Items

### Before Commit (Required)

1. Fix all ignored errors
2. Add missing godoc comments
3. Run `golangci-lint run --fix`
4. Ensure all tests pass with -race

### Next Sprint (Recommended)

1. Add benchmarks for [specific functions]
2. Improve test coverage in [packages]
3. Refactor [large interfaces]

## Compliance Score

- Interface Design: X/10
- Error Handling: X/10
- Testing: X/10
- Documentation: X/10
- Security: X/10
- Performance: X/10
- **Overall: X/10**

---

_Review generated following `docs/CODING_GUIDELINES.md` and `docs/examples/` patterns_
```

Save report to `docs/code-reviews/review-[YYYY-MM-DD]-[#].md` (check existing files first)
