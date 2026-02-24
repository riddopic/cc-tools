---
description: Evaluate a specific task's implementation quality against Go standards
allowed-tools:
  - Read
  - Grep
  - Glob
  - Write
  - Task
  - Bash
  - TaskCreate
  - TaskUpdate
  - TaskList
  - mcp__sequential-thinking__sequentialthinking
argument-hint: "<prp-name> <task-identifier>"
model: opus
---

# Evaluate PRP Task Implementation

## Arguments: $ARGUMENTS

Parse the input arguments as follows:

| Component           | Description                                      | Example                        |
|---------------------|--------------------------------------------------|--------------------------------|
| **PRP Name**        | First token - the PRP filename without extension | `feedback-driven-exploitation` |
| **Task Identifier** | Remaining tokens - the task that was executed    | `Task 1`, `Task 4`             |

**PRP File Path**: `docs/PRPs/{prp-name}.md`

### Parsing Examples

| Input                                   | PRP Path                                      | Target Task |
|-----------------------------------------|-----------------------------------------------|-------------|
| `feedback-driven-exploitation Task 1`   | `docs/PRPs/feedback-driven-exploitation.md`   | Task 1      |
| `feedback-driven-exploitation Task 4`   | `docs/PRPs/feedback-driven-exploitation.md`   | Task 4      |
| `multi-model-ensemble Task 2`           | `docs/PRPs/multi-model-ensemble.md`           | Task 2      |

### Task-Scoped Evaluation

When evaluating, focus ONLY on:

1. Code changes made for the specified task
2. Tests written for the task's functionality
3. Documentation added for the task

## Required Skills

This command uses the following skills (auto-loaded based on context):

- `go-coding-standards` - For Go idioms and patterns
- `tdd-workflow` - For test-first development
- `testing-patterns` - For table-driven tests and mocking
- `code-review` - For reviewing code quality

Evaluate the code and documentation produced by executing a **specific task** from a PRP against quanta's Go standards and idiomatic patterns. This command should be run AFTER `/execute-prp-task` has completed a task.

**Output Files**:

- Evaluation Report: `docs/PRPs/{prp-name}-task-{N}-evaluation.md`
- Fix PRP (if score < 8): `docs/PRPs/{prp-name}-task-{N}-fixes.md`

Example: If evaluating `feedback-driven-exploitation Task 1`:

- Creates: `docs/PRPs/feedback-driven-exploitation-task-1-evaluation.md`
- May create: `docs/PRPs/feedback-driven-exploitation-task-1-fixes.md`

## Agent Orchestration Strategy

Use the **product-manager-orchestrator** to coordinate specialized agents based on the types of fixes needed:

### Core Development Agents

1. **product-manager-orchestrator** - Strategic product leadership and multi-agent coordination

   - Breaks down requirements into specialist tasks
   - Manages feature development workflows
   - Handles crisis response coordination
   - Balances technical debt with new features

2. **backend-systems-engineer** - Core Go package development

   - Interface-first design implementation
   - Error handling with wrapping and context
   - Internal package structure
   - Dependency injection patterns

3. **cli-tool-developer** - Cobra/Viper CLI development

   - Command structure and subcommands
   - Flag validation and configuration
   - Help text and usage documentation
   - Shell completion scripts

4. **concurrency-specialist** - Goroutines and concurrent patterns

   - Worker pool implementations
   - Channel-based communication
   - Context cancellation patterns
   - Race condition prevention

5. **api-backend-engineer** - HTTP/gRPC service development
   - RESTful API design
   - Middleware implementation
   - Context propagation
   - Request/response handling

### Quality & Testing Agents

1. **code-review-specialist** - Go code quality assurance

   - Idiomatic Go verification
   - Interface design review
   - Error handling patterns
   - Performance and maintainability

2. **qa-test-engineer** - Table-driven testing and benchmarks

   - Test table creation
   - Subtests with t.Run
   - Benchmark implementation
   - Mock interface creation

3. **security-threat-analyst** - Security vulnerability assessment
   - Input validation review
   - Path traversal prevention
   - SQL injection prevention
   - Sensitive data handling

### Architecture & Performance Agents

1. **systems-architect** - Package structure and interfaces

   - Interface segregation design
   - Package boundaries
   - Internal vs pkg decisions
   - Dependency management

2. **performance-optimizer** - Go performance improvement

   - CPU and memory profiling
   - Allocation reduction
   - sync.Pool usage
   - Benchmark optimization

3. **database-schema-engineer** - Database design (if applicable)
   - Schema design
   - Query optimization
   - Migration planning
   - Connection pooling

### Analysis & Research Agents

1. **code-analyzer-debugger** - Systematic debugging

   - Race condition detection
   - Goroutine leak identification
   - Memory leak detection
   - Performance bottlenecks

2. **deep-research-specialist** - Go ecosystem research
   - Package evaluation
   - Best practice investigation
   - Library comparisons
   - Pattern research

### Documentation Agents

1. **technical-docs-writer** - Godoc documentation

   - Package documentation
   - Function/method docs
   - Example creation
   - README updates

2. **api-docs-writer** - API documentation
   - OpenAPI/Swagger specs
   - gRPC proto documentation
   - REST endpoint docs
   - Error code documentation

**IMPORTANT** The **product-manager-orchestrator** DOES NOT MAKE ANY CODE CHANGES, she only coordinates with the specialized agents to make the changes.

## Evaluation Process

1. **Identify Implementation Artifacts**

   Based on the PRP file, locate all code/docs produced:

   - Source files created/modified (cmd/_.go, internal/\*\*/_.go, pkg/\*_/_.go)
   - Test files (\*\_test.go)
   - Benchmark files (\*\_bench_test.go)
   - Documentation updates (\*.md, godoc comments)
   - Configuration changes (go.mod, go.sum, .golangci.yml)

2. **Evaluate Code Against Go Standards**

   For each source file created/modified, check against guidelines:

   - **Core Philosophy** (`docs/CODING_GUIDELINES.md`)

     - ✓ Simplicity and clarity prioritized
     - ✓ Early returns to reduce nesting
     - ✓ Interfaces accepted, concrete types returned
     - ✓ Errors handled explicitly

   - **Go Patterns** (`docs/examples/standards/go-specific.md`)

     - ✓ Zero values are useful
     - ✓ Consistent receiver types
     - ✓ Proper defer usage
     - ✓ Context as first parameter
     - ✓ No magic numbers

   - **Interface Design** (`docs/examples/standards/interfaces.md`)

     - ✓ Small, focused interfaces
     - ✓ Interface segregation
     - ✓ Defined in consumer package
     - ✓ Compile-time checks

   - **Documentation** (`docs/examples/standards/documentation.md`)

     - ✓ Package comments present
     - ✓ All exported items documented
     - ✓ Godoc format followed
     - ✓ Examples provided

   - **Testing Patterns** (`docs/examples/patterns/testing.md`)

     - ✓ Table-driven tests
     - ✓ Subtests with t.Run
     - ✓ Test helpers marked
     - ✓ Benchmarks for critical paths

   - **Concurrency** (`docs/examples/patterns/concurrency.md`)
     - ✓ No goroutine leaks
     - ✓ Proper channel closing
     - ✓ Context cancellation
     - ✓ Race-free code

3. **Verify Testing Implementation**

   Check test quality:

   - **Test Coverage**

     - ✓ All exported functions tested
     - ✓ Error paths tested
     - ✓ Edge cases covered
     - ✓ Benchmarks for performance-critical code

   - **Test Quality**
     - ✓ Table-driven structure
     - ✓ Descriptive test names
     - ✓ Tests are independent
     - ✓ Proper cleanup with defer

4. **Run Quality Checks**

   Execute validation commands:

   ```bash
   # Format verification
   gofmt -l .
   test -z "$(gofmt -l .)"

   # Import organization
   goimports -l .
   test -z "$(goimports -l .)"

   # Comprehensive linting
   golangci-lint run --timeout=5m

   # Run tests with race detector
   task test-race

   # Run benchmarks
   task bench

   # Test coverage
   task coverage
   # View coverage summary in terminal
   go tool cover -func=coverage/coverage.out | grep total

   # Build validation
   task build

   # Module tidiness
   task tidy
   git diff --exit-code go.mod go.sum

   # Security scan (if gosec is installed)
   gosec -quiet ./...

   # Shadow variable check
   go vet -shadow ./...

   # Check for ignored errors
   ! rg "_ =" --type go -g '!*_test.go' .

   # Verify godoc comments
   golint ./... | grep -c "exported" | test $(cat) -eq 0
   ```

5. **Performance Validation**

   - ✓ Benchmark results acceptable
   - ✓ Memory allocations minimized
   - ✓ No goroutine leaks
   - ✓ CPU profile clean

6. **Architecture Compliance**

   - ✓ internal/ for private packages
   - ✓ cmd/ for commands
   - ✓ Proper package boundaries
   - ✓ No circular dependencies

## Scoring Rubric (1-10 scale)

### Code Quality (3 points)

- [ ] Follows Go idioms and patterns (1.5 points)
- [ ] Clean golangci-lint output (1.5 points)

### Test Coverage (2 points)

- [ ] Table-driven tests (1 point)
- [ ] Coverage ≥80% (1 point)

### Documentation (2 points)

- [ ] Complete godoc coverage (1 point)
- [ ] Clear package documentation (0.5 points)
- [ ] Examples provided (0.5 points)

### Performance (1 point)

- [ ] Benchmarks pass (0.5 points)
- [ ] No race conditions (0.5 points)

### Error Handling (1 point)

- [ ] Proper error wrapping (0.5 points)
- [ ] No ignored errors (0.5 points)

### Architecture (1 point)

- [ ] Clean package boundaries (0.5 points)
- [ ] Interface-first design (0.5 points)

## Output Format

```markdown
# Implementation Evaluation: {input_prp_name}

## Score: X/10

### Files Evaluated

- `cmd/status.go` (created)
- `cmd/status_test.go` (created)
- `internal/statusline/renderer.go` (created)
- `internal/statusline/renderer_test.go` (created)
- `internal/config/loader.go` (modified)

### Compilation & Tests

- task build: Success
- task test: 25/25 tests pass
- task lint: 3 issues found
- task test-race: No races found
```

### Code Quality Assessment

#### Strengths

- ✅ Interface-first design followed
- ✅ Table-driven tests comprehensive
- ✅ Error wrapping with context
- ✅ Proper context propagation

#### Violations Found

**Critical** (must fix):

- ❌ Error ignored in status.go:47

```go
// Found:
_ = file.Close()

// Should be:
if err := file.Close(); err != nil {
    return fmt.Errorf("failed to close file: %w", err)
}
```

**Major** (should fix):

- ❌ Missing godoc on exported function `FormatStatus`

```go
// Found:
func FormatStatus(s *Status) string {
    return fmt.Sprintf("%s: %s", s.Name, s.Value)
}

// Should be:
// FormatStatus formats a status entry for display.
// It returns a string in the format "name: value".
func FormatStatus(s *Status) string {
    return fmt.Sprintf("%s: %s", s.Name, s.Value)
}
```

- ❌ Interface too large (7 methods)

```go
// Found:
type Manager interface {
    Start() error
    Stop() error
    Restart() error
    Status() string
    Configure(*Config) error
    Validate() error
    Reset() error
}

// Should be split into smaller interfaces:
type Runner interface {
    Start() error
    Stop() error
}

type Configurable interface {
    Configure(*Config) error
    Validate() error
}
```

- ❌ Magic number in code

```go
// Found:
time.Sleep(100 * time.Millisecond)

// Should be:
const retryDelay = 100 * time.Millisecond
time.Sleep(retryDelay)
```

**Minor** (consider fixing):

- ⚠️ Could use sync.Pool for buffer reuse
- ⚠️ Benchmark could be more comprehensive

### Performance Results

- ✅ Benchmark: 50000 ns/op, 1024 B/op, 10 allocs/op
- ✅ No goroutine leaks detected
- ⚠️ Consider preallocating slice in hot path

### Missing Implementation

- ❌ Graceful shutdown handling
- ❌ Context timeout handling
- ❌ Integration tests

## Recommended Actions

1. **Immediate** (blocks deployment):

   - Fix ignored errors
   - Add missing godoc comments

2. **Before PR** (quality gates):

   - Split large interface
   - Replace magic numbers with constants
   - Fix golangci-lint warnings

3. **Future** (technical debt):
   - Add integration tests
   - Implement graceful shutdown
   - Add performance profiling

## Commands to Fix

```bash
# Fix formatting
gofmt -w .
goimports -w .

# Run linter with fixes
golangci-lint run --fix

# Verify tests still pass
task test-race

# Check coverage
task coverage
```

## Follow-up PRP Generation

If score < 8/10, generate improvement PRP:

````markdown
# PRP: {input_name_without_extension}-fixes

## Context

- Original implementation score: X/10
- Critical issues: {list}
- PRP to fix: {original_prp_path}

## Tasks

### Task 1: Fix Critical Issues

**File**: `cmd/status.go`
**Changes**:

- Handle file.Close() error at line 47
- Add proper error wrapping

**Validation**:

```bash
golangci-lint run
! grep -r "_ =" --include="*.go" .
```
````

### Task 2: Add Missing Documentation

**Files**:

- `internal/statusline/formatter.go`: Add godoc to exported functions
- `README.md`: Update with new command documentation

**Validation**:

```bash
golint ./...
go doc -all ./internal/statusline
```

### Task 3: Improve Test Coverage

**File**: `internal/statusline/renderer_test.go`
**Tests**:

- Add benchmark tests
- Test error conditions
- Test concurrent access

**Validation**:

```bash
task test-race
task coverage
# Verify >80% coverage in the output
```

### Task 4: Refactor Interfaces

**File**: `internal/statusline/interfaces.go`
**Changes**:

- Split Manager interface into smaller interfaces
- Follow Interface Segregation Principle

**Validation**:

```bash
task build
# Verify no compilation errors
```

Save as: `docs/PRPs/{input_name_without_extension}-fixes.md`

## Evaluation Report Output

The evaluation report should be saved to: `docs/PRPs/{input_name_without_extension}-evaluation.md`

Where `{input_name_without_extension}` is derived from the input PRP filename.

Examples:

- Input: `status-command.md` → Output: `docs/PRPs/status-command-evaluation.md`
- Input: `metrics-collection.md` → Output: `docs/PRPs/metrics-collection-evaluation.md`
- Input: `config-management.md` → Output: `docs/PRPs/config-management-evaluation.md`

## Usage

```bash
# After running execute-prp-task for a specific task
/execute-prp-task feedback-driven-exploitation Task 1

# Evaluate the specific task's implementation
/evaluate-prp-task feedback-driven-exploitation Task 1

# If issues found, fix and re-evaluate
/fix-prp-task feedback-driven-exploitation Task 1
/evaluate-prp-task feedback-driven-exploitation Task 1

# When task passes, proceed to next task
/execute-prp-task feedback-driven-exploitation Task 2
```

Remember: This evaluation focuses on the ACTUAL CODE produced for the **specific task**, not the entire PRP. The goal is to ensure each task's implementation meets quanta's Go standards before proceeding to the next task.

## Follow-up

After evaluation:

1. If fixes needed, run `/fix-prp-task {prp-name} Task N`
2. Re-run `/evaluate-prp-task {prp-name} Task N` to verify fixes
3. When task passes, proceed to next task with `/execute-prp-task {prp-name} Task N+1`

**Workflow**: Execute → Evaluate → Fix (if needed) → Next Task

---

## AI Evaluation Prompt

**IMPORTANT**: After completing the evaluation report above, you should actively search for and identify any issues in the implemented code. Use the following process:

1. **Systematically review each file** mentioned in the PRP for common Go issues:

   - Error handling (ignored errors, poor wrapping)
   - Interface design (too large, wrong package)
   - Concurrency issues (goroutine leaks, race conditions)
   - Performance problems (excessive allocations, missing pooling)
   - Documentation gaps (missing godoc, unclear comments)
   - Security concerns (input validation, path traversal)

2. **Run mental validation** against each guideline file:

   - Does the code follow patterns from `docs/CODING_GUIDELINES.md`?
   - Are Go idioms from `docs/examples/standards/go-specific.md` applied?
   - Does it follow interface principles from `docs/examples/standards/interfaces.md`?
   - Is the documentation complete per `docs/examples/standards/documentation.md`?

3. **Check for project-specific violations**:

   - Any ignored errors (`_ = someFunc()`)?
   - Any panic for normal error handling?
   - Missing context as first parameter?
   - Magic numbers without constants?
   - Missing table-driven tests?

4. **Generate a comprehensive issues list** including:

   - File and line numbers where issues occur
   - Severity (Critical/Major/Minor)
   - Suggested fixes with code examples
   - Impact on functionality or maintainability

5. **Final Step**: Read and execute `.claude/commands/review-staged-unstaged.md`

Your evaluation should be thorough and constructive, helping developers understand not just what's wrong, but why it matters in Go and how to fix it idiomatically.
